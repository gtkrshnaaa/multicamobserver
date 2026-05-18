package models

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// GetBroadcasterByNodeID retrieves broadcaster details from PostgreSQL
func GetBroadcasterByNodeID(db *sql.DB, nodeID string) (*Broadcaster, error) {
	row := db.QueryRow("SELECT id, node_id, name, password_hash, created_at FROM broadcasters WHERE node_id = $1", nodeID)

	var b Broadcaster
	err := row.Scan(&b.ID, &b.NodeID, &b.Name, &b.PasswordHash, &b.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("broadcaster node not found")
		}
		return nil, err
	}

	return &b, nil
}

// AuthenticateBroadcaster checks if camera node credentials are valid
func AuthenticateBroadcaster(db *sql.DB, nodeID, password string) (*Broadcaster, error) {
	b, err := GetBroadcasterByNodeID(db, nodeID)
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
	rows, err := db.Query("SELECT id, node_id, name, created_at FROM broadcasters ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Broadcaster
	for rows.Next() {
		var b Broadcaster
		err := rows.Scan(&b.ID, &b.NodeID, &b.Name, &b.CreatedAt)
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
