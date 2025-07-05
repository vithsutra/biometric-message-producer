package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisRepository struct {
	redisConn *redis.Client
}

func NewRedisRepository(conn *redis.Client) *redisRepository {
	return &redisRepository{
		redisConn: conn,
	}
}

func (repo *redisRepository) CheckMessageDuplication(messageId string) (bool, error) {
	set, err := repo.redisConn.SetNX(context.Background(), messageId, "1", 10*time.Minute).Result()

	if err != nil {
		return false, err
	}

	return set, nil
}
