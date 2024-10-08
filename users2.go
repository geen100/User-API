package main

import (
	"database/sql"
	"errors"
	"regexp"
	"unicode/utf8"
)

// アカウントIDの正規表現: 4文字以上、64文字以下で、英数字およびアンダースコアのみ許可
var validID = regexp.MustCompile(`^[a-zA-Z0-9_]{4,64}$`)

// ValidateAccountID - アカウントIDのバリデーション
func ValidateAccountID(accountID string, currentID uint64, db *sql.DB) error {
	if !validID.MatchString(accountID) {
		return errors.New("account ID must be between 4 and 64 characters and contain only letters, numbers, and underscores")
	}

	// アカウントIDが既に存在しないか確認
	var existingID uint64
	query := "SELECT id FROM users WHERE account_id = ? AND id != ?"
	err := db.QueryRow(query, accountID, currentID).Scan(&existingID)
	if err != sql.ErrNoRows {
		return errors.New("account ID already exists")
	}
	return nil
}

// ValidateNameLength - 名前の長さのバリデーション
func ValidateNameLength(name string) error {
	if utf8.RuneCountInString(name) > 100 {
		return errors.New("name must be 100 characters or less")
	}
	return nil
}

// ValidateUser - ユーザー全体のバリデーション
func ValidateUser(user User, db *sql.DB) error {
	if err := ValidateAccountID(user.AccountID, user.ID, db); err != nil {
		return err
	}
	if err := ValidateNameLength(user.FirstName); err != nil {
		return errors.New("invalid first name: " + err.Error())
	}
	if err := ValidateNameLength(user.LastName); err != nil {
		return errors.New("invalid last name: " + err.Error())
	}
	return nil
}
