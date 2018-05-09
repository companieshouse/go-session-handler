package state

import (
	"crypto/rand"
	"errors"
	"strconv"
	"time"

	"github.com/companieshouse/go-session-handler/config"
	"github.com/companieshouse/go-session-handler/encoding"
	session "github.com/companieshouse/go-session-handler/session"
	"github.com/ian-kent/go-log/log"
	redis "gopkg.in/redis.v5"
)

//Multiples of 3 bytes avoids = padding in base64 string
//7 * 3 bytes = (21/3) * 4 = 28 base64 characters
const idOctets = 7 * 3
const signatureStart = (idOctets * 4) / 3
const signatureLength = 27 //160 bits, base 64 encoded
const cookieValueLength = signatureStart + signatureLength

//Store is the struct that is used to load/store the session.
type Store struct {
	ID      string
	Expires uint64
	Data    session.SessionData
	cache   *Cache
	config  *config.Config
}

//NewStore will properly initialise a new Store object.
func NewStore(cache *Cache, config *config.Config) *Store {

	return &Store{cache: cache, config: config}
}

//Load is used to try and get a session from the cache. If it succeeds it will
//load the session, otherwise it will return an error.
func (s *Store) Load(sessionID string) error {

	err := s.validateSessionID(sessionID)

	// If validateSessionID returns an error, we need to return an empty session
	// That said, no exceptions have occured so return a nil error
	if err != nil {
		log.Trace(err.Error())
		return nil
	}

	session, err := s.fetchSession()
	if err != nil {
		if err == redis.Nil {
			//If the session isn't stored in Redis, clear any data and return nil error
			s.clearSessionData()
			return nil
		}
		return err
	}

	s.Data, err = s.decodeSession(session)
	if err != nil {
		return err
	}

	// Create a new session if the data is nil (not sure how this is possible!)
	if s.Data == nil {
		s.clearSessionData()
		return nil
	}

	err = s.validateExpiration()
	if err != nil {
		// If the session has expired, clear the data and return nil
		s.clearSessionData()
		return nil
	}

	return nil
}

// Store will take a store struct, validate it, and attempt to save it
func (s *Store) Store() error {

	if s.Data == nil {
		s.clearSessionData() // Set session data to an empty map rather than nil

		// Since this should never happen, we'll add a log warning
		log.Warn("Session data was nil for ID " + s.ID)
		return nil
	}

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

	encodedData, err := s.encodeSessionData()
	if err != nil {
		return err
	}

	if err := s.storeSession(encodedData); err != nil {
		return err
	}

	return nil
}

//Delete will clear the requested session from the backing store. Note: Delete
//does not clear the loaded session. The Clear method will take care of that.
//If the string passed in is nil, it will delete the session with an id the same
//as that of s.ID
func (s *Store) Delete(id *string) error {
	sessionID := s.ID

	if id != nil && len(*id) > 0 {
		sessionID = *id
	}

	err := s.cache.deleteSessionData(sessionID)
	if err != nil {
		return err
	}

	return nil
}

//Clear destroys the current loaded session and removes it from the backing
//store. It will also regenerate the session ID.
func (s *Store) Clear() error {
	err := s.Delete(nil) //Delete the previously stored Session because we're going to regenerate the IDS
	if err != nil {
		return err
	}
	s.clearSessionData()
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

//GenerateSignature will generate a new signature based on the Store ID and
//the cookie secret.
func (s *Store) GenerateSignature() string {
	sum := encoding.GenerateSha1Sum([]byte(s.ID + s.config.CookieSecret))
	sig := encoding.EncodeBase64(sum[:])
	//Substring applied here to accomodate for base64 encoded padding of '='
	return sig[0:signatureLength]
}

//setupExpiration will set the 'Expires' variable against the Store
//This should only be called if an expiration is not already set
func (s *Store) setupExpiration() error {

	var err error

	now := uint64(time.Now().Unix())

	// First and foremost, we prioritise the expiration on session data
	expirationPeriod := s.Data.GetExpiration()

	if expirationPeriod == uint64(0) {
		// If that's zero, retrieve the default expiration from environment variables
		expirationPeriod, err = strconv.ParseUint(s.config.DefaultExpiration, 0, 64)
		if err != nil {
			return err
		}
	}

	s.Expires = now + expirationPeriod

	if s.Data != nil {
		s.Data["last_access"] = now
	}

	return nil
}

// validateSessionID will validate the session ID, ensuring it hasn't been
// manipulated
func (s *Store) validateSessionID(sessionID string) error {

	if len(sessionID) < cookieValueLength {
		s.clearSessionData()
		return errors.New("Cookie signature is less than the desired cookie length")
	}

	s.ID = sessionID[0:signatureStart]
	sig := sessionID[signatureStart:len(sessionID)]

	//Validate signature is the same
	if sig != s.GenerateSignature() {
		s.clearSessionData()
		return errors.New("Session signature does not match the expected value! " +
			"Have " + sig + ", but wanted " + s.GenerateSignature())
	}

	return nil
}

//fetchSession will get the session from the Cache
func (s *Store) fetchSession() (string, error) {

	storedSession, err := s.cache.getSessionData(s.ID)
	if err != nil {
		return "", err
	}

	return storedSession, nil
}

//decodeSession will try to base64 decode the session and then msgpack decode it.
func (s *Store) decodeSession(session string) (map[string]interface{}, error) {

	base64DecodedSession, err := encoding.DecodeBase64(session)
	if err != nil {
		return nil, err
	}

	msgpackDecodedSession, err := encoding.DecodeMsgPack(base64DecodedSession)
	if err != nil {
		return nil, err
	}

	return msgpackDecodedSession, nil
}

//validateExpiration validates that the Expires and Expiration values on the
//Store object are valid, and sets them if required.
func (s *Store) validateExpiration() error {

	s.Expires = uint64(s.Data["expires"].(uint32))

	if s.Expires == uint64(0) {
		s.setupExpiration()
	}

	now := uint64(time.Now().Unix())

	if s.Expires <= now {
		return errors.New("Store has expired")
	}

	return nil
}

//storeSession will take the valid Store object and save it in Redis
func (s *Store) storeSession(encodedData string) error {

	var err error
	_, err = s.cache.setSessionData(s.ID, encodedData).Result()
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

// clearSessionData will set the session data to an empty map
func (s *Store) clearSessionData() {
	s.Data = map[string]interface{}{}
}
