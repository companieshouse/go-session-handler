package state

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	encoding "github.com/companieshouse/go-session-handler/encoding"
	mockEncoding "github.com/companieshouse/go-session-handler/encoding/encoding_mocks"
	mockState "github.com/companieshouse/go-session-handler/state/state_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	redis "gopkg.in/redis.v5"
)

func (s *Store) setStoreData() {

	data := map[string]interface{}{
		"test": "Hello, world!",
	}

	s.Data = data
}

func setEnvVariables(variablesToOmit []string) {
	m := map[string]string{
		"ID_OCTETS":          "28",
		"DEFAULT_EXPIRATION": "60",
	}

	if len(variablesToOmit) > 0 {
		for i := 0; i < len(variablesToOmit); i++ {
			delete(m, variablesToOmit[i])
		}
	}

	for key, value := range m {
		os.Setenv(key, value)
	}
}

func clearEnvVariables() {
	os.Clearenv()
}

func getMockStoreObjects() (*mockEncoding.EncodingInterface, *mockState.SessionHandlerInterface, *Cache) {
	sessionHandler := &mockState.SessionHandlerInterface{}
	encoder := &mockEncoding.EncodingInterface{}
	connectionInfo := &redis.Options{}
	command := &mockState.RedisCommand{}

	command.On("SetRedisClient", connectionInfo).Return(nil)

	cache, _ := NewCache(connectionInfo, command)

	return encoder, sessionHandler, cache
}

// ----------------------------------------------------------------------------

// TestValidateStoreDataIsNil - Verify an error is thrown if session data is nil
func TestValidateStoreDataIsNil(t *testing.T) {

	setEnvVariables([]string{})
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()

	sessionHandler.On("RegenerateID").Return(nil)
	sessionHandler.On("SetupExpiration").Return(nil)

	s := NewStore(encoder, sessionHandler, cache)

	err := s.ValidateSession()
	assert.Equal("No session data to store", err.Error())

	clearEnvVariables()
}

// TestValidateStoreHappyPath - Verify no errors are returned from
// ValidateSession if the happy path is followed
func TestValidateStoreHappyPath(t *testing.T) {

	setEnvVariables([]string{})
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	sessionHandler.On("RegenerateID").Return(nil)
	sessionHandler.On("SetupExpiration").Return(nil)

	s := NewStore(encoder, sessionHandler, cache)
	s.setStoreData()

	err := s.ValidateSession()
	assert.Nil(err)

	clearEnvVariables()
}

// TestValidateStoreErrorRegeneratingID - Verify error trapping is enforced if
// there's an error regenerating an ID
func TestValidateStoreErrorRegeneratingID(t *testing.T) {

	setEnvVariables([]string{})
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	sessionHandler.On("RegenerateID").Return(errors.New("Error Regenerating ID"))

	s := NewStore(encoder, sessionHandler, cache)

	err := s.ValidateSession()
	assert.NotNil(err)

	clearEnvVariables()
}

// TestValidateStoreErrorSettingExpiration - Verify error trapping is enforced
// if there's an error setting expiration on the store
func TestValidateStoreErrorSettingExpiration(t *testing.T) {

	setEnvVariables([]string{})
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	sessionHandler.On("RegenerateID").Return(nil)
	sessionHandler.On("SetupExpiration").Return(errors.New("Error setting expiration"))

	s := NewStore(encoder, sessionHandler, cache)

	err := s.ValidateSession()
	assert.NotNil(err)

	clearEnvVariables()
}

// ------------------- Routes Through setupExpiration() -------------------

// TestSetupExpirationDefaultPeriodEnvVarMissing - Verify an error is thrown if
// the 'DEFAULT_EXPIRATION' env var is not set
func TestSetupExpirationDefaultPeriodEnvVarMissing(t *testing.T) {

	assert := assert.New(t)

	setEnvVariables([]string{defaultExpirationEnv})

	encoder, sessionHandler, cache := getMockStoreObjects()

	s := NewStore(encoder, sessionHandler, cache)

	err := s.SetupExpiration()
	assert.NotNil(err)

	clearEnvVariables()
}

// TestSetupExpirationDataIsNil - Verify 'Data' remains nil on setupExpiration
func TestSetupExpirationDataIsNil(t *testing.T) {

	assert := assert.New(t)

	setEnvVariables([]string{})
	encoder, sessionHandler, cache := getMockStoreObjects()

	s := NewStore(encoder, sessionHandler, cache)

	_ = s.SetupExpiration()
	assert.NotZero(s.Expires)

	// Session data remains nil
	assert.Nil(s.Data)
}

