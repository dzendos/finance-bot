package redis

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/config"
)

func newPool(config *config.Service) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,

		Dial: func() (redis.Conn, error) {
			address := fmt.Sprintf("%s:%s", config.GetHost(), "6379")
			c, err := redis.Dial("tcp", address)
			if err != nil {
				return nil, errors.Wrap(err, "cannot redis.Dial")
			}
			return c, nil
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return errors.Wrap(err, "cannot do ping")
		},
	}
}
