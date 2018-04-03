package state

import (
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	mockEncoding "github.com/companieshouse/go-session-handler/encoding/encoding_mocks"
	mockState "github.com/companieshouse/go-session-handler/state/state_mocks"
	"github.com/stretchr/testify/assert"

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

// ----------------------------------------------------------------------------

// TestValidateStoreDataIsNil - Verify an error is thrown if session data is nil
func TestValidateStoreDataIsNil(t *testing.T) {

	setEnvVariables([]string{})
	assert := assert.New(t)

	sessionHandler := &mockState.SessionHandlerInterface{}
	sessionHandler.On("RegenerateID").Return(nil)
	sessionHandler.On("SetupExpiration").Return(nil)

	s := &Store{SessionHandler: sessionHandler}

	err := s.ValidateSession()
	assert.Equal("No session data to store", err.Error())

	clearEnvVariables()
}

// TestValidateStoreHappyPath - Verify no errors are returned from
// ValidateSession if the happy path is followed
func TestValidateStoreHappyPath(t *testing.T) {

	setEnvVariables([]string{})
	assert := assert.New(t)

	sessionHandler := &mockState.SessionHandlerInterface{}
	sessionHandler.On("RegenerateID").Return(nil)
	sessionHandler.On("SetupExpiration").Return(nil)

	s := &Store{SessionHandler: sessionHandler}
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

	sessionHandler := &mockState.SessionHandlerInterface{}
	sessionHandler.On("RegenerateID").Return(errors.New("Error Regenerating ID"))

	s := &Store{SessionHandler: sessionHandler}

	err := s.ValidateSession()
	assert.NotNil(err)

	clearEnvVariables()
}

// TestValidateStoreErrorSettingExpiration - Verify error trapping is enforced
// if there's an error setting expiration on the store
func TestValidateStoreErrorSettingExpiration(t *testing.T) {

	setEnvVariables([]string{})
	assert := assert.New(t)

	sessionHandler := &mockState.SessionHandlerInterface{}
	sessionHandler.On("RegenerateID").Return(nil)
	sessionHandler.On("SetupExpiration").Return(errors.New("Error setting expiration"))

	s := &Store{SessionHandler: sessionHandler}

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

	s := &Store{}

	err := s.SetupExpiration()
	assert.NotNil(err)

	clearEnvVariables()
}

// TestSetupExpirationDataIsNil - Verify 'Data' remains nil on setupExpiration
func TestSetupExpirationDataIsNil(t *testing.T) {

	assert := assert.New(t)

	setEnvVariables([]string{})
	s := &Store{}

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
	s := &Store{}
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

	s := &Store{}
	s.setStoreData()

	c := &Cache{}

	command := &mockState.RedisCommand{}
	command.On("SetSessionData", "", "", time.Duration(0)).
		Return(redis.NewStatusResult("", errors.New("Unsuccessful save")))

	c.command = command

	s.Cache = c

	err := s.SetSession("")
	assert.NotNil(err)
}

// TestSetSessionSuccessfulSave - Verify happy path is followed if session is
// saved in Redis
func TestSetSessionSuccessfulSave(t *testing.T) {

	assert := assert.New(t)

	s := &Store{}
	s.setStoreData()

	c := &Cache{}

	command := &mockState.RedisCommand{}
	command.On("SetSessionData", "", "", time.Duration(0)).
		Return(redis.NewStatusResult("Success", nil))

	c.command = command

	s.Cache = c

	err := s.SetSession("")
	assert.Nil(err)
}

// ------------------- Routes Through encodeSessionData() -------------------

// TestEncodeSessionDataMessagePackError - Verify error trapping is enforced if
// there's an error when messagepack encoding
func TestEncodeSessionDataMessagePackError(t *testing.T) {
	assert := assert.New(t)

	s := &Store{}
	s.setStoreData()

	encodingInterface := &mockEncoding.EncodingInterface{}
	encodingInterface.On("EncodeMsgPack", s.Data).
		Return([]uint8{}, errors.New("Error encoding"))

	s.Encoder = encodingInterface

	_, err := s.EncodeSessionData()

	assert.NotNil(err)
}

// TestEncodeSessionDataHappyPath - Verify no errors are thrown when following
// the 'happy path'
func TestEncodeSessionDataHappyPath(t *testing.T) {
	assert := assert.New(t)

	s := &Store{}
	s.setStoreData()

	encodingInterface := &mockEncoding.EncodingInterface{}
	encodingInterface.On("EncodeMsgPack", s.Data).Return([]uint8{}, nil)
	encodingInterface.On("EncodeBase64", []uint8{}).Return("")

	s.Encoder = encodingInterface

	_, err := s.EncodeSessionData()

	assert.Nil(err)
}

// ------------------- Routes Through decodeSession() -------------------

// TestDecodeSessionDataBaseError - Verify error trapping is enforced if
// there's an error when base64 decoding
func TestDecodeSessionDataBase64Error(t *testing.T) {
	assert := assert.New(t)

	s := &Store{}

	encodingInterface := &mockEncoding.EncodingInterface{}
	encodingInterface.On("DecodeBase64", "").Return([]byte{}, errors.New("Error base 64 decoding"))

	s.Encoder = encodingInterface

	_, err := s.DecodeSession(new(http.Request), "")

	assert.NotNil(err)
}

// TestDecodeSessionDataMessagePackError - Verify error trapping is enforced if
// there's an error when messagepack decoding
func TestDecodeSessionDataMessagePackError(t *testing.T) {
	assert := assert.New(t)

	s := &Store{}

	encodingInterface := &mockEncoding.EncodingInterface{}
	encodingInterface.On("DecodeBase64", "").Return([]byte{}, nil)
	encodingInterface.On("DecodeMsgPack", []byte{}).
		Return(map[string]interface{}{}, errors.New("Error encoding"))

	s.Encoder = encodingInterface

	_, err := s.DecodeSession(new(http.Request), "")

	assert.NotNil(err)
}

// TestDecodeSessionDataHappyPath - Verify no errors are thrown when following
// the 'happy path'
func TestDecodeSessionDataHappyPath(t *testing.T) {
	assert := assert.New(t)

	s := &Store{}

	encodingInterface := &mockEncoding.EncodingInterface{}
	encodingInterface.On("DecodeBase64", "").Return([]byte{}, nil)
	encodingInterface.On("DecodeMsgPack", []byte{}).
		Return(map[string]interface{}{}, nil)

	s.Encoder = encodingInterface

	_, err := s.DecodeSession(new(http.Request), "")

	assert.Nil(err)
}

// ------------------- Routes Through Store() -------------------

// TestStoreErrorInValidateStore - Verify error trapping is enforced if
// there's an error when validating session data
func TestStoreErrorInValidateStore(t *testing.T) {
	assert := assert.New(t)

	s := &Store{}

	sessionHandler := &mockState.SessionHandlerInterface{}
	sessionHandler.On("ValidateSession").Return(errors.New("Error validating session"))

	s.SessionHandler = sessionHandler

	err := s.Store()

	assert.NotNil(err)
}

// TestStoreErrorInInitCache - Verify error trapping is enforced if
// there's an error when initiating a cache
func TestStoreErrorInInitCache(t *testing.T) {
	assert := assert.New(t)

	s := &Store{}
	s.setStoreData()

	sessionHandler := &mockState.SessionHandlerInterface{}
	sessionHandler.On("ValidateSession").Return(nil)
	sessionHandler.On("InitCache").Return(errors.New("Error initiating cache"))

	s.SessionHandler = sessionHandler

	err := s.Store()

	assert.NotNil(err)
}

// TestStoreErrorInEncodeSessionData - Verify error trapping is enforced if
// there's an error when encoding session data
func TestStoreErrorInEncodeSessionData(t *testing.T) {
	assert := assert.New(t)

	s := &Store{}
	s.setStoreData()

	sessionHandler := &mockState.SessionHandlerInterface{}
	sessionHandler.On("ValidateSession").Return(nil)
	sessionHandler.On("InitCache").Return(nil)
	sessionHandler.On("EncodeSessionData").Return("", errors.New("Error encoding session data"))

	s.SessionHandler = sessionHandler

	err := s.Store()

	assert.NotNil(err)
}

// TestStoreErrorInSetSession - Verify error trapping is enforced if
// there's an error when setting session data
func TestStoreErrorInSetSession(t *testing.T) {
	assert := assert.New(t)

	s := &Store{}
	s.setStoreData()

	sessionHandler := &mockState.SessionHandlerInterface{}
	sessionHandler.On("ValidateSession").Return(nil)
	sessionHandler.On("InitCache").Return(nil)
	sessionHandler.On("EncodeSessionData").Return("", nil)
	sessionHandler.On("SetSession", "").Return(errors.New("Error setting session"))

	s.SessionHandler = sessionHandler

	err := s.Store()

	assert.NotNil(err)
}

// TestStoreHappyPath - Verify no errors are returned if when storing data the
// happy path is followed
func TestStoreHappyPath(t *testing.T) {
	assert := assert.New(t)

	s := &Store{}
	s.setStoreData()

	sessionHandler := &mockState.SessionHandlerInterface{}
	sessionHandler.On("ValidateSession").Return(nil)
	sessionHandler.On("InitCache").Return(nil)
	sessionHandler.On("EncodeSessionData").Return("", nil)
	sessionHandler.On("SetSession", "").Return(nil)

	s.SessionHandler = sessionHandler

	err := s.Store()

	assert.Nil(err)
}

// ------------------- Routes Through NewStore() -------------------

// TestNewStore - Verify that when initiating a new Store struct, each of the
// components are also initialised
func TestNewStore(t *testing.T) {
	assert := assert.New(t)

	s := NewStore()

	assert.NotNil(s.Encoder)
	assert.NotNil(s.SessionHandler)
	assert.NotNil(s.Cache)
}

// ------------------- Routes Through Load() -------------------

// TestLoadErrorInValidateCookieSignature - Verify that error trapping is enforced
// when validating cookie signature on load
func TestLoadErrorInValidateCookieSignature(t *testing.T) {
	assert := assert.New(t)

	s := &Store{}

	sessionHandler := &mockState.SessionHandlerInterface{}
	sessionHandler.On("ValidateCookieSignature", new(http.Request), "").
		Return(errors.New("Error validating cookie signature"))

	s.SessionHandler = sessionHandler

	err := s.Load(new(http.Request))

	assert.NotNil(err)
}

// TestLoadErrorInGetStoredSession - Verify that error trapping is enforced
// if there's an error when retrieving the stored session
func TestLoadErrorInGetStoredSession(t *testing.T) {
	assert := assert.New(t)

	s := &Store{}

	sessionHandler := &mockState.SessionHandlerInterface{}
	sessionHandler.On("ValidateCookieSignature", new(http.Request), "").Return(nil)
	sessionHandler.On("ExtractAndValidateCookieSignatureParts", new(http.Request), "").Return()
	sessionHandler.On("GetStoredSession", new(http.Request)).Return("",
		errors.New("Error retrieving stored session"))

	s.SessionHandler = sessionHandler

	err := s.Load(new(http.Request))

	assert.NotNil(err)
}

// TestLoadErrorInDecodeSession - Verify that error trapping is enforced
// if there's an error when decoding session data
func TestLoadErrorInDecodeSession(t *testing.T) {
	assert := assert.New(t)

	s := &Store{}

	sessionHandler := &mockState.SessionHandlerInterface{}
	sessionHandler.On("ValidateCookieSignature", new(http.Request), "").Return(nil)
	sessionHandler.On("ExtractAndValidateCookieSignatureParts", new(http.Request), "").Return()
	sessionHandler.On("GetStoredSession", new(http.Request)).Return("", nil)
	sessionHandler.On("DecodeSession", new(http.Request), "").
		Return(nil, errors.New("Error decoding session"))

	s.SessionHandler = sessionHandler

	err := s.Load(new(http.Request))

	assert.NotNil(err)
}

// TestLoadDecodedSessionIsNil - Verify that if decoded session data is nil,
// Clear is called on the store
func TestLoadDecodedSessionIsNil(t *testing.T) {
	assert := assert.New(t)

	s := &Store{}

	sessionHandler := &mockState.SessionHandlerInterface{}
	sessionHandler.On("ValidateCookieSignature", new(http.Request), "").Return(nil)
	sessionHandler.On("ExtractAndValidateCookieSignatureParts", new(http.Request), "").Return()
	sessionHandler.On("GetStoredSession", new(http.Request)).Return("", nil)
	sessionHandler.On("DecodeSession", new(http.Request), "").Return(nil, nil)
	sessionHandler.On("Clear", new(http.Request)).Return()

	s.SessionHandler = sessionHandler

	err := s.Load(new(http.Request))

	assert.Nil(err)
	sessionHandler.AssertCalled(t, "Clear", new(http.Request))
}

// TestLoadErrorInValidateExpiration - Verify that error trapping is enforced if
// there's an issue in ValidateExpiration
func TestLoadErrorInValidateExpiration(t *testing.T) {
	assert := assert.New(t)

	s := &Store{}

	sessionHandler := &mockState.SessionHandlerInterface{}
	sessionHandler.On("ValidateCookieSignature", new(http.Request), "").Return(nil)
	sessionHandler.On("ExtractAndValidateCookieSignatureParts", new(http.Request), "").Return()
	sessionHandler.On("GetStoredSession", new(http.Request)).Return("", nil)
	sessionHandler.On("DecodeSession", new(http.Request), "").
		Return(map[string]interface{}{"Test": "Hello, World!"}, nil)
	sessionHandler.On("ValidateExpiration", new(http.Request)).
		Return(errors.New("Error validating expiration"))

	s.SessionHandler = sessionHandler

	err := s.Load(new(http.Request))

	assert.NotNil(err)
}

// TestLoadHappyPath - Verify that no errors are returned if the Load
// 'happy path' is followed
func TestLoadHappyPath(t *testing.T) {
	assert := assert.New(t)

	s := &Store{}

	sessionHandler := &mockState.SessionHandlerInterface{}
	sessionHandler.On("ValidateCookieSignature", new(http.Request), "").Return(nil)
	sessionHandler.On("ExtractAndValidateCookieSignatureParts", new(http.Request), "").Return()
	sessionHandler.On("GetStoredSession", new(http.Request)).Return("", nil)
	sessionHandler.On("DecodeSession", new(http.Request), "").
		Return(map[string]interface{}{"Test": "Hello, World!"}, nil)
	sessionHandler.On("ValidateExpiration", new(http.Request)).Return(nil)

	s.SessionHandler = sessionHandler

	err := s.Load(new(http.Request))

	assert.Nil(err)
}
