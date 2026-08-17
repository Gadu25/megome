package emailverification

import (
	"database/sql"
	"errors"
	"time"
)

var ErrInvalidOTP = errors.New("invalid or expired verification code")
var ErrTooManyAttempts = errors.New("too many failed attempts, a new code has been sent")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveOTP(userId int, email string, otpHash string, exp time.Time) error {
	_, err := r.db.Exec(`
		INSERT INTO email_verification_otps (userId, email, otpHash, expiresAt, createdAt)
		VALUES (?, ?, ?, ?, ?)
	`, userId, email, otpHash, exp, time.Now())
	return err
}

func (r *Repository) LastOTPSentAt(userId int) (time.Time, error) {
	var createdAt time.Time
	err := r.db.QueryRow(`
		SELECT createdAt
		FROM email_verification_otps
		WHERE userId = ?
		ORDER BY createdAt DESC
		LIMIT 1
	`, userId).Scan(&createdAt)
	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	return createdAt, err
}

func (r *Repository) DeleteOTPs(userId int) error {
	_, err := r.db.Exec("DELETE FROM email_verification_otps WHERE userId = ?", userId)
	return err
}

func (r *Repository) VerifyOTP(email string, otpHash string) (int, error) {
	var userId int
	var storedHash string
	var expiresAt time.Time
	var usedAt sql.NullTime
	var failedAttempts int

	err := r.db.QueryRow(`
		SELECT userId, otpHash, expiresAt, usedAt, failedAttempts
		FROM email_verification_otps
		WHERE email = ?
		ORDER BY createdAt DESC
		LIMIT 1
	`, email).Scan(&userId, &storedHash, &expiresAt, &usedAt, &failedAttempts)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrInvalidOTP
		}
		return 0, err
	}

	if usedAt.Valid {
		return 0, ErrInvalidOTP
	}

	if time.Now().After(expiresAt) {
		return 0, ErrInvalidOTP
	}

	if storedHash != otpHash {
		_, updateErr := r.db.Exec(`
			UPDATE email_verification_otps
			SET failedAttempts = failedAttempts + 1
			WHERE email = ?
		`, email)
		if updateErr != nil {
			return 0, updateErr
		}

		if failedAttempts+1 >= MaxOTPAttempts {
			_, deleteErr := r.db.Exec("DELETE FROM email_verification_otps WHERE email = ?", email)
			if deleteErr != nil {
				return 0, deleteErr
			}
			return 0, ErrTooManyAttempts
		}

		return 0, ErrInvalidOTP
	}

	return userId, nil
}

func (r *Repository) MarkVerified(userId int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		UPDATE users
		SET emailVerifiedAt = ?
		WHERE id = ?
	`, time.Now(), userId); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		DELETE FROM email_verification_otps
		WHERE userId = ?
	`, userId); err != nil {
		return err
	}

	return tx.Commit()
}
