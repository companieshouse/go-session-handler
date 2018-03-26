/*
Package state contains the go implementation for storing and loading the Session
from the cache.
*/

package state

import (
	"errors"
	"net/http"

	"github.com/companieshouse/chs.go/log"

	redis "gopkg.in/redis.v5"
)

const cookieName = "__SID"

//Multiples of 3 bytes avoids = padding in base64 string
//7 * 3 bytes = (21/3) * 4 = 28 base64 characters
const idOctets = 7 * 3
const signatureStart = (idOctets * 4) / 3
const signatureLength = 27 //160 bits, base 64 encoded
const cookieValueLength = signatureStart + signatureLength

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

func (s *Store) Store() {

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

func (s *Store) regenerateID() {

}

func (s *Store) generateSignature() string {
	return ""
}

//getCookieFromRequest will attempt to pull the Cookie from the request. If err
//is not nil, it will create a new Cookie and return that instead.
func (s *Store) getCookieFromRequest(req *http.Request) (*http.Cookie, error) {

	var cookie *http.Cookie
	var err error

	if cookie, err = req.Cookie(cookieName); err != nil {
		log.InfoR(req, err.Error())
		cookie = &http.Cookie{}
	}

	return cookie, err
}

//validateCookieSignature will try to validate that the length of the Cookie
//value is not equal to the calculated length of the signature
func (s *Store) validateCookieSignature(req *http.Request, cookieSignature string) {

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

func (c *Cache) getSession(id string) []byte {
	var a []byte
	return a
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
