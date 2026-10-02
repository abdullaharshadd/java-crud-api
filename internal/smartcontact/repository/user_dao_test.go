package repository

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"migrated-app/internal/model"
)

func newTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite3: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func newTestRepo(t *testing.T) *UserRepository {
	t.Helper()
	db := newTestDB(t)
	repo, err := NewUserRepository(context.Background(), db)
	if err != nil {
		t.Fatalf("NewUserRepository: %v", err)
	}
	return repo
}

func makeUser(name, email, password, role, about string) *model.User {
	u := &model.User{}
	u.SetName(name)
	u.SetEmail(email)
	u.SetPassword(password)
	u.SetRole(role)
	u.SetAbout(about)
	return u
}

// TestSave covers the Save (insert) and Merge (update) paths that together
// implement JpaRepository.save semantics.
func TestSave(t *testing.T) {
	tests := []struct {
		name    string
		run     func(t *testing.T, repo *UserRepository)
	}{
		{
			name: "new user gets generated id",
			run: func(t *testing.T, repo *UserRepository) {
				u := makeUser("alice", "alice@example.com", "secret", "USER", "about alice")
				saved, err := repo.Save(context.Background(), u)
				if err != nil {
					t.Fatalf("Save: %v", err)
				}
				if saved.GetID() == 0 {
					t.Error("expected non-zero generated id")
				}
				if saved.GetName() != "alice" {
					t.Errorf("expected name alice, got %s", saved.GetName())
				}
			},
		},
		{
			name: "existing user updated via Merge returns updated user",
			run: func(t *testing.T, repo *UserRepository) {
				u := makeUser("bob", "bob@example.com", "pass", "ADMIN", "")
				saved, err := repo.Save(context.Background(), u)
				if err != nil {
					t.Fatalf("Save: %v", err)
				}
				saved.SetEmail("bob2@example.com")
				merged, err := repo.Merge(context.Background(), saved)
				if err != nil {
					t.Fatalf("Merge: %v", err)
				}
				if merged.GetEmail() != "bob2@example.com" {
					t.Errorf("expected updated email, got %s", merged.GetEmail())
				}
				// Verify persistence
				found, ok, err := repo.FindByID(context.Background(), saved.GetID())
				if err != nil || !ok {
					t.Fatalf("FindByID after merge: err=%v ok=%v", err, ok)
				}
				if found.GetEmail() != "bob2@example.com" {
					t.Errorf("DB not updated: got %s", found.GetEmail())
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newTestRepo(t)
			tc.run(t, repo)
		})
	}
}

// TestFindByID covers findById Optional semantics.
func TestFindByID(t *testing.T) {
	tests := []struct {
		name      string
		setupName string
		lookupID  func(savedID int) int
		wantFound bool
	}{
		{
			name:      "user exists returns user",
			setupName: "carol",
			lookupID:  func(savedID int) int { return savedID },
			wantFound: true,
		},
		{
			name:      "user does not exist returns not found",
			setupName: "dave",
			lookupID:  func(_ int) int { return 99999 },
			wantFound: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newTestRepo(t)
			u := makeUser(tc.setupName, "e@e.com", "p", "r", "a")
			saved, err := repo.Save(context.Background(), u)
			if err != nil {
				t.Fatalf("Save: %v", err)
			}
			id := tc.lookupID(saved.GetID())
			found, ok, err := repo.FindByID(context.Background(), id)
			if err != nil {
				t.Fatalf("FindByID: %v", err)
			}
			if ok != tc.wantFound {
				t.Errorf("wantFound=%v got=%v", tc.wantFound, ok)
			}
			if tc.wantFound && found == nil {
				t.Error("expected non-nil user when found")
			}
			if !tc.wantFound && found != nil {
				t.Error("expected nil user when not found")
			}
		})
	}
}

// TestFindAll covers findAll list semantics.
func TestFindAll(t *testing.T) {
	tests := []struct {
		name      string
		seedCount int
	}{
		{"no users returns empty non-nil slice", 0},
		{"multiple users returns all", 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newTestRepo(t)
			for i := 0; i < tc.seedCount; i++ {
				u := makeUser("user", "e@e.com", "p", "r", "a")
				if _, err := repo.Save(context.Background(), u); err != nil {
					t.Fatalf("Save: %v", err)
				}
			}
			all, err := repo.FindAll(context.Background())
			if err != nil {
				t.Fatalf("FindAll: %v", err)
			}
			if all == nil {
				t.Fatal("FindAll returned nil slice")
			}
			if len(all) != tc.seedCount {
				t.Errorf("expected %d users, got %d", tc.seedCount, len(all))
			}
		})
	}
}

