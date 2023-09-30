package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

const CREATE_MSG_TABLE = `
CREATE TABLE IF NOT EXISTS messages (
	id int PRIMARY KEY,
	date int NOT NULL,
   	from_id int NOT NULL,
	chat_id int NOT NULL,
	text text NOT NULL
)
`

func InitDB(dsnURI string) {
	db, err := sql.Open("sqlite", dsnURI)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(CREATE_MSG_TABLE)
	if err != nil {
		panic(err)
	}
}
