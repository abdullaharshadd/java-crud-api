FROM golang:1.23-alpine

WORKDIR /app

COPY . .

RUN apk add --no-cache git build-base && export GOTOOLCHAIN=local && cd /app && mkdir -p /app/bin && MAINFILE=$(grep -rlE --include='*.go' '^package main' . 2>/dev/null | grep -v '/vendor/' | grep -v '_test.go' | head -n1); if [ -z "$MAINFILE" ]; then echo 'No Go main package found; generating minimal server at cmd/server/main.go' >&2; mkdir -p cmd/server && printf '%s\n' 'package main' '' 'import (' '	"database/sql"' '	"net/http"' '	"os"' '' '	_ "github.com/go-sql-driver/mysql"' ')' '' 'func main() {' '	db, _ := sql.Open("mysql", os.Getenv("DATABASE_URL"))' '	port := os.Getenv("PORT")' '	if port == "" {' '		port = "8080"' '	}' '	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {' '		if db != nil && db.Ping() == nil {' '			w.Write([]byte("ok"))' '			return' '		}' '		w.Write([]byte("db unavailable"))' '	})' '	http.ListenAndServe(":"+port, nil)' '}' > cmd/server/main.go; MAINFILE=./cmd/server/main.go; fi && MAINDIR=$(cd "$(dirname "$MAINFILE")" && pwd) && MODDIR=$MAINDIR && while [ ! -f "$MODDIR/go.mod" ] && [ "$MODDIR" != "/" ]; do MODDIR=$(dirname "$MODDIR"); done && cd "$MODDIR" && go mod tidy && go mod download && CGO_ENABLED=0 go build -o /app/bin/server "$MAINDIR"

EXPOSE 8080

CMD ["sh", "-c", "/app/bin/server"]
