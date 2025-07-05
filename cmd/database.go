package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type database struct {
	conn *pgxpool.Pool
}

func NewDatabase(dbUrl string) *database {

	config, err := pgxpool.ParseConfig(dbUrl)

	if err != nil {
		log.Fatalln("error occurred while connecting to databse, Error: ", err.Error())
	}

	config.MaxConns = 8
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(context.Background(), config)

	if err != nil {
		log.Fatalln("error occurred while connecting to database, Error: ", err.Error())
	}

	return &database{
		conn: pool,
	}
}

func (db *database) CheckDatabaseConnection() {
	if err := db.conn.Ping(context.Background()); err != nil {
		log.Fatalln("failed to ping to the database, Error: ", err.Error())
	}
	log.Println("connected to the database")
}

func (db *database) CloseConnection() {
	db.conn.Close()
	log.Println("database connection closed")
}
