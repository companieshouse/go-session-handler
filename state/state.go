/*
Package state contains the go implementation for storing and loading the Session
from the cache.
*/

package state

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/go-session-handler/encoding"
	"github.com/companieshouse/go-session-handler/exception"
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

const cookieName = "__SID"

//Multiples of 3 bytes avoids = padding in base64 string
//7 * 3 bytes = (21/3) * 4 = 28 base64 characters
const idOctets = 7 * 3
const signatureStart = (idOctets * 4) / 3
const signatureLength = 27 //160 bits, base 64 encoded
const cookieValueLength = signatureStart + signatureLength

const defaultExpiration = "DEFAULT_EXPIRATION"
const idOctetsStr = "ID_OCTETS"

/*
   STORE
*/

//Load is used to try and get a session from the cache. If it succeeds it will
//load the session, otherwise it will return an error.
func (s *Store) Load(req *http.Request) {

	cookie, _ := s.getCookieFromRequest(req)

	s.validateCookieSignature(req, cookie.Value)
	s.extractAndValidateCookieSignatureParts(req, cookie.Value)

	storedSession, err := s.getStoredSession(req)

	if err != nil {
		log.InfoR(req, err.Error())
	}

	if storedSession == nil {

	}
}

// Store will take a store struct, validate it, and attempt to save it
func (s *Store) Store() error {

	jsonData, _ := json.Marshal(s.Data)
	log.Info("Attempting to store session with data: " + string(jsonData))

	if err := s.validateStore(); err != nil {
		log.Error(fmt.Errorf("Error validating store: %s", err))
		return err
	}

	c := &Cache{}
	if err := c.setRedisClient(); err != nil {
		log.Error(fmt.Errorf("Error connecting to Redis client: %s", err))
		return err
	}

	if err := c.setSession(s); err != nil {
		log.Error(fmt.Errorf("Error setting session data: %s", err))
		return err
	}

	log.Info("Session data successfully stored with ID: " + s.ID)
	return nil
}

//Delete will clear the requested session from the backing store. Note: Delete
//does not clear the loaded session. The Clear method will take care of that.
//If the string passed in is nil, it will delete the session with an id the same
//as that of s.ID
func (s *Store) Delete(req *http.Request, id *string) {
	sessionID := s.ID

	if id != nil && len(*id) > 0 {
		sessionID = *id
	}

	cache := s.getRedisClientFromCache()

	_, err := cache.connection.Del(sessionID).Result()

	if err != nil {
		log.InfoR(req, err.Error())
	}
}

//Clear destroys the current loaded session and removes it from the backing
//store. It will also regenerate the session ID.
func (s *Store) Clear(req *http.Request) {
	s.Data = nil
	s.Delete(req, nil) //Delete the previously stored Session
	s.regenerateID()
}

// regenerateID refreshes the token against the Store struct
func (s *Store) regenerateID() error {
	idOctets, err := strconv.Atoi(os.Getenv(idOctetsStr))
	if err != nil {
		return exception.EnvironmentVariableMissingException(idOctetsStr)
	}

	octets := make([]byte, idOctets)

	if _, err := rand.Read(octets); err != nil {
		return err
	}

	s.ID = encoding.EncodeBase64(octets)
	return nil
}

func (s *Store) generateSignature() string {
	return ""
}

// setupExpiration will set the 'Expires' variable against the Store
// This should only be called if an expiration is not already set
func (s *Store) setupExpiration() error {

	now := uint64(time.Now().Unix())

	expirationPeriod, err := strconv.ParseUint(os.Getenv(defaultExpiration), 0, 64)
	if err != nil {
		return exception.EnvironmentVariableMissingException(defaultExpiration)
	}

	s.Expires = now + expirationPeriod

	if s.Data != nil {
		s.Data["last_access"] = now
	}

	return nil
}

// validateStore will be called to authenticate the session store
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

//getCookieFromRequest will attempt to pull the Cookie from the request. If err
//is not nil, it will create a new Cookie and return that instead.
func (s *Store) getCookieFromRequest(req *http.Request) (*http.Cookie, error) {

	var cookie *http.Cookie
	var err error

	cookieName := os.Getenv("COOKIE_NAME")

	if cookie, err = req.Cookie(cookieName); err != nil {
		log.InfoR(req, err.Error())
		cookie = &http.Cookie{}
	}

	return cookie, err
}

//validateCookieSignature will try to validate that the length of the Cookie
//value is not equal to the calculated length of the signature
func (s *Store) validateCookieSignature(req *http.Request, cookieSignature string) {

	cookieValueLength, err := strconv.Atoi(os.Getenv("ID_LENGTH"))
	if err != nil {
		log.Error(exception.EnvironmentVariableMissingException("ID_LENGTH"))
	}

	if len(cookieSignature) != cookieValueLength {
		log.InfoR(req, "Cookie signature is not the correct length")

		s.Clear(req)
		return
	}
}

//extractAndValidateCookieSignatureParts will split the cookieSignature into
//two parts, and set the first part to s.ID, with the second part being validated
//against a generated ID.
func (s *Store) extractAndValidateCookieSignatureParts(req *http.Request, cookieSignature string) {
	s.ID = cookieSignature[0:signatureStart]
	sig := cookieSignature[signatureStart:len(cookieSignature)]

	//Validate signature is the same
	if sig != s.generateSignature() {
		s.Clear(req)
		return
	}
}

//getRedisClientFromCache will construct a new Cache and invoke getRedisClient.
func (s *Store) getRedisClientFromCache() *Cache {
	cache := &Cache{}
	cache.getRedisClient()

	return cache
}

//getStoredSession will get the session from the Cache, and validate it.
//If it is invalid, it will return an error.
func (s *Store) getStoredSession(req *http.Request) ([]byte, error) {
	cache := s.getRedisClientFromCache()
	storedSession := cache.getSession(s.ID)

	if len(storedSession) == 0 {
		err := errors.New("There is no stored session")
		return nil, err
	}

	return storedSession, nil
}

/*
   CACHE
*/

func (c *Cache) getRedisClient() {

}

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

func (c *Cache) getSession(id string) []byte {
	var a []byte
	return a
}

// SetSession will take the valid Store object and save it in Redis
func (c *Cache) setSession(s *Store) error {
	msgpackEncodedData, err := encoding.EncodeMsgPack(s.Data)
	if err != nil {
		return err
	}
	b64EncodedData := encoding.EncodeBase64(msgpackEncodedData)

	_, err = c.connection.Set(s.ID, b64EncodedData, 0).Result()
	if err != nil {
		return err
	}

	return nil
}
