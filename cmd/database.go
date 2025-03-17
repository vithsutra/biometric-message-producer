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
	pool, err := pgxpool.New(context.Background(), dbUrl)

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
}
