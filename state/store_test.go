package state

import (
	"errors"
	"os"
	"testing"
	"time"

	mockState "github.com/companieshouse/go-session-handler/state/state_mocks"
	. "github.com/smartystreets/goconvey/convey"

	redis "gopkg.in/redis.v5"
)

func setEnvVariables() {
	m := map[string]string{
		"ID_OCTETS":          "28",
		"DEFAULT_EXPIRATION": "60",
	}

	for key, value := range m {
		os.Setenv(key, value)
	}
}

func clearEnvVariables() {
	os.Clearenv()
}

// ------------------- Routes Through SetSession() -------------------

// TestSetSessionErrorOnSave - Verify error trapping if any errors are returned
// from redis
func TestSetSessionErrorOnSave(t *testing.T) {

	Convey("Given a Redis error is thrown when saving session data", t, func() {

		connection := &mockState.Connection{}
		connection.On("Set", "", "", time.Duration(0)).
			Return(redis.NewStatusResult("", errors.New("Unsuccessful save")))

		Convey("When I initialise the Store and try to save it", func() {

			cache := &Cache{connection: connection}

			s := &Store{cache: cache}

			err := s.setSession("")

			Convey("Then I expect the error to be caught and returned", func() {

				So(err, ShouldNotBeNil)
				So("Unsuccessful save", ShouldEqual, err.Error())
			})
		})
	})
}

// ------------------- Routes Through setupExpiration() -------------------

// TestSetupExpirationDefaultPeriodEnvVarMissing - Verify an error is thrown if
// the 'DEFAULT_EXPIRATION' env var is not set
func TestSetupExpirationDefaultPeriodEnvVarMissing(t *testing.T) {

	Convey("Given I haven't set environement variables and I initialise a new store with no expiration",
		t, func() {

			s := NewStore(nil)
			So(s.Expiration, ShouldEqual, 0)

			Convey("When I set up the expiration", func() {

				err := s.setupExpiration()

				Convey("Then I expect an error to be caught and returned", func() {

					So(err, ShouldNotBeNil)
				})
			})
		})
}

// ------------------- Routes Through validateSession() -------------------

// TestValidateSessionDataIsNil - Verify error trapping if when validating
// session data, there's no data to store
func TestValidateSessionDataIsNil(t *testing.T) {

	setEnvVariables()

	Convey("Given I initialise a store without any data", t, func() {

		s := NewStore(nil)

		Convey("When I validate the store", func() {

			err := s.validateSession()

			Convey("Then I expect an error to be caught and returned", func() {

				So(err, ShouldNotBeNil)
				So("No session data to store", ShouldEqual, err.Error())
			})
		})
	})

	clearEnvVariables()
}

// TestValidateSessionErrorInSetupExpiration - Verify error trapping if there's
// a problem setting the expiration on the store
func TestValidateSessionErrorInSetupExpiration(t *testing.T) {

	Convey("Given I haven't set any environment variables and I initialise a store",
		t, func() {

			s := NewStore(nil)

			Convey("When I validate the store", func() {

				err := s.validateSession()

				Convey("Then I expect an error to be caught and returned", func() {

					So(err, ShouldNotBeNil)
					So("strconv.ParseUint: parsing \"\": invalid syntax", ShouldEqual, err.Error())
				})
			})
		})
}

// TestValidateSessionHappyPath - Verify no errors are returned from ValidateSession
// if the happy path is followed
func TestValidateSessionHappyPath(t *testing.T) {

	setEnvVariables()

	Convey("Given I initialise a store with data", t, func() {

		s := NewStore(nil)
		s.Data = map[string]interface{}{
			"test": "hello, world!",
		}

		Convey("When I validate the store", func() {

			err := s.validateSession()

			Convey("Then I expect no errors to be returned", func() {

				So(err, ShouldBeNil)
			})
		})
	})

	clearEnvVariables()
}
