package config

import (
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/gofigure"
)

// Config holds the session handler configuration
type Config struct {
	gofigure          interface{} `order:"env,flag"`
	DefaultExpiration string      `env:"DEFAULT_SESSION_EXPIRATION" flag:"default-expiration" flagDesc:"Default Expiration"`
	CookieName        string      `env:"COOKIE_NAME"                flag:"cookie-name"        flagDesc:"Cookie Name"`
	CookieSecret      string      `env:"COOKIE_SECRET"              flag:"cookie-secret"      flagDesc:"Cookie Secret"`
	CacheServer       string      `env:"CACHE_SERVER"               flag:"cache-server"       flagDesc:"Cache Server"`
	CacheDB           int         `env:"CACHE_DB"                   flag:"cache-db"           flagDesc:"Cache DB"`
	CachePassword     string      `env:"CACHE_PASSWORD"             flag:"cache-password"     flagDesc:"Cache Password"`
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
