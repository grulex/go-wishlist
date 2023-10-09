package db

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"log"
)

type Config struct {
	Host     string
	Port     uint16
	Database string
	User     string
	Password string
}

func CreateDBConnection(config Config) (*sqlx.DB, error) {
	connConfig := parseConfig(config)
	nativeDB := stdlib.OpenDB(*connConfig)
	if err := nativeDB.Ping(); err != nil {
		return nil, err
	}
	return sqlx.NewDb(nativeDB, "pgx").Unsafe(), nil
}

func parseConfig(config Config) *pgx.ConnConfig {
	connConfig, err := pgx.ParseConfig("")
	if err != nil {
		log.Fatal(err)
	}

	connConfig.Host = config.Host
	connConfig.Port = config.Port
	connConfig.Database = config.Database
	connConfig.User = config.User
	connConfig.Password = config.Password
	connConfig.RuntimeParams = map[string]string{"standard_conforming_strings": "on"}

	return connConfig
}