// TestFindByName covers the derived findByName query.
func TestFindByName(t *testing.T) {
	tests := []struct {
		name       string
		seedNames  []string
		searchName string
		wantErr    bool
		wantNil    bool
	}{
		{
			name:       "user with name exists returns user",
			seedNames:  []string{"eve"},
			searchName: "eve",
			wantNil:    false,
		},
		{
			name:       "no user with name returns ErrUserNotFound",
			seedNames:  []string{},
			searchName: "nobody",
			wantErr:    true,
			wantNil:    true,
		},
		{
			name:       "multiple users same name returns one (LIMIT 1)",
			seedNames:  []string{"frank", "frank"},
			searchName: "frank",
			wantNil:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newTestRepo(t)
			for _, n := range tc.seedNames {
				u := makeUser(n, "e@e.com", "p", "r", "a")
				if _, err := repo.Save(context.Background(), u); err != nil {
					t.Fatalf("Save: %v", err)
				}
			}
			found, err := repo.FindByName(context.Background(), tc.searchName)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tc.wantNil && found != nil {
				t.Error("expected nil user")
			}
			if !tc.wantNil && found == nil {
				t.Error("expected non-nil user")
			}
		})
	}
}

// TestDeleteByID covers deleteById success and not-found error.
func TestDeleteByID(t *testing.T) {
	tests := []struct {
		name    string
		seedIt  bool
		wantErr bool
	}{
		{"existing user is deleted", true, false},
		{"non-existent id returns error", false, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newTestRepo(t)
			id := 99999
			if tc.seedIt {
				u := makeUser("grace", "g@g.com", "p", "r", "a")
				saved, err := repo.Save(context.Background(), u)
				if err != nil {
					t.Fatalf("Save: %v", err)
				}
				id = saved.GetID()
			}
			err := repo.DeleteByID(context.Background(), id)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tc.wantErr {
				_, ok, err2 := repo.FindByID(context.Background(), id)
				if err2 != nil {
					t.Fatalf("FindByID after delete: %v", err2)
				}
				if ok {
					t.Error("user still exists after deletion")
				}
			}
		})
	}
}

// TestCountAndExistsByID covers count and existsById semantics.
func TestCountAndExistsByID(t *testing.T) {
	t.Run("count returns 0 when empty", func(t *testing.T) {
		repo := newTestRepo(t)
		n, err := repo.Count(context.Background())
		if err != nil {
			t.Fatalf("Count: %v", err)
		}
		if n != 0 {
			t.Errorf("expected 0, got %d", n)
		}
	})

	t.Run("count reflects number of saved users", func(t *testing.T) {
		repo := newTestRepo(t)
		for i := 0; i < 2; i++ {
			u := makeUser("h", "h@h.com", "p", "r", "a")
			if _, err := repo.Save(context.Background(), u); err != nil {
				t.Fatalf("Save: %v", err)
			}
		}
		n, err := repo.Count(context.Background())
		if err != nil {
			t.Fatalf("Count: %v", err)
		}
		if n != 2 {
			t.Errorf("expected 2, got %d", n)
		}
	})

	t.Run("existsById true when user present", func(t *testing.T) {
		repo := newTestRepo(t)
		u := makeUser("ivan", "i@i.com", "p", "r", "a")
		saved, err := repo.Save(context.Background(), u)
		if err != nil {
			t.Fatalf("Save: %v", err)
		}
		ok, err := repo.ExistsByID(context.Background(), saved.GetID())
		if err != nil {
			t.Fatalf("ExistsByID: %v", err)
		}
		if !ok {
			t.Error("expected true, got false")
		}
	})

	t.Run("existsById false when user absent", func(t *testing.T) {
		repo := newTestRepo(t)
		ok, err := repo.ExistsByID(context.Background(), 99999)
		if err != nil {
			t.Fatalf("ExistsByID: %v", err)
		}
		if ok {
			t.Error("expected false, got true")
		}
	})
}