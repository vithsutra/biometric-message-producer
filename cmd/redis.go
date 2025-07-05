package main

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

type redisConn struct {
	client *redis.Client
}

func NewRedisConnection(redisAddress string) *redisConn {
	options, err := redis.ParseURL(redisAddress)

	if err != nil {
		log.Fatalf("error occurred while connecting to redis, Error: %v\n", err.Error())
	}

	rdb := redis.NewClient(options)

	return &redisConn{
		client: rdb,
	}
}

func (c *redisConn) CheckRedisConnection() {
	_, err := c.client.Ping(context.Background()).Result()

	if err != nil {
		log.Fatalln("error occurred while connecting to redis, Error: ", err.Error())
	}

	log.Println("connected to redis")
}

func (c *redisConn) CloseConnection() {
	if err := c.client.Close(); err != nil {
		log.Fatalln("error occurred while closing the redis connection, Error: ", err.Error())
	}
	log.Println("redis connection closed")
}
