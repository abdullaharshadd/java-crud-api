module migrated-app

go 1.23

require (
	github.com/go-chi/chi/v5 v5.1.0
	github.com/go-sql-driver/mysql v1.8.1
	github.com/jmoiron/sqlx v1.4.0
	github.com/lib/pq v1.10.9
	github.com/rs/zerolog v1.33.0
	github.com/spf13/viper v1.19.0
)

require filippo.io/edwards25519 v1.1.0 // indirect

// Pin versions compatible with the local go 1.23 toolchain (GOTOOLCHAIN=local);
// newer releases (e.g. go-sql-driver/mysql v1.10.x) require go >= 1.24.
replace (
	github.com/go-chi/chi/v5 => github.com/go-chi/chi/v5 v5.1.0
	github.com/go-sql-driver/mysql => github.com/go-sql-driver/mysql v1.8.1
	github.com/rs/zerolog => github.com/rs/zerolog v1.33.0
	github.com/spf13/viper => github.com/spf13/viper v1.19.0
)
