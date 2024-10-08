package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID        uint64 `json:"id"`
	AccountID string `json:"account_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       uint8  `json:"age"`
}

func ConnectDB() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DATABASE"))
	return sql.Open("mysql", dsn)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	dbConn, err := ConnectDB()
	if err != nil {
		http.Error(w, "Failed to connect to DB", http.StatusInternalServerError)
		return
	}
	defer dbConn.Close()

	// バリデーションの実行
	if err := ValidateUser(user, dbConn); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := "INSERT INTO users (account_id, first_name, last_name, age) VALUES (?, ?, ?, ?)"
	result, err := dbConn.Exec(query, user.AccountID, user.FirstName, user.LastName, user.Age)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to retrieve user ID", http.StatusInternalServerError)
		return
	}
	user.ID = uint64(id)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/users/")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var updates User
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	dbConn, err := ConnectDB()
	if err != nil {
		http.Error(w, "Failed to connect to DB", http.StatusInternalServerError)
		return
	}
	defer dbConn.Close()

	// 既存ユーザー情報を取得
	var existingUser User
	query := "SELECT id, account_id, first_name, last_name, age FROM users WHERE id = ?"
	err = dbConn.QueryRow(query, id).Scan(&existingUser.ID, &existingUser.AccountID, &existingUser.FirstName, &existingUser.LastName, &existingUser.Age)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to retrieve user", http.StatusInternalServerError)
		return
	}

	// バリデーションの実行（変更があれば上書きしてバリデーション）
	if updates.FirstName != "" {
		existingUser.FirstName = updates.FirstName
	}
	if updates.LastName != "" {
		existingUser.LastName = updates.LastName
	}
	if updates.Age != 0 {
		existingUser.Age = updates.Age
	}

	if err := ValidateUser(existingUser, dbConn); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updateQuery := "UPDATE users SET first_name = ?, last_name = ?, age = ? WHERE id = ?"
	_, err = dbConn.Exec(updateQuery, existingUser.FirstName, existingUser.LastName, existingUser.Age, id)
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(existingUser)
}

func main() {
	http.HandleFunc("/users", CreateUser)
	http.HandleFunc("/users/", UpdateUser) // 更新時のエンドポイント
	http.ListenAndServe(":8080", nil)
}
