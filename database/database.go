package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// TODO read from env vars
const (
	host = "localhost"
	port = 5432
	user = "admin"
	password = "admin"
	dbname = "rtlamr"
)

func Init() (*sql.DB, error) {
	db, err := connect()
	if err != nil {
		return nil, err
	}

	err = createSchema(db)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func connect() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func createSchema(db *sql.DB) (error) {
	createSchemaQuery := `
		CREATE TABLE IF NOT EXISTS rtlamr_data (
			meter_id varchar(255),
			meter_type int,
			consumption bigint,
			usage int,
			created_at timestamp without time zone default now()
		);
	`
	_, err := db.Exec(createSchemaQuery)
	if err != nil {
		return err
	}

	return nil
}
