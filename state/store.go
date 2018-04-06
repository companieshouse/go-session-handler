package state

import (
	"crypto/rand"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/go-session-handler/encoding"
)

//Multiples of 3 bytes avoids = padding in base64 string
//7 * 3 bytes = (21/3) * 4 = 28 base64 characters
const idOctets = 7 * 3
const signatureStart = (idOctets * 4) / 3
const signatureLength = 27 //160 bits, base 64 encoded
const cookieValueLength = signatureStart + signatureLength

//Store is the struct that is used to load/store the session.
type Store struct {
	ID         string
	Expiration uint64
	Expires    uint64
	Data       map[string]interface{}
	cache      *Cache
	config     *StoreConfig
}

//StoreConfig holds the necessary config required for the Store object to be able
//to perform the various actions.
type StoreConfig struct {
	defaultExpiration string
	cookieName        string
	cookieSecret      string
}

//NewStore will properly initialise a new Store object.
func NewStore(cache *Cache, config *StoreConfig) *Store {

	return &Store{cache: cache, config: config}
}

func NewStoreConfig(defaultExpiration string, cookieName string, cookieSecret string) *StoreConfig {

	return &StoreConfig{
		defaultExpiration: defaultExpiration,
		cookieName:        cookieName,
		cookieSecret:      cookieSecret,
	}
}

//Load is used to try and get a session from the cache. If it succeeds it will
//load the session, otherwise it will return an error.
func (s *Store) Load(req *http.Request) error {

	cookie := s.getCookieFromRequest(req)

	err := s.validateCookieSignature(req, cookie.Value)

	if err != nil {
		return err
	}

	s.extractAndValidateCookieSignatureParts(req, cookie.Value)

	storedSession, err := s.getStoredSession(req)

	if err != nil {
		return err
	}

	s.Data, err = s.decodeSession(req, storedSession)

	if err != nil {
		return err
	}

	//Create a new session if the data is nil
	if s.Data == nil {
		s.Clear(req)
		return nil
	}

	err = s.validateExpiration(req)
	if err != nil {
		return err
	}

	return nil
}

// Store will take a store struct, validate it, and attempt to save it
func (s *Store) Store() error {

	log.Info("Attempting to store session with the following data: ", s.Data)

	if err := s.validateSession(); err != nil {
		log.Error(err)
		return err
	}

	encodedData, err := s.encodeSessionData()
	if err != nil {
		log.Error(err)
		return err
	}

	if err := s.setSession(encodedData); err != nil {
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
func (s *Store) Delete(req *http.Request, id *string) error {
	sessionID := s.ID

	if id != nil && len(*id) > 0 {
		sessionID = *id
	}

	err := s.cache.deleteSessionData(sessionID)

	if err != nil {
		log.InfoR(req, err.Error())
		return err
	}

	return nil
}

//Clear destroys the current loaded session and removes it from the backing
//store. It will also regenerate the session ID.
func (s *Store) Clear(req *http.Request) error {
	err := s.Delete(req, nil) //Delete the previously stored Session
	if err != nil {
		return err
	}
	s.Data = nil
	s.regenerateID()
	return nil
}

//regenerateID refreshes the token against the Store struct
func (s *Store) regenerateID() error {
	octets := make([]byte, idOctets)

	if _, err := rand.Read(octets); err != nil {
		return err
	}

	s.ID = encoding.EncodeBase64(octets)
	return nil
}

//generateSignature will generate a new signature based on the Store ID and
//the cookie secret.
func (s *Store) generateSignature() string {
	sum := encoding.GenerateSha1Sum([]byte(s.ID + s.config.cookieSecret))
	return encoding.EncodeBase64(sum[:])
}

//setupExpiration will set the 'Expires' variable against the Store
//This should only be called if an expiration is not already set
func (s *Store) setupExpiration() error {

	now := uint64(time.Now().Unix())

	expirationPeriod := s.Expiration
	var err error

	if expirationPeriod == 0 {
		expirationPeriod, err = strconv.ParseUint(s.config.defaultExpiration, 0, 64)

		if err != nil {
			log.Info(err.Error())
			return err
		}

		log.Info("Setting expiration period on session ID: " + s.ID + " to " +
			strconv.FormatUint(expirationPeriod, 10) + " seconds")
	}

	s.Expires = now + expirationPeriod

	if s.Data != nil {
		s.Data["last_access"] = now
	}

	return nil
}

//validateSession will be called to authenticate the session store
func (s *Store) validateSession() error {

	if len(s.ID) == 0 {
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
		return errors.New("No session data to store")
	}

	return nil
}

//getCookieFromRequest will attempt to pull the Cookie from the request. If err
//is not nil, it will create a new Cookie and return that instead.
func (s *Store) getCookieFromRequest(req *http.Request) *http.Cookie {

	var cookie *http.Cookie
	var err error

	if cookie, err = req.Cookie(s.config.cookieName); err != nil {
		log.InfoR(req, err.Error())
		cookie = &http.Cookie{}
	}

	return cookie
}

//validateCookieSignature will try to validate that the length of the Cookie
//value is less than the calculated length of the signature
func (s *Store) validateCookieSignature(req *http.Request, cookieSignature string) error {

	if len(cookieSignature) < cookieValueLength {
		err := errors.New("Cookie signature is less than the desired cookie length")
		log.InfoR(req, err.Error())

		s.Clear(req)

		return err
	}

	return nil
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

//getStoredSession will get the session from the Cache
func (s *Store) getStoredSession(req *http.Request) (string, error) {

	storedSession, err := s.cache.getSessionData(s.ID)

	if err != nil {
		log.InfoR(req, err.Error())
		return "", err
	}

	return storedSession, nil
}

//decodeSession will try to base64 decode the session and then msgpack decode it.
func (s *Store) decodeSession(req *http.Request, session string) (map[string]interface{}, error) {
	base64DecodedSession, err := encoding.DecodeBase64(session)

	if err != nil {
		log.InfoR(req, err.Error())
		return nil, err
	}

	msgpackDecodedSession, err := encoding.DecodeMsgPack(base64DecodedSession)

	if err != nil {
		log.InfoR(req, err.Error())
		return nil, err
	}

	return msgpackDecodedSession, nil
}

//validateExpiration validates that the Expires and Expiration values on the
//Store object are valid, and sets them if required.
func (s *Store) validateExpiration(req *http.Request) error {
	s.Expiration = s.Data["expiration"].(uint64)
	s.Expires = s.Data["expires"].(uint64)

	if s.Expires == uint64(0) {
		s.setupExpiration()
	}

	now := uint64(time.Now().Unix())

	if s.Expires <= now {
		err := errors.New("Store has expired")
		s.Data = nil

		return err
	}

	return nil
}

//setSession will take the valid Store object and save it in Redis
func (s *Store) setSession(encodedData string) error {

	var err error
	_, err = s.cache.setSessionData(s.ID, encodedData, 0).Result()
	return err
}

//encodeSessionData performs the messagepack and base 64 encoding on the
//session data and returns the result, or an error if one occurs
func (s *Store) encodeSessionData() (string, error) {

	msgpackEncodedData, err := encoding.EncodeMsgPack(s.Data)
	if err != nil {
		return "", err
	}

	b64EncodedData := encoding.EncodeBase64(msgpackEncodedData)
	return b64EncodedData, nil
}
