package models

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// GetUserByUsername retrieves a user by their unique username from PostgreSQL
func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	row := db.QueryRow("SELECT id, username, email, password_hash, created_at FROM users WHERE username = $1", username)

	var u User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &u, nil
}

// GetUserByEmail retrieves a user by their email address from PostgreSQL
func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	row := db.QueryRow("SELECT id, username, email, password_hash, created_at FROM users WHERE email = $1", email)

	var u User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &u, nil
}

// AuthenticateUser checks if an administrator's credentials are valid using their username
func AuthenticateUser(db *sql.DB, username, password string) (*User, error) {
	u, err := GetUserByUsername(db, username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	return u, nil
}

// UpdateUser modifies an administrator's username, email, and optionally their password hash
func UpdateUser(db *sql.DB, id int, username, email, plainPassword string) error {
	if plainPassword != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		_, err = db.Exec("UPDATE users SET username = $1, email = $2, password_hash = $3 WHERE id = $4", username, email, string(hashed), id)
		return err
	}

	_, err := db.Exec("UPDATE users SET username = $1, email = $2 WHERE id = $3", username, email, id)
	return err
}
