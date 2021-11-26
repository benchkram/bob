package database

import (
	"github.com/go-redis/redis/v8"
)

type Database struct {
	*redis.Client
}

func New(addr string) *Database {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	return &Database{
		rdb,
	}
}