// TestSetupExpirationDataNotNil - Verify 'Data' is updated on setupExpiration
// to include a 'last_access' timestamp (seconds since epoch)
func TestSetupExpirationDataNotNil(t *testing.T) {

	assert := assert.New(t)

	setEnvVariables([]string{})
	encoder, sessionHandler, cache := getMockStoreObjects()

	s := NewStore(encoder, sessionHandler, cache)
	s.setStoreData()

	_ = s.SetupExpiration()
	assert.NotZero(s.Expires)
	assert.Contains(s.Data, "last_access")
}

// ------------------- Routes Through SetSession() -------------------

// TestSetSessionErrorOnSave - Verify error trapping if any errors are returned
// from redis
func TestSetSessionErrorOnSave(t *testing.T) {

	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()

	command := &mockState.RedisCommand{}
	command.On("SetSessionData", "", "", time.Duration(0)).
		Return(redis.NewStatusResult("", errors.New("Unsuccessful save")))

	cache.command = command

	s := NewStore(encoder, sessionHandler, cache)
	s.setStoreData()

	err := s.SetSession("")
	assert.NotNil(err)
}

// TestSetSessionSuccessfulSave - Verify happy path is followed if session is
// saved in Redis
func TestSetSessionSuccessfulSave(t *testing.T) {

	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()

	command := &mockState.RedisCommand{}
	command.On("SetSessionData", "", "", time.Duration(0)).
		Return(redis.NewStatusResult("Success", nil))

	cache.command = command

	s := NewStore(encoder, sessionHandler, cache)
	s.setStoreData()

	err := s.SetSession("")
	assert.Nil(err)
}

// ------------------- Routes Through encodeSessionData() -------------------

// TestEncodeSessionDataMessagePackError - Verify error trapping is enforced if
// there's an error when messagepack encoding
func TestEncodeSessionDataMessagePackError(t *testing.T) {
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	s := NewStore(encoder, sessionHandler, cache)
	s.setStoreData()

	encoder.On("EncodeMsgPack", s.Data).
		Return([]uint8{}, errors.New("Error encoding"))

	_, err := s.EncodeSessionData()

	assert.NotNil(err)
}

// TestEncodeSessionDataHappyPath - Verify no errors are thrown when following
// the 'happy path'
func TestEncodeSessionDataHappyPath(t *testing.T) {
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	s := NewStore(encoder, sessionHandler, cache)

	s.setStoreData()

	encoder.On("EncodeMsgPack", s.Data).Return([]uint8{}, nil)
	encoder.On("EncodeBase64", []uint8{}).Return("")

	_, err := s.EncodeSessionData()

	assert.Nil(err)
}

// ------------------- Routes Through decodeSession() -------------------

// TestDecodeSessionDataBaseError - Verify error trapping is enforced if
// there's an error when base64 decoding
func TestDecodeSessionDataBase64Error(t *testing.T) {
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	s := NewStore(encoder, sessionHandler, cache)

	encoder.On("DecodeBase64", "").Return([]byte{}, errors.New("Error base 64 decoding"))

	_, err := s.DecodeSession(new(http.Request), "")

	assert.NotNil(err)
}

// TestDecodeSessionDataMessagePackError - Verify error trapping is enforced if
// there's an error when messagepack decoding
func TestDecodeSessionDataMessagePackError(t *testing.T) {
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	s := NewStore(encoder, sessionHandler, cache)

	encoder.On("DecodeBase64", "").Return([]byte{}, nil)
	encoder.On("DecodeMsgPack", []byte{}).
		Return(map[string]interface{}{}, errors.New("Error encoding"))

	_, err := s.DecodeSession(new(http.Request), "")

	assert.NotNil(err)
}

// TestDecodeSessionDataHappyPath - Verify no errors are thrown when following
// the 'happy path'
func TestDecodeSessionDataHappyPath(t *testing.T) {
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	s := NewStore(encoder, sessionHandler, cache)

	encoder.On("DecodeBase64", "").Return([]byte{}, nil)
	encoder.On("DecodeMsgPack", []byte{}).
		Return(map[string]interface{}{}, nil)

	_, err := s.DecodeSession(new(http.Request), "")

	assert.Nil(err)
}

