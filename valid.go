func ValidateAccountID(ctx context.Context, accountID string, currentID uint64, db *sql.DB) error {
	if !validID.MatchString(accountID) {
		return errors.New("account ID must be between 4 and 64 characters and contain only letters, numbers, and underscores")
	}

	var existingID uint64
	query := "SELECT id FROM users WHERE account_id = ? AND id != ?"
	// QueryRowContextを使って、contextからDBクエリを実行する
	err := db.QueryRowContext(ctx, query, accountID, currentID).Scan(&existingID)
	if err != nil {
		// エラーチェックでsql.ErrNoRowsを正確に確認
		if err == sql.ErrNoRows {
			return nil // ユニークなアカウントID
		}
		return fmt.Errorf("database query error: %w", err)
	}

	// 同じアカウントIDが既に存在する場合のエラー
	return errors.New("account ID already exists")
}

func ValidateNameLength(name string) error {
	if utf8.RuneCountInString(name) > 100 {
		return errors.New("name must be 100 characters or less")
	}
	return nil
}

func ValidateUser(ctx context.Context, user User, db *sql.DB) error {
	// アカウントIDのバリデーションにcontextを渡す
	if err := ValidateAccountID(ctx, user.AccountID, user.ID, db); err != nil {
		return err
	}

	// 名前の長さのバリデーション
	if err := ValidateNameLength(user.FirstName); err != nil {
		return fmt.Errorf("invalid first name: %w", err)
	}
	if err := ValidateNameLength(user.LastName); err != nil {
		return fmt.Errorf("invalid last name: %w", err)
	}
	return nil
}
