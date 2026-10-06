package auth

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID    string
	Name  string
	Email string
	Role  string
}

var ErrInvalid = errors.New("email atau password salah")

func NewSessionManager(key string) *scs.SessionManager {
	sess := scs.New()
	sess.Cookie.Name = "inventaris_session"
	sess.Cookie.HttpOnly = true
	sess.Cookie.SameSite = http.SameSiteLaxMode
	// Secure otomatis: ikuti APP_URL / APP_ENV (F1)
	if strings.HasPrefix(os.Getenv("APP_URL"), "https://") || os.Getenv("APP_ENV") == "prod" {
		sess.Cookie.Secure = true
	}
	sess.Lifetime = 24 * 7 * time.Hour
	if key != "dev-only-insecure-key" {
		sess.Cookie.Persist = true
	}
	return sess
}

var _ = time.Now

func HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

func GetUserByEmail(ctx context.Context, pool *pgxpool.Pool, email string) (*User, string, error) {
	row := pool.QueryRow(ctx,
		`SELECT id, name, email, role, password_hash FROM users WHERE email = $1`, strings.ToLower(email))
	var u User
	var hash string
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &hash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", ErrInvalid
		}
		return nil, "", err
	}
	return &u, hash, nil
}

func CreateUserInteractive(ctx context.Context, pool *pgxpool.Pool, email string) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Nama lengkap: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Admin"
	}
	fmt.Printf("Password: ")
	pw, err := readPassword(reader)
	if err != nil {
		return err
	}
	if len(pw) < 8 {
		return errors.New("password minimal 8 karakter")
	}
	role := "admin"
	if os.Getenv("USER_ROLE") == "petugas" {
		role = "petugas"
	}
	hash, err := HashPassword(pw)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx,
		`INSERT INTO users (name, email, password_hash, role) VALUES ($1,$2,$3,$4)
		 ON CONFLICT (email) DO UPDATE SET password_hash=$3, role=$4, name=$1`,
		name, strings.ToLower(email), hash, role)
	return err
}

// readPassword reads a line without echo on a TTY; falls back to the buffered
// reader when stdin is piped (scripts/CI).
func readPassword(reader *bufio.Reader) (string, error) {
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		b, err := term.ReadPassword(fd)
		fmt.Println()
		return string(b), err
	}
	line, err := reader.ReadString('\n')
	return strings.TrimSpace(line), err
}

var _ = time.Now