// ------------------- Routes Through Store() -------------------

// TestStoreErrorInValidateStore - Verify error trapping is enforced if
// there's an error when validating session data
func TestStoreErrorInValidateStore(t *testing.T) {
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	s := NewStore(encoder, sessionHandler, cache)

	sessionHandler.On("ValidateSession").Return(errors.New("Error validating session"))

	err := s.Store()

	assert.NotNil(err)
}

// TestStoreErrorInInitCache - Verify error trapping is enforced if
// there's an error when initiating a cache
func TestStoreErrorInInitCache(t *testing.T) {
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	sessionHandler.On("ValidateSession").Return(nil)
	sessionHandler.On("EncodeSessionData").Return("", errors.New(""))

	s := NewStore(encoder, sessionHandler, cache)
	s.setStoreData()

	s.sessionHandler = sessionHandler

	err := s.Store()

	assert.NotNil(err)
}

// TestStoreErrorInEncodeSessionData - Verify error trapping is enforced if
// there's an error when encoding session data
func TestStoreErrorInEncodeSessionData(t *testing.T) {
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	s := NewStore(encoder, sessionHandler, cache)
	s.setStoreData()

	sessionHandler.On("ValidateSession").Return(nil)
	sessionHandler.On("EncodeSessionData").Return("", errors.New("Error encoding session data"))

	err := s.Store()

	assert.NotNil(err)
}

// TestStoreErrorInSetSession - Verify error trapping is enforced if
// there's an error when setting session data
func TestStoreErrorInSetSession(t *testing.T) {
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	s := NewStore(encoder, sessionHandler, cache)
	s.setStoreData()

	sessionHandler.On("ValidateSession").Return(nil)
	sessionHandler.On("InitCache").Return(nil)
	sessionHandler.On("EncodeSessionData").Return("", nil)
	sessionHandler.On("SetSession", "").Return(errors.New("Error setting session"))

	err := s.Store()

	assert.NotNil(err)
}

// TestStoreHappyPath - Verify no errors are returned if when storing data the
// happy path is followed
func TestStoreHappyPath(t *testing.T) {
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	s := NewStore(encoder, sessionHandler, cache)
	s.setStoreData()

	sessionHandler.On("ValidateSession").Return(nil)
	sessionHandler.On("InitCache").Return(nil)
	sessionHandler.On("EncodeSessionData").Return("", nil)
	sessionHandler.On("SetSession", "").Return(nil)

	err := s.Store()

	assert.Nil(err)
}

// ------------------- Routes Through NewStore() -------------------

// TestNewStore - Verify that when initiating a new Store struct, each of the
// components are also initialised
func TestNewStore(t *testing.T) {
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()

	s := NewStore(encoder, sessionHandler, cache)

	assert.NotNil(s)
	assert.NotNil(s.encoder)
	assert.NotNil(s.sessionHandler)
	assert.NotNil(s.cache)
}

func TestNewCache(t *testing.T) {
	assert := assert.New(t)

	connectionInfo := &redis.Options{}

	command := &mockState.RedisCommand{}

	command.On("SetRedisClient", connectionInfo).Return(nil)

	cache, err := NewCache(connectionInfo, command)

	assert.Nil(err)
	assert.NotNil(cache)
}

// ------------------- Routes Through Load() -------------------

// TestLoadErrorInValidateCookieSignature - Verify that error trapping is enforced
// when validating cookie signature on load
func TestLoadErrorInValidateCookieSignature(t *testing.T) {
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	s := NewStore(encoder, sessionHandler, cache)

	sessionHandler.On("ValidateCookieSignature", new(http.Request), "").
		Return(errors.New("Error validating cookie signature"))

	err := s.Load(new(http.Request))

	assert.NotNil(err)
}

// TestLoadErrorInGetStoredSession - Verify that error trapping is enforced
// if there's an error when retrieving the stored session
func TestLoadErrorInGetStoredSession(t *testing.T) {
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	s := NewStore(encoder, sessionHandler, cache)

	sessionHandler.On("ValidateCookieSignature", new(http.Request), "").Return(nil)
	sessionHandler.On("ExtractAndValidateCookieSignatureParts", new(http.Request), "").Return()
	sessionHandler.On("GetStoredSession", new(http.Request)).Return("",
		errors.New("Error retrieving stored session"))

	err := s.Load(new(http.Request))

	assert.NotNil(err)
}

