package state

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/companieshouse/chs.go/log"
	"github.com/vmihailenco/msgpack"
	redis "gopkg.in/redis.v5"
)

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

// Store will take a store struct, validate it, and attempt to save it
func (s *Store) Store() error {
	log.Event("debug", "", s.Data)
	if err := s.validateStore(); err != nil {
		return err
	}

	c := &Cache{}
	if err := c.setRedisClient(); err != nil {
		return err
	}

	if err := c.setSession(s); err != nil {
		return err
	}

	return nil
}

// RegenerateID refreshes the token against the Store struct
func (s *Store) regenerateID() error {
	idOctets, err := strconv.Atoi(os.Getenv("ID_OCTETS"))
	if err != nil {
		return err
	}

	octets := make([]byte, idOctets)

	if _, err := rand.Read(octets); err != nil {
		return err
	}

	s.ID = encodeBase64(octets)
	return nil
}

func (s *Store) generateSignature() {

}

// SetupExpiration will set the 'Expires' variable against the Store
// This should only be called if an expiration is not already set
func (s *Store) setupExpiration() error {
	now := uint64(time.Now().Unix())

	expirationPeriod, err := strconv.ParseUint(os.Getenv("DEFAULT_EXPIRATION"), 0, 64)
	if err != nil {
		return err
	}

	s.Expires = now + expirationPeriod
	s.Data["last_access"] = now
	return nil
}

// ValidateStore will be called to authenticate the session store
func (s *Store) validateStore() error {
	if s.ID == "" {
		if err := s.regenerateID(); err != nil {
			return err
		}
	}

	if s.Expires == 0 {
		if err := s.setupExpiration(); err != nil {
			return err
		}
	}

	if s.Data == nil {
		return errors.New("No session data to store!")
	}

	return nil
}

/*
   CACHE
*/

// SetRedisClient into the Cache struct
func (c *Cache) setRedisClient() error {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if _, err := client.Ping().Result(); err != nil {
		return err
	}

	c.connection = client
	return nil
}

func (c *Cache) getSession() {

}

// SetSession will take the valid Store object and save it in Redis
func (c *Cache) setSession(s *Store) error {
	msgpackEncodedData, err := encodeMsgPack(s.Data)
	if err != nil {
		return err
	}
	b64EncodedData := encodeBase64(msgpackEncodedData)

	_, err = c.connection.Set(s.ID, b64EncodedData, 0).Result()
	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) decodeSessionBase64() {

}

// EncodeBase64 takes a byte array and base 64 encodes it
func encodeBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

func (c *Cache) decodeSessionMsgPack() {

}

// EncodeMsgPack performs message pack encryption
// Currently this takes a map[string]interface{} parameter because we only
// want to message pack encode JSON objects
func encodeMsgPack(data map[string]interface{}) ([]byte, error) {
	var encoded []byte
	encBuf := bytes.NewBuffer(encoded)
	enc := msgpack.NewEncoder(encBuf)

	if err := enc.Encode(data); err != nil {
		return nil, err
	}

	return encBuf.Bytes(), nil
}
