package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"

	"migrated-app/internal/config"
	"migrated-app/internal/httpapi"
	"migrated-app/internal/model"
	"migrated-app/internal/service"
	"migrated-app/internal/store"
)

// userRepoAdapter adapts store.MySQLUserRepository to service.UserRepository.
type userRepoAdapter struct {
	repo *store.MySQLUserRepository
}

func (a *userRepoAdapter) Save(ctx context.Context, u *model.User) (*model.User, error) {
	return a.repo.Save(ctx, u)
}

func (a *userRepoAdapter) FindAll(ctx context.Context) ([]model.User, error) {
	return a.repo.FindAll(ctx)
}

func (a *userRepoAdapter) FindByID(ctx context.Context, id int32) (*model.User, bool, error) {
	u, ok, err := a.repo.FindByID(ctx, id)
	if err != nil || !ok {
		return nil, ok, err
	}
	return &u, true, nil
}

func (a *userRepoAdapter) DeleteByID(ctx context.Context, id int32) error {
	return a.repo.DeleteByID(ctx, id)
}

func (a *userRepoAdapter) FindByName(ctx context.Context, name string) (*model.User, error) {
	u, ok, err := a.repo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return &u, nil
}

// memUserRepo is an in-memory fallback used only when no MySQL driver is
// registered in this binary.
type memUserRepo struct {
	mu     sync.Mutex
	nextID int32
	users  map[int32]model.User
	order  []int32
}

func newMemUserRepo() *memUserRepo {
	return &memUserRepo{nextID: 1, users: map[int32]model.User{}}
}

func (m *memUserRepo) Save(_ context.Context, u *model.User) (*model.User, error) {
	if u == nil {
		return nil, errors.New("memrepo: save user: entity must not be nil")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	saved := *u
	if _, ok := m.users[u.ID]; u.ID == 0 || !ok {
		saved.ID = m.nextID
		m.nextID++
		m.order = append(m.order, saved.ID)
	}
	m.users[saved.ID] = saved
	return &saved, nil
}

func (m *memUserRepo) FindAll(_ context.Context) ([]model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.User, 0, len(m.order))
	for _, id := range m.order {
		if u, ok := m.users[id]; ok {
			out = append(out, u)
		}
	}
	return out, nil
}

func (m *memUserRepo) FindByID(_ context.Context, id int32) (*model.User, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return nil, false, nil
	}
	return &u, true, nil
}

func (m *memUserRepo) DeleteByID(_ context.Context, id int32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[id]; !ok {
		return fmt.Errorf("memrepo: delete user id %d: %w", id, store.ErrNoRowsDeleted)
	}
	delete(m.users, id)
	for i, v := range m.order {
		if v == id {
			m.order = append(m.order[:i], m.order[i+1:]...)
			break
		}
	}
	return nil
}

func (m *memUserRepo) FindByName(_ context.Context, name string) (*model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var found *model.User
	for _, id := range m.order {
		u, ok := m.users[id]
		if !ok || u.Name == nil || *u.Name != name {
			continue
		}
		if found != nil {
			return nil, store.ErrMultipleResults
		}
		c := u
		found = &c
	}
	return found, nil
}

// mysqlDSN converts a URL-style DATABASE_URL (mysql://user:pass@host:port/db)
// into the go-sql-driver DSN format; native DSNs are passed through unchanged.
func mysqlDSN(raw string) string {
	if !strings.HasPrefix(raw, "mysql://") {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	var b strings.Builder
	if u.User != nil {
		b.WriteString(u.User.Username())
		if p, ok := u.User.Password(); ok {
			b.WriteString(":")
			b.WriteString(p)
		}
		b.WriteString("@")
	}
	b.WriteString("tcp(")
	b.WriteString(u.Host)
	b.WriteString(")/")
	b.WriteString(strings.TrimPrefix(u.Path, "/"))
	if u.RawQuery != "" {
		b.WriteString("?")
		b.WriteString(u.RawQuery)
	}
	return b.String()
}

func driverRegistered(name string) bool {
	for _, d := range sql.Drivers() {
		if d == name {
			return true
		}
	}
	return false
}

func buildUserRepo(cfg *config.Config) service.UserRepository {
	if !driverRegistered("mysql") {
		log.Error().Msg("mysql driver not registered; using in-memory user repository")
		return newMemUserRepo()
	}
	if cfg.DatabaseURL == "" {
		log.Warn().Msg("DATABASE_URL is not set")
	}
	db, err := sql.Open("mysql", mysqlDSN(cfg.DatabaseURL))
	if err != nil {
		log.Error().Err(err).Msg("open database; using in-memory user repository")
		return newMemUserRepo()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := model.EnsureUserSchema(ctx, db); err != nil {
		log.Error().Err(err).Msg("ensure user schema")
	}
	return &userRepoAdapter{repo: store.NewMySQLUserRepository(db)}
}

func buildRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	cfg, err := config.Load()
	if err != nil {
		log.Error().Err(err).Msg("load config")
		cfg = &config.Config{}
	}

	svc := service.NewUserService(buildUserRepo(cfg))
	r.Mount("/", httpapi.NewRouter(svc))
	return r
}