// TestLoadErrorInDecodeSession - Verify that error trapping is enforced
// if there's an error when decoding session data
func TestLoadErrorInDecodeSession(t *testing.T) {
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	s := NewStore(encoder, sessionHandler, cache)

	sessionHandler.On("ValidateCookieSignature", new(http.Request), "").Return(nil)
	sessionHandler.On("ExtractAndValidateCookieSignatureParts", new(http.Request), "").Return()
	sessionHandler.On("GetStoredSession", new(http.Request)).Return("", nil)
	sessionHandler.On("DecodeSession", new(http.Request), "").
		Return(nil, errors.New("Error decoding session"))

	err := s.Load(new(http.Request))

	assert.NotNil(err)
}

// TestLoadDecodedSessionIsNil - Verify that if decoded session data is nil,
// Clear is called on the store
func TestLoadDecodedSessionIsNil(t *testing.T) {
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	s := NewStore(encoder, sessionHandler, cache)

	sessionHandler.On("ValidateCookieSignature", new(http.Request), "").Return(nil)
	sessionHandler.On("ExtractAndValidateCookieSignatureParts", new(http.Request), "").Return()
	sessionHandler.On("GetStoredSession", new(http.Request)).Return("", nil)
	sessionHandler.On("DecodeSession", new(http.Request), "").Return(nil, nil)
	sessionHandler.On("Clear", new(http.Request)).Return()

	err := s.Load(new(http.Request))

	assert.Nil(err)
	sessionHandler.AssertCalled(t, "Clear", new(http.Request))
}

// TestLoadErrorInValidateExpiration - Verify that error trapping is enforced if
// there's an issue in ValidateExpiration
func TestLoadErrorInValidateExpiration(t *testing.T) {
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	s := NewStore(encoder, sessionHandler, cache)

	sessionHandler.On("ValidateCookieSignature", new(http.Request), "").Return(nil)
	sessionHandler.On("ExtractAndValidateCookieSignatureParts", new(http.Request), "").Return()
	sessionHandler.On("GetStoredSession", new(http.Request)).Return("", nil)
	sessionHandler.On("DecodeSession", new(http.Request), "").
		Return(map[string]interface{}{"Test": "Hello, World!"}, nil)
	sessionHandler.On("ValidateExpiration", new(http.Request)).
		Return(errors.New("Error validating expiration"))

	err := s.Load(new(http.Request))

	assert.NotNil(err)
}

// TestLoadHappyPath - Verify that no errors are returned if the Load
// 'happy path' is followed
func TestLoadHappyPath(t *testing.T) {
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	s := NewStore(encoder, sessionHandler, cache)

	sessionHandler.On("ValidateCookieSignature", new(http.Request), "").Return(nil)
	sessionHandler.On("ExtractAndValidateCookieSignatureParts", new(http.Request), "").Return()
	sessionHandler.On("GetStoredSession", new(http.Request)).Return("", nil)
	sessionHandler.On("DecodeSession", new(http.Request), "").
		Return(map[string]interface{}{"Test": "Hello, World!"}, nil)
	sessionHandler.On("ValidateExpiration", new(http.Request)).Return(nil)

	err := s.Load(new(http.Request))

	assert.Nil(err)
}

// ------------------- Routes Through GetStoredSession() -------------------

// TestGetStoredSessionRedisError - Verify that when retrieving a stored session,
// if there's a Redis error it's trapped and returned
func TestGetStoredSessionRedisError(t *testing.T) {
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()

	command := &mockState.RedisCommand{}
	command.On("GetSessionData", mock.AnythingOfType("string")).
		Return("", errors.New("Redis error thrown"))

	cache.command = command

	s := NewStore(encoder, sessionHandler, cache)

	session, err := s.GetStoredSession(new(http.Request))
	assert.NotNil(err)
	assert.Equal("", session)
}

// TestGetStoredSessionHappyPath - Verify that when retrieving a stored session,
// if the happy path is followed no errors are returned
func TestGetStoredSessionHappyPath(t *testing.T) {
	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()

	command := &mockState.RedisCommand{}
	command.On("GetSessionData", mock.AnythingOfType("string")).
		Return("Test", nil)

	cache.command = command

	s := NewStore(encoder, sessionHandler, cache)

	session, err := s.GetStoredSession(new(http.Request))
	assert.Nil(err)
	assert.Equal("Test", session)
}

// ------------------- Routes Through ValidateExpiration() -------------------

