/*
Package state contains the go implementation for storing and loading the Session
from the cache.
*/
package state

import (
	"crypto/rand"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/go-session-handler/encoding"
	redis "gopkg.in/redis.v5"
)

//Store is the struct that is used to load/store the session.
type Store struct {
	ID             string
	Expiration     uint64
	Expires        uint64
	Data           map[string]interface{}
	Encoder        encoding.EncodingInterface
	SessionHandler SessionHandlerInterface
	Cache          *Cache
}

//Cache is the struct that contains the connection info for retrieving/saving
//The session data.
type Cache struct {
	connection *redis.Client
	command    RedisCommand
}

type RedisCommand interface {
	SetSessionData(key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	GetSessionData(key string) (string, error)
}

//SessionHandlerInterface is the interface for the SessionHandler. It is an interface
//so that it can be mocked for unit tests.
type SessionHandlerInterface interface {
	ValidateSession() error
	InitCache() error
	EncodeSessionData() (string, error)
	RegenerateID() error
	SetupExpiration() error
	SetSession(encodedData string) error
	ValidateCookieSignature(req *http.Request, cookieSignature string) error
	GetStoredSession(req *http.Request) (string, error)
	ExtractAndValidateCookieSignatureParts(req *http.Request, cookieSignature string)
	DecodeSession(req *http.Request, session string) (map[string]interface{}, error)
	Clear(req *http.Request)
	ValidateExpiration(req *http.Request) error
	GenerateSignature() string
}

//Multiples of 3 bytes avoids = padding in base64 string
//7 * 3 bytes = (21/3) * 4 = 28 base64 characters
const idOctets = 7 * 3
const signatureStart = (idOctets * 4) / 3
const signatureLength = 27 //160 bits, base 64 encoded
const cookieValueLength = signatureStart + signatureLength

const defaultExpirationEnv = "DEFAULT_EXPIRATION"
const cookieNameEnv = "COOKIE_NAME"
const cookieSecretEnv = "COOKIE_SECRET"

/*
   STORE
*/

//NewStore will properly initialise a new Store object.
func NewStore() *Store {
	s := &Store{Encoder: encoding.New(), Cache: &Cache{}}
	s.InitSessionHandler()

	return s
}

//Load is used to try and get a session from the cache. If it succeeds it will
//load the session, otherwise it will return an error.
func (s *Store) Load(req *http.Request) error {

	cookie := s.getCookieFromRequest(req)

	err := s.SessionHandler.ValidateCookieSignature(req, cookie.Value)

	if err != nil {
		return err
	}

	s.SessionHandler.ExtractAndValidateCookieSignatureParts(req, cookie.Value)

	storedSession, err := s.SessionHandler.GetStoredSession(req)

	if err != nil {
		return err
	}

	s.Data, err = s.SessionHandler.DecodeSession(req, storedSession)

	if err != nil {
		return err
	}

	//Create a new session if the data is nil
	if s.Data == nil {
		s.SessionHandler.Clear(req)
		return nil
	}

	err = s.SessionHandler.ValidateExpiration(req)
	if err != nil {
		return err
	}

	return nil
}

// Store will take a store struct, validate it, and attempt to save it
func (s *Store) Store() error {

	log.Info("Attempting to store session with the following data: ", s.Data)

	if err := s.SessionHandler.ValidateSession(); err != nil {
		log.Error(err)
		return err
	}

	err := s.SessionHandler.InitCache()
	if err != nil {
		log.Error(err)
		return err
	}

	encodedData, err := s.SessionHandler.EncodeSessionData()
	if err != nil {
		log.Error(err)
		return err
	}

	if err := s.SessionHandler.SetSession(encodedData); err != nil {
		log.Error(err)
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

	err := s.InitCache()
	if err != nil {
		log.InfoR(req, err.Error())
	}

	_, err = s.Cache.connection.Del(sessionID).Result()

	if err != nil {
		log.InfoR(req, err.Error())
	}
}

//Clear destroys the current loaded session and removes it from the backing
//store. It will also regenerate the session ID.
func (s *Store) Clear(req *http.Request) {
	s.Data = nil
	s.Delete(req, nil) //Delete the previously stored Session
	s.RegenerateID()
}

// RegenerateID refreshes the token against the Store struct
func (s *Store) RegenerateID() error {
	octets := make([]byte, idOctets)

	if _, err := rand.Read(octets); err != nil {
		return err
	}

	s.ID = s.Encoder.EncodeBase64(octets)
	return nil
}

func (s *Store) GenerateSignature() string {
	sum := s.Encoder.GenerateSha1Sum([]byte(s.ID + cookieSecretEnv))
	return s.Encoder.EncodeBase64(sum[:])
}

// SetupExpiration will set the 'Expires' variable against the Store
// This should only be called if an expiration is not already set
func (s *Store) SetupExpiration() error {

	now := uint64(time.Now().Unix())

	expirationPeriod := s.Expiration

	if expirationPeriod == 0 {
		expirationPeriod, err := strconv.ParseUint(os.Getenv(defaultExpirationEnv), 0, 64)
		if err != nil {
			log.Info(err.Error())
			return err
		} else {
			log.Info("Setting expiration period on session ID: " + s.ID + " to " +
				string(expirationPeriod))
		}
	}

	s.Expires = now + expirationPeriod

	if s.Data != nil {
		s.Data["last_access"] = now
	}

	return nil
}

// ValidateSession will be called to authenticate the session store
func (s *Store) ValidateSession() error {

	if len(s.ID) == 0 {
		if err := s.SessionHandler.RegenerateID(); err != nil {
			return err
		}
	}

	if s.Expires == 0 {
		if err := s.SessionHandler.SetupExpiration(); err != nil {
			return err
		}
	}

	if s.Data == nil {
		return errors.New("No session data to store")
	}

	return nil
}

//getCookieFromRequest will attempt to pull the Cookie from the request. If err
//is not nil, it will create a new Cookie and return that instead.
func (s *Store) getCookieFromRequest(req *http.Request) *http.Cookie {

	var cookie *http.Cookie
	var err error

	cookieName := os.Getenv(cookieNameEnv)

	if cookie, err = req.Cookie(cookieName); err != nil {
		log.InfoR(req, err.Error())
		cookie = &http.Cookie{}
	}

	return cookie
}

//ValidateCookieSignature will try to validate that the length of the Cookie
//value is less than the calculated length of the signature
func (s *Store) ValidateCookieSignature(req *http.Request, cookieSignature string) error {

	if len(cookieSignature) < cookieValueLength {
		err := errors.New("Cookie signature is less than the desired cookie length")
		log.InfoR(req, err.Error())

		s.SessionHandler.Clear(req)

		return err
	}

	return nil
}

//ExtractAndValidateCookieSignatureParts will split the cookieSignature into
//two parts, and set the first part to s.ID, with the second part being validated
//against a generated ID.
func (s *Store) ExtractAndValidateCookieSignatureParts(req *http.Request, cookieSignature string) {
	s.ID = cookieSignature[0:signatureStart]
	sig := cookieSignature[signatureStart:len(cookieSignature)]

	//Validate signature is the same
	if sig != s.SessionHandler.GenerateSignature() {
		s.SessionHandler.Clear(req)
		return
	}
}

//init will construct a new Cache and invoke setRedisClient.
func (s *Store) InitCache() error {
	cache := &Cache{}

	var redisCommand RedisCommand = cache
	cache.command = redisCommand

	if err := cache.setRedisClient(); err != nil {
		return err
	}

	s.Cache = cache

	return nil
}

//GetStoredSession will get the session from the Cache
func (s *Store) GetStoredSession(req *http.Request) (string, error) {

	storedSession, err := s.Cache.command.GetSessionData(s.ID)

	if err != nil {
		log.InfoR(req, err.Error())
		return "", err
	}

	return storedSession, nil
}

//DecodeSession will try to base64 decode the session and then msgpack decode it.
func (s *Store) DecodeSession(req *http.Request, session string) (map[string]interface{}, error) {
	base64DecodedSession, err := s.Encoder.DecodeBase64(session)

	if err != nil {
		log.InfoR(req, err.Error())
		return nil, err
	}

	msgpackDecodedSession, err := s.Encoder.DecodeMsgPack(base64DecodedSession)

	if err != nil {
		log.InfoR(req, err.Error())
		return nil, err
	}

	return msgpackDecodedSession, nil
}

func (s *Store) ValidateExpiration(req *http.Request) error {
	s.Expiration = s.Data["expiration"].(uint64)
	s.Expires = s.Data["expires"].(uint64)

	if s.Expires == uint64(0) {
		s.SetupExpiration()
	}

	now := uint64(time.Now().Unix())

	if s.Expires <= now {
		err := errors.New("Store has expired")
		s.Data = nil

		return err
	}

	s.Expires = 0

	return nil
}

/*
   CACHE
*/

func (c *Cache) SetSessionData(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return c.connection.Set(key, value, expiration)
}

func (c *Cache) GetSessionData(key string) (string, error) {
	return c.connection.Get(key).Result()
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

// SetSession will take the valid Store object and save it in Redis
func (s *Store) SetSession(encodedData string) error {

	var err error
	_, err = s.Cache.command.SetSessionData(s.ID, encodedData, 0).Result()
	return err
}

// EncodeSessionData performs the messagepack and base 64 encoding on the
// session data and returns the result, or an error if one occurs
func (s *Store) EncodeSessionData() (string, error) {

	msgpackEncodedData, err := s.Encoder.EncodeMsgPack(s.Data)
	if err != nil {
		return "", err
	}

	b64EncodedData := s.Encoder.EncodeBase64(msgpackEncodedData)
	return b64EncodedData, nil
}

func (s *Store) InitSessionHandler() {
	var sessionHandler SessionHandlerInterface = s
	s.SessionHandler = sessionHandler
}
