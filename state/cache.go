/*
Package state contains the go implementation for storing and loading the Session
from the cache.
*/
package state

import (
	"time"

	redis "gopkg.in/redis.v5"
)

type Connection interface {
	Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(key string) *redis.StringCmd
	Del(key ...string) *redis.IntCmd
}

//Cache is the struct that contains the connection info for retrieving/saving
//The session data.
type Cache struct {
	connection Connection
}

//NewCache will properly initialise a new Cache object.
func NewCache(connectionInfo *redis.Options) (*Cache, error) {
	cache := &Cache{}

	if err := cache.setRedisClient(connectionInfo); err != nil {
		return nil, err
	}

	return cache, nil
}

/*
   CACHE
*/

//setSessionData stores the Session data in the Cache.
func (c *Cache) setSessionData(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return c.connection.Set(key, value, expiration)
}

//getSessionData loads the Session data from the Cache.
func (c *Cache) getSessionData(key string) (string, error) {
	return c.connection.Get(key).Result()
}

//deleteSessionData removes the Session data from the Cache.
func (c *Cache) deleteSessionData(key string) error {
	_, err := c.connection.Del(key).Result()
	return err
}

//setRedisClient into the Cache struct
func (c *Cache) setRedisClient(options *redis.Options) error {
	client := redis.NewClient(options)

	if _, err := client.Ping().Result(); err != nil {
		return err
	}

	c.connection = client
	return nil
}
