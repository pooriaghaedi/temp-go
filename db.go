package main

import (
	"database/sql"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	dsn := os.Getenv("DB_DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}
	db.AutoMigrate(&User{}, &Tweet{}, &Like{}, &Friend{})
	return db
}

func addUser(db *sql.DB, email string) (userid int, err error) {
	// Start a new transaction

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}

	fmt.Println(email)
	// Check if the user already exists
	var id int
	// err = tx.QueryRow("SELECT id FROM users WHERE email = 'test' ").Scan(&id)
	// err = tx.QueryRow("SELECT id FROM users WHERE username = ?", "poori").Scan(&id)
	err = tx.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&id)

	if err != nil && err != sql.ErrNoRows {
		// An error occurred while querying for the user
		tx.Rollback()
		fmt.Println("error1")
		return 0, err

	} else if err == sql.ErrNoRows {
		// The user does not exist, insert a new user
		_, err = tx.Exec("INSERT INTO users (email, created_at) VALUES ($1, CURRENT_TIMESTAMP)", email)
		if err != nil {
			tx.Rollback()
			fmt.Println("error2")
			return 0, err
		}
	} else {
		// The user exists, update is_valid
		_, err = tx.Exec("UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = $1", id)
		if err != nil {
			tx.Rollback()
			fmt.Println("error3")
			return id, err
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func addTwitt(db *sql.DB, tg_id int, twitt string) {
	// Start a new transaction

	tx, err := db.Begin()
	if err != nil {
		return
	}

	// Check if the user already exists
	var id int
	err = tx.QueryRow("SELECT id FROM user WHERE id = ?", id).Scan(&id)

	_, err = tx.Exec("INSERT INTO twitt (user_id, context, created_at) VALUES (?, ?, CURRENT_TIMESTAMP)", id, twitt)
	fmt.Println("added")
	if err != nil {
		tx.Rollback()
		return
	}
	// return nil
}

// func getUsers(db *sql.DB) []User {
// 	rows, _ := db.Query("SELECT id, id, username, is_valid, credits, created_at FROM user")
// 	defer rows.Close()

// 	var users []User
// 	for rows.Next() {
// 		var user User
// 		rows.Scan(&user.ID, &user.TelegramId, &user.Username, &user.IsValid, &user.Credits, &user.CreatedAt)
// 		user = append(user, user)
// 	}

// 	return user
// }

func isUserValid(db *sql.DB, id int) bool {
	var isValid bool
	err := db.QueryRow("SELECT is_valid FROM user WHERE id = ?", id).Scan(&isValid)
	if err != nil {
		return false
	}

	return isValid
}
