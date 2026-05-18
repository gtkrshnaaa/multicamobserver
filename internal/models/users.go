package models

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// GetUserByEmail retrieves a user by their email address from PostgreSQL
func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	row := db.QueryRow("SELECT id, email, password_hash, created_at FROM users WHERE email = $1", email)

	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &u, nil
}

// AuthenticateUser checks if an administrator's credentials are valid
func AuthenticateUser(db *sql.DB, email, password string) (*User, error) {
	u, err := GetUserByEmail(db, email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	return u, nil
}

// UpdateUser modifies an administrator's email/username and optionally their password hash
func UpdateUser(db *sql.DB, id int, email, plainPassword string) error {
	if plainPassword != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		_, err = db.Exec("UPDATE users SET email = $1, password_hash = $2 WHERE id = $3", email, string(hashed), id)
		return err
	}

	_, err := db.Exec("UPDATE users SET email = $1 WHERE id = $2", email, id)
	return err
}