// TestValidateExpirationExpiresIsZero - Verify that when expires is zero,
// setupExpiration is called
func TestValidateExpirationExpiresIsZero(t *testing.T) {

	encoder, sessionHandler, cache := getMockStoreObjects()
	s := NewStore(encoder, sessionHandler, cache)

	sessionHandler.On("SetupExpiration").Return(nil)

	data := map[string]interface{}{"expires": uint64(0), "expiration": uint64(60)}
	s.Data = data

	s.ValidateExpiration(new(http.Request))
}

// TestValidateExpirationHasExpired - Verify that when a session has expired we
// throw an error
func TestValidateExpirationHasExpired(t *testing.T) {

	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	s := NewStore(encoder, sessionHandler, cache)

	now := uint64(time.Now().Unix())
	expires := now - uint64(60)

	data := map[string]interface{}{"expires": expires, "expiration": uint64(60)}
	s.Data = data

	err := s.ValidateExpiration(new(http.Request))

	assert.Equal("Store has expired", err.Error())
}

// TestValidateExpirationHappyPath - Verify that no errors are thrown when the
// 'happy path' is followed when validating session expiration
func TestValidateExpirationHappyPath(t *testing.T) {

	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	s := NewStore(encoder, sessionHandler, cache)

	now := uint64(time.Now().Unix())
	expires := now + uint64(60)

	data := map[string]interface{}{"expires": expires, "expiration": uint64(60)}

	s.Data = data

	err := s.ValidateExpiration(new(http.Request))

	assert.Nil(err)
	assert.Equal(s.Expires, uint64(0))
}

// ---------------- Routes Through ValidateCookieSignature() ----------------

// TestValidateCookieSignatureLengthInvalid - Verify that if the signature from
// the cookie is too short, an appropriate error is thrown
func TestValidateCookieSignatureLengthInvalid(t *testing.T) {

	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	s := NewStore(encoder, sessionHandler, cache)

	sessionHandler.On("Clear", new(http.Request)).Return()

	err := s.ValidateCookieSignature(new(http.Request), "")

	assert.Equal("Cookie signature is less than the desired cookie length", err.Error())
}

// TestValidateCookieSignatureHappyPath - Verify that no errors are thrown when
// following the validate cookie signature 'happy path'
func TestValidateCookieSignatureHappyPath(t *testing.T) {

	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()
	s := NewStore(encoder, sessionHandler, cache)

	err := s.ValidateCookieSignature(new(http.Request), strings.Repeat("a", cookieValueLength))

	assert.Nil(err)
}

// --------- Routes Through ExtractAndValidateCookieSignatureParts() ---------

// TestValidateCookieSignatureParts - Verify that if the cookie signature parts
// don't match, we clear the session data
func TestValidateCookieSignatureParts(t *testing.T) {

	encoder, sessionHandler, cache := getMockStoreObjects()
	s := NewStore(encoder, sessionHandler, cache)

	sessionHandler.On("GenerateSignature").Return("abc")
	sessionHandler.On("Clear", new(http.Request)).Return()

	s.ExtractAndValidateCookieSignatureParts(new(http.Request),
		strings.Repeat("a", signatureStart))
}

// -------------------- Routes Through Delete() --------------------

// TestDelete - Verify that errors are logged if there's a Redis connection
// error whilst deleting session data
func TestDelete(t *testing.T) {

	encoder, sessionHandler, cache := getMockStoreObjects()

	command := &mockState.RedisCommand{}
	command.On("DeleteSessionData", mock.AnythingOfType("string")).
		Return(errors.New("Error deleting session data"))

	cache.command = command

	s := NewStore(encoder, sessionHandler, cache)

	testID := "test"

	s.Delete(new(http.Request), &testID)
}

// -------------------- Routes Through RegenerateID() --------------------

// TestRegenerateId - Verify that when regerating an ID, the new ID is base 64
// encoded
func TestRegenerateId(t *testing.T) {

	assert := assert.New(t)

	encoder, sessionHandler, cache := getMockStoreObjects()

	s := NewStore(encoder, sessionHandler, cache)

	var Encoder encoding.Encoder

	encoder.On("EncodeBase64", mock.AnythingOfType("[]uint8")).
		Return(Encoder.EncodeBase64([]byte("test")))

	s.RegenerateID()

	_, err := Encoder.DecodeBase64(s.ID)

	assert.Nil(err)
}
