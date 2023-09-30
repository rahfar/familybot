package db

import (
	"database/sql"
)

func InsertMessage(db *sql.DB, text string, fromID, chatID, date int) error {
	insertSQL := `
        INSERT INTO messages (text, from_id, chat_id, date)
        VALUES (?, ?, ?, ?);
    `

	// Prepare the SQL statement
	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute the prepared statement with the values
	_, err = stmt.Exec(text, fromID, chatID, date)
	if err != nil {
		return err
	}

	return nil
}
