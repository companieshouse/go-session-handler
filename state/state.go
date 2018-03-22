package state

import redis "gopkg.in/redis.v5"

type Store struct {
	ID      string
	Expires uint64
	Data    map[string]interface{}
}

type Cache struct {
	connection *redis.Client
}

/*
   STORE
*/

func (s *Store) Load() {

}

func (s *Store) Store() {

}

func (s *Store) regenerateID() {

}

func (s *Store) generateSignature() {

}

/*
   CACHE
*/

func (c *Cache) getRedisClient() {

}

func (c *Cache) getSession() {

}

func (c *Cache) setSession() {

}

func (c *Cache) decodeSessionBase64() {

}

func (c *Cache) encodeSessionBase64() {

}

func (c *Cache) decodeSessionMsgPack() {

}

func (c *Cache) encodeSessionMsgPack() {

}
