/*
Package state contains the go implementation for storing and loading the Session
from the cache.
*/

package state

import (
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

}

func (s *Store) Store() {

}

func (s *Store) regenerateID() {

}

func (s *Store) generateSignature() {

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
	}
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
