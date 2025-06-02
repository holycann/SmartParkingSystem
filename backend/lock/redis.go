package lock

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/go-redsync/redsync/v4"
	redsync_goredis "github.com/go-redsync/redsync/v4/redis/goredis/v9"
	goredis "github.com/redis/go-redis/v9"
)

var (
	Redsync *redsync.Redsync
)

func InitializeRedisLock() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	client := goredis.NewClient(&goredis.Options{
		Addr: redisAddr,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	pool := redsync_goredis.NewPool(client)
	Redsync = redsync.New(pool)

	log.Println("Redis lock initialized successfully")
}

func AcquireLock(name string, ttl time.Duration) (*redsync.Mutex, error) {
	mutex := Redsync.NewMutex(name, redsync.WithExpiry(ttl))
	if err := mutex.Lock(); err != nil {
		return nil, err
	}
	return mutex, nil
}

func ReleaseLock(mutex *redsync.Mutex) error {
	_, err := mutex.Unlock()
	return err
}
