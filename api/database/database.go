package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var Ctx context.Context = context.Background()

type dbClient struct {
	rdClient *redis.Client
}

var client dbClient

func InitRedis(dbAddress string, password string) error {

	rdb := redis.NewClient(&redis.Options{
		Addr:     dbAddress,
		Password: password,
		DB:       0,
	})
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("error connecting db: %v", err)
	}
	//defer cancel()
	client = dbClient{
		rdClient: rdb,
	}
	return nil
}

func SetValue(ctx context.Context, key string, value interface{}, timeout time.Duration) error {
	err := client.rdClient.Set(ctx, key, value, timeout).Err()
	if err != nil {
		log.Println("error at set value:", err)
	}
	return err
}

func GetValue(ctx context.Context, key string) (string, error) {
	value, err := client.rdClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", err
	}
	if err != nil {
		log.Println("error at get value:", err)
	}
	return value, err
}

func Increment(ctx context.Context, key string) error {
	err := client.rdClient.Incr(ctx, key).Err()
	if err != nil {
		log.Println("error at database.incr ", err)
		return err
	}
	return nil
}

func GetTTl(ctx context.Context, key string) (time.Duration, error) {
	live, err := client.rdClient.TTL(ctx, key).Result()
	if err != nil {
		log.Println("error at getTTl:", err)
		return 0, err
	}
	return live, nil
}

func Decrement(ctx context.Context, key string) error {
	err := client.rdClient.Decr(ctx, key).Err()
	return err
}
