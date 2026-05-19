package databases

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

// ConnectRedis connects to Redis and returns the client.
// Returns nil (without crashing) if Redis is unavailable
func ConnectRedis() *redis.Client {
	var opts *redis.Options
	var err error

	redisURL := os.Getenv("REDIS_URL")
	if redisURL != "" {
		opts, err = redis.ParseURL(redisURL)
		if err != nil {
			fmt.Println("Cannot parse REDIS_URL, running without cache")
			return nil
		}
	} else {
		opts = &redis.Options{
			Addr: fmt.Sprintf("%s:%s",
				os.Getenv("REDIS_HOST"),
				os.Getenv("REDIS_PORT"),
			),
		}
	}

	client := redis.NewClient(opts)
	if _, err = client.Ping(context.Background()).Result(); err != nil {
		fmt.Println("Redis unavailable, running without cache")
		return nil
	}

	fmt.Println("Connected to Redis successfully")
	return client
}