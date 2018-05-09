package config

import (
	"github.com/companieshouse/chs.go/log"
	"github.com/ian-kent/gofigure"
)

// Config holds the session handler configuration
type Config struct {
	gofigure          interface{} `order:"env,flag"`
	DefaultExpiration string      `env:"DEFAULT_EXPIRATION"		flag:"default-expiration"   flagDesc:"Default Expiration"`
	CookieName        string      `env:"COOKIE_NAME"					flag:"cookie-name"          flagDesc:"Cookie Name"`
	CookieSecret      string      `env:"COOKIE_SECRET"				flag:"cookie-secret"        flagDesc:"Cookie Secret"`
	RedisServer       string      `env:"REDIS_SERVER"					flag:"redis-server"        	flagDesc:"Redis Server"`
	RedisDB           int         `env:"REDIS_DB"							flag:"redis-db"       			flagDesc:"Redis DB"`
	RedisPassword     string      `env:"REDIS_PASSWORD"				flag:"redis-password"       flagDesc:"Redis Password"`
}

var cfg *Config

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
