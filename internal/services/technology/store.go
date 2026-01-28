package technology

import (
	"database/sql"
	"megome/internal/services/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func GetTechnologies(userID int) (*types.Technology, error) {
	return nil, nil
}

func CreateTechnology(technology types.Technology) error {
	return nil
}

func UpdateTechnology(id int, technology types.Technology) error {
	return nil
}

func DeleteTechnology(id int) error {
	return nil
}
