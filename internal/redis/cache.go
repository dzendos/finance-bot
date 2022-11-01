package redis

import (
	"log"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/e.gerasimov/telegram-bot/internal/config"
)

type Cache struct {
	pool *redis.Pool
}

func New(config *config.Service) (*Cache, error) {
	cache := &Cache{
		pool: newPool(config),
	}
	err := cache.Ping()

	return cache, err
}

func (c *Cache) Ping() error {
	conn := c.pool.Get()
	defer conn.Close()

	_, err := redis.String(conn.Do("PING"))
	if err != nil {
		return errors.Wrap(err, "cannot redis.Do")
	}
	return nil
}

func (c *Cache) Get(key string) ([]byte, error) {
	conn := c.pool.Get()
	defer conn.Close()

	var data []byte
	data, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return data, errors.Wrap(err, "cannot redis.Bytes")
	}
	return data, err
}

func (c *Cache) Set(key string, value []byte) error {
	conn := c.pool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", key, value)
	if err != nil {
		return errors.Wrap(err, "cannot conn.Do")
	}

	return nil
}

func (c *Cache) Exists(key string) (bool, error) {
	conn := c.pool.Get()
	defer conn.Close()

	ok, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return ok, errors.Wrap(err, "cannot redis.Bool")
	}
	return ok, err
}

func (c *Cache) Delete(key string) error {
	conn := c.pool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", key)
	return errors.Wrap(err, "cannot conn.Do")
}

func (c *Cache) Close() {
	log.Println("Closing cache")
	c.pool.Close()
}
