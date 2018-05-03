package config

import (
	"os"
	"strconv"

	"github.com/companieshouse/chs.go/log"
	"github.com/ian-kent/gofigure"
	redis "gopkg.in/redis.v5"
)

// Config holds the session handler configuration
type Config struct {
	gofigure          interface{} `order:"env,flag"`
	DefaultExpiration string      `env:"DEFAULT_EXPIRATION"		flag:"default-expiration"   flagDesc:"Default Expiration"`
	CookieName        string      `env:"COOKIE_NAME"					flag:"cookie-name"          flagDesc:"Cookie Name"`
	CookieSecret      string      `env:"COOKIE_SECRET"				flag:"cookie-secret"        flagDesc:"Cookie Secret"`
}

var cfg *Config
var redisOpt *redis.Options

// Get returns a populated Config struct
func Get() *Config {

	if cfg != nil {
		return cfg
	}

	cfg = &Config{}

	if err := gofigure.Gofigure(cfg); err != nil {
		log.Error(err)
		return nil
	}

	return cfg
}

// GetRedisOptions returns an initialised RedisOptions struct
func GetRedisOptions() *redis.Options {
	if redisOpt != nil {
		return redisOpt
	}

	db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		db = 0
	}

	redisOpt = &redis.Options{
		Addr:     os.Getenv("REDIS_SERVER"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	}

	return redisOpt
}
