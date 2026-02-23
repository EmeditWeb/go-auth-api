package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"crypto/sha256"

	"golang.org/x/crypto/bcrypt"
)
// anonymous variable reps the unauthenticated user
var AnonymousUser = &User{}

// IsAnonymous checks if a user is anonymous
func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}


var ErrRecordNotFound = errors.New("record not found")

// Models acts as a wrapper for our various model types
type Models struct {
	Tokens TokenModel
	Users  UserModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users:  UserModel{DB: db},
		Tokens: TokenModel{DB: db},
	}
}

// --- USER MODEL ---

type UserModel struct {
	DB *sql.DB
}

// InsertUser inserts a new user into the database 
func (m *UserModel) InsertUser(user *User) error {
	query := `
        INSERT INTO users (username, email, password_hash, role)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at`

	args := []interface{}{
		user.Username,
		user.Email,
		user.Password.hash,
		user.Role,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// We must scan into both ID and CreatedAt because of the RETURNING clause
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt)
}

// GetUserByEmail retrieves a user from the database based on their email address
func (m *UserModel) GetUserByEmail(email string) (*User, error) {
	query := `
        SELECT id, created_at, username, email, password_hash, role
        FROM users
        WHERE email = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.Role,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// --- TOKEN MODEL ---

type TokenModel struct {
	DB *sql.DB
}

func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	return token, err
}

func (m TokenModel) Insert(token *Token) error {
	query := `
        INSERT INTO tokens (hash, user_id, expiry, scope)
        VALUES ($1, $2, $3, $4)`

	args := []interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

// DeleteAllForUser deletes all tokens for a specific user and scope
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
    query := `
        DELETE FROM tokens 
        WHERE scope = $1 AND user_id = $2`

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    _, err := m.DB.ExecContext(ctx, query, scope, userID)
    return err
}
// --- STRUCTS & METHODS ---

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Role      string    `json:"role"`
}

type password struct {
	plainText string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plainText = plaintextPassword
	p.hash = hash
	return nil
}

func (p *password) Matches(password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(password))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func (m *UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
    // We calculate the SHA256 hash of the plaintext token provided by the user
    // to match it against the hash stored in our database
    tokenHash := sha256.Sum256([]byte(tokenPlaintext))

    query := `
        SELECT users.id, users.created_at, users.username, users.email, users.password_hash, users.role
        FROM users
        INNER JOIN tokens
        ON users.id = tokens.user_id
        WHERE tokens.hash = $1
        AND tokens.scope = $2
        AND tokens.expiry > $3`

    args := []interface{}{tokenHash[:], tokenScope, time.Now()}

    var user User

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    err := m.DB.QueryRowContext(ctx, query, args...).Scan(
        &user.ID,
        &user.CreatedAt,
        &user.Username,
        &user.Email,
        &user.Password.hash,
        &user.Role,
    )

    if err != nil {
        switch {
        case errors.Is(err, sql.ErrNoRows):
            return nil, ErrRecordNotFound
        default:
            return nil, err
        }
    }

	return &user, nil
}