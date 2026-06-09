package user

import (
	"database/sql"
	"fmt"
	"megome/internal/services/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetUserByEmail(email string) (*types.User, error) {
	rows, err := s.db.Query("SELECT * FROM users WHERE email = ?", email)
	if err != nil {
		return nil, err
	}

	u := new(types.User)
	for rows.Next() {
		u, err = scanRowIntoUser(rows)
		if err != nil {
			return nil, err
		}
	}

	if u.ID == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return u, nil
}

func (s *Store) GetUserByEmailOrUsername(input string) (*types.User, error) {
	rows, err := s.db.Query("SELECT * FROM users WHERE email = ? OR username = ?", input, input)
	if err != nil {
		return nil, err
	}

	u := new(types.User)
	for rows.Next() {
		u, err = scanRowIntoUser(rows)
		if err != nil {
			return nil, err
		}
	}

	if u.ID == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return u, nil
}

func scanRowIntoUser(rows *sql.Rows) (*types.User, error) {
	user := new(types.User)

	err := rows.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Store) GetUserByID(id int) (*types.User, error) {
	rows, err := s.db.Query("SELECT * FROM users WHERE id = ?", id)
	if err != nil {
		return nil, err
	}

	u := new(types.User)
	for rows.Next() {
		u, err = scanRowIntoUser(rows)
		if err != nil {
			return nil, err
		}
	}

	if u.ID == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return u, nil
}

func (s *Store) CreateUser(user types.User) (*types.User, error) {
	result, err := s.db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
		user.Username,
		user.Email,
		user.Password,
	)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	user.ID = int(id)

	return &user, nil
}

func (s *Store) GetOAuthAccount(
	provider string,
	providerUserID string,
) (*types.OAuthAccount, error) {

	row := s.db.QueryRow(`
		SELECT
			id,
			user_id,
			provider,
			provider_user_id,
			email,
			created_at,
			updated_at
		FROM oauth_accounts
		WHERE provider = ?
		AND provider_user_id = ?
	`,
		provider,
		providerUserID,
	)

	var account types.OAuthAccount

	err := row.Scan(
		&account.ID,
		&account.UserID,
		&account.Provider,
		&account.ProviderUserID,
		&account.Email,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (s *Store) CreateOAuthAccount(
	account types.OAuthAccount,
) error {

	_, err := s.db.Exec(`
		INSERT INTO oauth_accounts
		(
			user_id,
			provider,
			provider_user_id,
			email
		)
		VALUES (?, ?, ?, ?)
	`,
		account.UserID,
		account.Provider,
		account.ProviderUserID,
		account.Email,
	)

	return err
}
