package state

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/companieshouse/go-session-handler/encoding"
	"github.com/companieshouse/go-session-handler/state/mocks"
	"github.com/stretchr/testify/assert"

	redis "gopkg.in/redis.v5"
)

func (s *Store) setStoreData() {

	data := map[string]interface{}{
		"test": "Hello, world!",
	}

	s.Data = data
}

func encodeData(jsonData map[string]interface{}) string {
	msgpackEncodedData, _ := encoding.EncodeMsgPack(jsonData)
	b64EncodedData := encoding.EncodeBase64(msgpackEncodedData)
	return b64EncodedData
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

func assertEnvVariableMissing(envVar string, err error, t *testing.T) {
	assert.Equal(t, "Environment variable not set: "+envVar, err.Error())
}

func clearEnvVariables() {
	os.Clearenv()
}

// ----------------------------------------------------------------------------

// TestValidateStoreDataIsNil - Verify an error is thrown if session data is nil
func TestValidateStoreDataIsNil(t *testing.T) {

	setEnvVariables([]string{})
	assert := assert.New(t)

	s := &Store{}
	err := s.validateSession()
	assert.Equal("No session data to store", err.Error())

	clearEnvVariables()
}

// ------------------- Routes Through regenerateID() -------------------

// TestRegenerateIDOctetsEnvVarMissing - Verify an error is thrown in the event
// that the 'ID_OCTETS' env var is not set
func TestRegenerateIDOctetsEnvVarMissing(t *testing.T) {

	setEnvVariables([]string{idOctetsStr})

	s := &Store{}

	err := s.regenerateID()
	assertEnvVariableMissing(idOctetsStr, err, t)

	clearEnvVariables()
}

// ------------------- Routes Through setupExpiration() -------------------

// TestSetupExpirationDefaultPeriodEnvVarMissing - Verify an error is thrown if
// the 'DEFAULT_EXPIRATION' env var is not set
func TestSetupExpirationDefaultPeriodEnvVarMissing(t *testing.T) {

	setEnvVariables([]string{defaultExpiration})

	s := &Store{}

	err := s.setupExpiration()
	assertEnvVariableMissing(defaultExpiration, err, t)

	clearEnvVariables()
}

// TestSetupExpirationDataIsNil - Verify 'Data' remains nil on setupExpiration
func TestSetupExpirationDataIsNil(t *testing.T) {

	assert := assert.New(t)

	setEnvVariables([]string{})
	s := &Store{}

	_ = s.setupExpiration()
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

	_ = s.setupExpiration()
	assert.NotZero(s.Expires)
	assert.Contains(s.Data, "last_access")
}

// ------------------- Routes Through setSession() -------------------

// TestSetSessionErrorOnSave - Verify error trapping if any errors are returned
// from redis
func TestSetSessionErrorOnSave(t *testing.T) {

	assert := assert.New(t)

	s := &Store{}
	s.setStoreData()

	encodedData := encodeData(s.Data)

	c := &Cache{}

	command := &mocks.RedisCommand{}
	command.On("Set", "", encodedData, time.Duration(0)).
		Return(redis.NewStatusResult("", errors.New("Unsuccessful save")))

	c.command = command

	err := c.setSession(s)
	assert.NotNil(err)
}

// TestSetSessionSuccessfulSave - Verify happy path is followed if session is
// saved in Redis
func TestSetSessionSuccessfulSave(t *testing.T) {

	assert := assert.New(t)

	s := &Store{}
	s.setStoreData()

	encodedData := encodeData(s.Data)

	c := &Cache{}

	command := &mocks.RedisCommand{}
	command.On("Set", "", encodedData, time.Duration(0)).
		Return(redis.NewStatusResult("Success", nil))

	c.command = command

	err := c.setSession(s)
	assert.Nil(err)
}
