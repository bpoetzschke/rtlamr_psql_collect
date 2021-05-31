package database

import (
	"database/sql"
	"fmt"

	"github.com/caarlos0/env/v6"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type config struct {
	Host         string `env:"DB_HOST,required"`
	Port         int32  `env:"DB_PORT,required"`
	User         string `env:"DB_USER,required"`
	Password     string `env:"DB_PASSWORD,required"`
	DatabaseName string `env:"DB_DATABASE,required"`
}

func Init() (*sql.DB, error) {
	db, err := connect()
	if err != nil {
		return nil, err
	}

	err = createSchema(db)
	if err != nil {
		closeErr := db.Close()
		if closeErr != nil {
			log.Errorf("Error while closing db connection: %s", closeErr)
		}
		return nil, err
	}

	return db, nil
}

func connect() (*sql.DB, error) {
	dbConfig := config{}
	err := env.Parse(&dbConfig)
	if err != nil {
		return nil, fmt.Errorf("error while reading database config %w", err)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DatabaseName)

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

func createSchema(db *sql.DB) error {
	createSchemaQuery := `
		CREATE TABLE IF NOT EXISTS rtlamr_data (
			meter_id varchar(255),
			meter_type int,
			current_reading bigint,
			difference int,
			created_at timestamp without time zone,
			primary key (meter_id, current_reading)
		);
	`
	_, err := db.Exec(createSchemaQuery)
	if err != nil {
		return err
	}

	return nil
}
