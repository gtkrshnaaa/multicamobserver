package models

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// GetBroadcasterByNodeID retrieves broadcaster details from PostgreSQL
func GetBroadcasterByNodeID(db *sql.DB, nodeID string) (*Broadcaster, error) {
	row := db.QueryRow("SELECT id, node_id, username, name, password_hash, created_at FROM broadcasters WHERE node_id = $1", nodeID)

	var b Broadcaster
	err := row.Scan(&b.ID, &b.NodeID, &b.Username, &b.Name, &b.PasswordHash, &b.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("broadcaster node not found")
		}
		return nil, err
	}

	return &b, nil
}

// GetBroadcasterByUsername retrieves broadcaster details from PostgreSQL using username
func GetBroadcasterByUsername(db *sql.DB, username string) (*Broadcaster, error) {
	row := db.QueryRow("SELECT id, node_id, username, name, password_hash, created_at FROM broadcasters WHERE username = $1", username)

	var b Broadcaster
	err := row.Scan(&b.ID, &b.NodeID, &b.Username, &b.Name, &b.PasswordHash, &b.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("broadcaster node not found")
		}
		return nil, err
	}

	return &b, nil
}

// AuthenticateBroadcaster checks if camera node credentials are valid using username and password
func AuthenticateBroadcaster(db *sql.DB, username, password string) (*Broadcaster, error) {
	b, err := GetBroadcasterByUsername(db, username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(b.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("invalid camera credentials")
	}

	return b, nil
}

// GetAllBroadcasters retrieves all registered camera nodes in the database
func GetAllBroadcasters(db *sql.DB) ([]*Broadcaster, error) {
	rows, err := db.Query("SELECT id, node_id, username, name, created_at FROM broadcasters ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Broadcaster
	for rows.Next() {
		var b Broadcaster
		err := rows.Scan(&b.ID, &b.NodeID, &b.Username, &b.Name, &b.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, &b)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

// CreateBroadcaster registers a new physical camera broadcaster node securely
func CreateBroadcaster(db *sql.DB, username, name, plainPassword string) (*Broadcaster, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	row := db.QueryRow(
		"INSERT INTO broadcasters (username, name, password_hash) VALUES ($1, $2, $3) RETURNING id, node_id, username, name, created_at",
		username, name, string(hashed),
	)

	var b Broadcaster
	err = row.Scan(&b.ID, &b.NodeID, &b.Username, &b.Name, &b.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

// UpdateBroadcaster updates camera node credentials dynamically
func UpdateBroadcaster(db *sql.DB, id int, username, name, plainPassword string) error {
	if plainPassword != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		_, err = db.Exec(
			"UPDATE broadcasters SET username = $1, name = $2, password_hash = $3 WHERE id = $4",
			username, name, string(hashed), id,
		)
		return err
	}

	_, err := db.Exec("UPDATE broadcasters SET username = $1, name = $2 WHERE id = $3", username, name, id)
	return err
}

// DeleteBroadcaster removes a physical camera broadcaster node from the system
func DeleteBroadcaster(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM broadcasters WHERE id = $1", id)
	return err
}

// PurgeAllBroadcasters permanently deletes all broadcaster camera nodes from the system
func PurgeAllBroadcasters(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM broadcasters")
	return err
}
