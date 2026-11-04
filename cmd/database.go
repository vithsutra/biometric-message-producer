package main

import (
	"context"
	"log"
	"time"

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

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)

	if err != nil {
		log.Fatalln("error occurred while connecting to database, Error: ", err.Error())
	}

	return &database{
		conn: pool,
	}
}

func (db *database) CheckDatabaseConnection() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.conn.Ping(ctx); err != nil {
		log.Fatalln("failed to ping to the database, Error: ", err.Error())
	}
	log.Println("connected to the database")
}

func (db *database) CloseConnection() {
	db.conn.Close()
	log.Println("database connection closed")
}
