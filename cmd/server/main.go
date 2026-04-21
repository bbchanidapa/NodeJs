package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// ===== Basic handlers =====

func hello(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprint(w, "Hello World!")
}

func newUUID(w http.ResponseWriter, r *http.Request) {
	id := uuid.New()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, id.String())
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status":"ok"}`)
}

// ===== Request / Response models =====

type createUserRequest struct {
	ID        string `json:"id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

type updateUserRequest struct {
	Firstname *string `json:"firstname"`
	Lastname  *string `json:"lastname"`
}

type userRecord struct {
	ID        string `json:"id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

// ===== Database setup =====

func nonEmpty(v string) string {
	return strings.TrimSpace(v)
}

func postgresDSNFromEnv() string {
	if dsn := nonEmpty(os.Getenv("DATABASE_URL")); dsn != "" {
		return dsn
	}

	host := nonEmpty(os.Getenv("PGHOST"))
	if host == "" {
		host = "localhost"
	}
	port := nonEmpty(os.Getenv("PGPORT"))
	if port == "" {
		port = "5432"
	}
	user := nonEmpty(os.Getenv("PGUSER"))
	if user == "" {
		user = "postgres"
	}
	dbName := nonEmpty(os.Getenv("PGDATABASE"))
	if dbName == "" {
		dbName = "postgres"
	}
	sslMode := nonEmpty(os.Getenv("PGSSLMODE"))
	if sslMode == "" {
		sslMode = "disable"
	}
	password := nonEmpty(os.Getenv("PGPASSWORD"))

	parts := []string{
		"host=" + host,
		"port=" + port,
		"user=" + user,
		"dbname=" + dbName,
		"sslmode=" + sslMode,
	}
	if password != "" {
		parts = append(parts, "password="+password)
	}
	return strings.Join(parts, " ")
}

func initPostgres(ctx context.Context) (*sql.DB, error) {
	dsn := postgresDSNFromEnv()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, err
	}

	bootstrapCtx, bootstrapCancel := context.WithTimeout(ctx, 5*time.Second)
	defer bootstrapCancel()
	if err := ensureUsersTable(bootstrapCtx, db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func ensureUsersTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE EXTENSION IF NOT EXISTS pgcrypto;

		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			firstname VARCHAR(100) NOT NULL,
			lastname VARCHAR(100) NOT NULL
		);

		ALTER TABLE users
		ALTER COLUMN id
		SET DEFAULT gen_random_uuid();
	`)
	return err
}

// ===== Utility =====

func listUsersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if db == nil {
			http.Error(w, "postgres is not configured", http.StatusServiceUnavailable)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		rows, err := db.QueryContext(ctx, `SELECT id, firstname, lastname FROM users ORDER BY firstname, lastname`)
		if err != nil {
			http.Error(w, "failed to query users", http.StatusBadGateway)
			return
		}
		defer rows.Close()

		users := make([]userRecord, 0)
		for rows.Next() {
			var u userRecord
			if err := rows.Scan(&u.ID, &u.Firstname, &u.Lastname); err != nil {
				http.Error(w, "failed to parse users", http.StatusBadGateway)
				return
			}
			users = append(users, u)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "failed to read users", http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":    true,
			"count": len(users),
			"users": users,
		})
	}
}

func extractUserID(path string) (string, bool, bool) {
	id := strings.TrimPrefix(path, "/users/")
	id = strings.TrimSpace(id)
	if id == "" || strings.Contains(id, "/") {
		return "", false, false
	}
	if _, err := uuid.Parse(id); err != nil {
		return "", false, true
	}
	return id, true, true
}

// ===== GET /users/:id =====

func getUserByIDHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if db == nil {
			http.Error(w, "postgres is not configured", http.StatusServiceUnavailable)
			return
		}

		id, found, valid := extractUserID(r.URL.Path)
		if !found {
			http.NotFound(w, r)
			return
		}
		if !valid {
			http.Error(w, "id must be a valid uuid", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var u userRecord
		err := db.QueryRowContext(
			ctx,
			`SELECT id, firstname, lastname FROM users WHERE id = $1`,
			id,
		).Scan(&u.ID, &u.Firstname, &u.Lastname)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "user not found", http.StatusNotFound)
				return
			}
			http.Error(w, "failed to query user", http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":   true,
			"user": u,
		})
	}
}

// ===== PUT/PATCH /users/:id =====

func updateUserByIDHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut && r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if db == nil {
			http.Error(w, "postgres is not configured", http.StatusServiceUnavailable)
			return
		}

		id, found, valid := extractUserID(r.URL.Path)
		if !found {
			http.NotFound(w, r)
			return
		}
		if !valid {
			http.Error(w, "id must be a valid uuid", http.StatusBadRequest)
			return
		}

		var req updateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}

		var firstnameParam any
		var lastnameParam any

		if req.Firstname != nil {
			firstname := strings.TrimSpace(*req.Firstname)
			if firstname == "" {
				http.Error(w, "firstname cannot be empty", http.StatusBadRequest)
				return
			}
			firstnameParam = firstname
		}
		if req.Lastname != nil {
			lastname := strings.TrimSpace(*req.Lastname)
			if lastname == "" {
				http.Error(w, "lastname cannot be empty", http.StatusBadRequest)
				return
			}
			lastnameParam = lastname
		}

		if r.Method == http.MethodPut {
			if firstnameParam == nil || lastnameParam == nil {
				http.Error(w, "firstname and lastname are required", http.StatusBadRequest)
				return
			}
		} else {
			// PATCH ต้องมีอย่างน้อย 1 ฟิลด์ที่ต้องการแก้
			if firstnameParam == nil && lastnameParam == nil {
				http.Error(w, "at least one field is required for patch", http.StatusBadRequest)
				return
			}
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var updated userRecord
		err := db.QueryRowContext(
			ctx,
			`UPDATE users
			 SET firstname = COALESCE($2, firstname),
			     lastname = COALESCE($3, lastname)
			 WHERE id = $1
			 RETURNING id, firstname, lastname`,
			id, firstnameParam, lastnameParam,
		).Scan(&updated.ID, &updated.Firstname, &updated.Lastname)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "user not found", http.StatusNotFound)
				return
			}
			http.Error(w, "failed to update user", http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":   true,
			"user": updated,
		})
	}
}

// ===== POST /users =====

func createUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if db == nil {
			http.Error(w, "postgres is not configured", http.StatusServiceUnavailable)
			return
		}

		var req createUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}
		if req.Firstname == "" || req.Lastname == "" {
			http.Error(w, "firstname and lastname are required", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var userID string
		idFromRequest := strings.TrimSpace(req.ID)
		if idFromRequest == "" {
			err := db.QueryRowContext(
				ctx,
				`INSERT INTO users (firstname, lastname) VALUES ($1, $2) RETURNING id`,
				req.Firstname, req.Lastname,
			).Scan(&userID)
			if err != nil {
				http.Error(w, "failed to write user to postgres", http.StatusBadGateway)
				return
			}
		} else {
			if _, err := uuid.Parse(idFromRequest); err != nil {
				http.Error(w, "id must be a valid uuid", http.StatusBadRequest)
				return
			}

			err := db.QueryRowContext(
				ctx,
				`INSERT INTO users (id, firstname, lastname) VALUES ($1, $2, $3) RETURNING id`,
				idFromRequest, req.Firstname, req.Lastname,
			).Scan(&userID)
			if err != nil {
				if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
					http.Error(w, "user id already exists", http.StatusConflict)
					return
				}
				http.Error(w, "failed to write user to postgres", http.StatusBadGateway)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok": true,
			"user": map[string]any{
				"id":        userID,
				"firstname": req.Firstname,
				"lastname":  req.Lastname,
			},
		})
	}
}

// ===== Routers =====

func userByIDHandler(db *sql.DB) http.HandlerFunc {
	getByID := getUserByIDHandler(db)
	updateByID := updateUserByIDHandler(db)

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getByID(w, r)
		case http.MethodPut:
			updateByID(w, r)
		case http.MethodPatch:
			updateByID(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func usersHandler(db *sql.DB) http.HandlerFunc {
	create := createUserHandler(db)
	list := listUsersHandler(db)

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			list(w, r)
		case http.MethodPost:
			create(w, r)
		case http.MethodPut, http.MethodPatch:
			http.Error(w, "please use /users/{id} for update", http.StatusBadRequest)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// ===== main =====

func main() {
	db, err := initPostgres(context.Background())
	if err != nil {
		fmt.Println("warning: PostgreSQL disabled -", err)
	} else {
		defer db.Close()
	}

	http.HandleFunc("/health", health)
	http.HandleFunc("/uuid", newUUID)
	http.HandleFunc("/users", usersHandler(db))
	http.HandleFunc("/users/", userByIDHandler(db))
	http.HandleFunc("/", hello)

	// พอร์ต: อ่านจากตัวแปรสภาพแวดล้อม PORT ถ้าไม่มีใช้ 3000
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("เปิดเซิร์ฟเวอร์ที่พอร์ต", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("error:", err)
	}
}
