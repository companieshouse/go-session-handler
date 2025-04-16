package state

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/companieshouse/go-session-handler/config"
	"github.com/companieshouse/go-session-handler/encoding"
	mockState "github.com/companieshouse/go-session-handler/state/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"

	redis "gopkg.in/redis.v5"
)

func getConfig() *config.Config {
	return &config.Config{
		DefaultExpiration: "60",
		CookieName:        "TEST",
		CookieSecret:      strings.Repeat("b", signatureLength),
	}
}

func initConfig() {
	os.Setenv("COOKIE_SECRET", "hello")
	os.Setenv("COOKIE_NAME", "TEST")
}

func cleanupConfig() {
	os.Unsetenv("COOKIE_SECRET")
	os.Unsetenv("COOKIE_NAME")
}

// ------------------- Routes Through SetSession() -------------------

// TestUnitSetSessionErrorOnSave - Verify error trapping if any errors are returned
// from redis
func TestUnitSetSessionErrorOnSave(t *testing.T) {

	Convey("Given a Redis error is thrown when saving session data", t, func() {

		connection := &mockState.Connection{}
		connection.On("Set", "", "", time.Duration(0)).
			Return(redis.NewStatusResult("", errors.New("Unsuccessful save")))

		Convey("When I initialise the Store and try to save it", func() {

			cache := &Cache{connection: connection}

			s := &Store{cache: cache}

			err := s.storeSession("")

			Convey("Then I expect the error to be caught and returned", func() {

				So(err, ShouldNotBeNil)
				So("Unsuccessful save", ShouldEqual, err.Error())
			})
		})
	})
}

// ------------------- Routes Through GetSession() -------------------

// TestUnitGetSessionErrorPath - Verify error trapping if any errors are returned
// from redis
func TestUnitGetSessionErrorPath(t *testing.T) {

	Convey("Given a Redis error is thrown when retrieving session data", t, func() {

		dummySessionData := "foo"

		connection := &mockState.Connection{}
		connection.On("Get", mock.AnythingOfType("string")).
			Return(redis.NewStringResult(dummySessionData, errors.New("Unsuccessful session retrieval")))

		Convey("When I initialise the Store and try to get the session", func() {

			cache := &Cache{connection: connection}

			s := &Store{cache: cache}

			session, err := s.fetchSession()

			Convey("Then I expect the error to be caught and returned, and session data should be blank",
				func() {

					So(err, ShouldNotBeNil)
					So("Unsuccessful session retrieval", ShouldEqual, err.Error())
					So(session, ShouldBeBlank)
				})
		})
	})
}

// TestUnitGetSessionHappyPath - Verify no errors are returned when following the
// GetSession 'happy path'
func TestUnitGetSessionHappyPath(t *testing.T) {

	Convey("Given no errors are thrown when retrieving session data", t, func() {

		dummySessionData := "foo"

		connection := &mockState.Connection{}
		connection.On("Get", mock.AnythingOfType("string")).
			Return(redis.NewStringResult(dummySessionData, nil))

		Convey("When I initialise the Store and try to get the session", func() {

			cache := &Cache{connection: connection}

			s := &Store{cache: cache}

			session, err := s.fetchSession()

			Convey("Then I expect the session to be returned, and no errors",
				func() {

					So(err, ShouldBeNil)
					So(session, ShouldEqual, dummySessionData)
				})
		})
	})
}

// ------------------- Routes Through Store() -------------------

// TestUnitStoreErrorInValidateSession - Verify session data is cleared if there's an
// issue when validating the session data
func TestUnitStoreErrorInValidateSession(t *testing.T) {

	initConfig()

	Convey("Given I create a store with no data", t, func() {

		s := NewStore(nil)

		Convey("When I store the session", func() {

			err := s.Store()

			Convey("Then I expect no errors but an empty session map", func() {

				So(err, ShouldBeNil)
				So(len(s.Data), ShouldEqual, 0)
			})
		})
	})

	cleanupConfig()
}

// TestUnitStoreErrorInSetSession - Verify error trapping is enforced if there's an
// issue when saving the session data
func TestUnitStoreErrorInSetSession(t *testing.T) {

	initConfig()

	Convey("Given I create a store with valid data but there's an error when saving the session",
		t, func() {

			connection := &mockState.Connection{}
			connection.On("Set", mock.AnythingOfType("string"),
				mock.AnythingOfType("string"), time.Duration(0)).
				Return(redis.NewStatusResult("", errors.New("Error saving session data")))

			c := &Cache{connection: connection}

			data := map[string]interface{}{
				"test": "hello, world!",
				"signin_info": map[string]interface{}{
					"access_token": map[string]interface{}{
						"expires_in": uint16(123),
					},
				},
			}

			s := NewStore(c)
			s.Data = data

			Convey("When I store the session", func() {

				err := s.Store()

				Convey("Then I expect the error to be caught and returned", func() {

					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "Error saving session data")
				})
			})
		})

	cleanupConfig()
}

// TestUnitStoreHappyPath - Verify no errors are returned from Store if the happy
// path is followed
func TestUnitStoreHappyPath(t *testing.T) {

	initConfig()

	Convey("Given I create a store with valid data and I follow the 'happy path'",
		t, func() {

			connection := &mockState.Connection{}
			connection.On("Set", mock.AnythingOfType("string"),
				mock.AnythingOfType("string"), time.Duration(0)).
				Return(redis.NewStatusResult("", nil))

			c := &Cache{connection: connection}

			data := map[string]interface{}{
				"test": "hello, world!",
				"signin_info": map[string]interface{}{
					"access_token": map[string]interface{}{
						"expires_in": uint16(123),
					},
				},
			}

			s := NewStore(c)
			s.Data = data

			Convey("When I store the session", func() {

				err := s.Store()

				Convey("Then I expect no errors to be returned", func() {

					So(err, ShouldBeNil)
				})
			})
		})

	cleanupConfig()
}

// ------------------- Routes Through validateExpiration() -------------------

// TestUnitValidateExpirationSessionHasExpired - Verify that when a session has
// expired we throw an error
func TestUnitValidateExpirationSessionHasExpired(t *testing.T) {

	initConfig()

	Convey("Given I have an expired session", t, func() {

		s := NewStore(nil)

		now := uint64(time.Now().Unix())
		expires := uint32(now - uint64(60))

		data := map[string]interface{}{"expires": expires, "expiration": uint64(60)}
		s.Data = data

		Convey("Given I call validate expiration on the store", func() {

			err := s.validateExpiration()

			Convey("Then an appropriate error is returned", func() {

				So(err.Error(), ShouldEqual, "Store has expired")
			})
		})
	})

	cleanupConfig()
}

// TestUnitValidateExpirationNoExpirationSet - Verify that when 'expires' isn't set
// on the store, it is set in validateExpiration
func TestUnitValidateExpirationNoExpirationSet(t *testing.T) {

	initConfig()

	Convey("Given I have an session store with expires set to 0", t, func() {

		s := NewStore(nil)

		data := map[string]interface{}{
			"expires": uint32(0),
			"signin_info": map[string]interface{}{
				"access_token": map[string]interface{}{
					"expires_in": uint16(123),
				},
			},
		}

		s.Data = data

		Convey("Given I call validate expiration on the store", func() {

			err := s.validateExpiration()

			Convey("Then no errors are returned and expires has been set", func() {

				So(err, ShouldBeNil)
				So(s.Expires, ShouldNotEqual, uint64(0))
			})
		})
	})

	cleanupConfig()
}

// TestUnitValidateExpirationExpirationNil - Verify that when 'expires' is nil
// on the store, it is set in validateExpiration
func TestUnitValidateExpirationExpirationNil(t *testing.T) {

	initConfig()

	Convey("Given I have an session store with expires set to 0", t, func() {

		s := NewStore(nil)

		data := map[string]interface{}{
			"expires": nil,
			"signin_info": map[string]interface{}{
				"access_token": map[string]interface{}{
					"expires_in": uint16(123),
				},
			},
		}

		s.Data = data

		Convey("Given I call validate expiration on the store", func() {

			err := s.validateExpiration()

			Convey("Then no errors are returned and expires has been set", func() {

				So(err, ShouldBeNil)
				So(s.Expires, ShouldNotEqual, uint64(0))
			})
		})
	})

	cleanupConfig()
}

// ------------------- Routes Through Delete() -------------------

// TestUnitDeleteErrorPath - Verify error trapping is enforced if there's an
// issue when deleting session data
func TestUnitDeleteErrorPath(t *testing.T) {

	initConfig()

	Convey("Given a Redis error is thrown when deleting session data", t, func() {

		connection := &mockState.Connection{}
		connection.On("Del", "abc").
			Return(redis.NewIntResult(0, errors.New("Unsuccessful Delete")))

		Convey("When I initialise the Store and try to delete it", func() {

			cache := &Cache{connection: connection}

			s := NewStore(cache)

			test := "abc"

			err := s.Delete(&test)

			Convey("Then the error should be caught and returned", func() {

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "Unsuccessful Delete")
			})
		})
	})

	cleanupConfig()
}

// TestUnitDeleteHappyPath - Verify no errors are returned when following the 'happy
// path' whilst deleting session data
func TestUnitDeleteHappyPath(t *testing.T) {

	initConfig()

	Convey("Given a the happy path is followed when deleting session data", t, func() {

		connection := &mockState.Connection{}
		connection.On("Del", "abc").
			Return(redis.NewIntResult(0, nil))

		Convey("When I initialise the Store and try to delete it", func() {

			cache := &Cache{connection: connection}

			s := NewStore(cache)

			test := "abc"

			err := s.Delete(&test)

			Convey("No errors should be returned", func() {

				So(err, ShouldBeNil)
			})
		})
	})

	cleanupConfig()
}

// ------------------- Routes Through Clear() -------------------

// TestUnitClearErrorPath - Verify error trapping is enforced if there's an
// issue when clearing session data
func TestUnitClearErrorPath(t *testing.T) {

	initConfig()

	Convey("Given a Redis error is thrown when deleting session data", t, func() {

		connection := &mockState.Connection{}
		connection.On("Del", "abc").
			Return(redis.NewIntResult(0, errors.New("Unsuccessful Delete")))

		Convey("When I initialise the Store and try to clear it", func() {

			cache := &Cache{connection: connection}

			s := NewStore(cache)

			s.ID = "abc"

			err := s.Clear()

			Convey("Then the error should be caught and returned and ID should remain unchanged", func() {

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "Unsuccessful Delete")
				So(s.ID, ShouldEqual, "abc")
			})
		})
	})

	cleanupConfig()
}

// TestUnitClearHappyPath - Verify no errors are returned from the Clear() happy path
func TestUnitClearHappyPath(t *testing.T) {

	initConfig()

	Convey("Given no errors are thrown when deleting session data", t, func() {

		connection := &mockState.Connection{}
		connection.On("Del", "abc").Return(redis.NewIntResult(0, nil))

		Convey("When I initialise the Store and try to clear it", func() {

			cache := &Cache{connection: connection}

			s := NewStore(cache)

			s.ID = "abc"
			s.Data = map[string]interface{}{
				"test": "Hello, world!",
			}

			err := s.Clear()

			Convey("Then no error should be returned, data should be empty, and the token should be refreshed",
				func() {

					So(err, ShouldBeNil)
					So(s.ID, ShouldNotEqual, "abc")
					So(len(s.Data), ShouldEqual, 0)
				})
		})
	})

	cleanupConfig()
}

// ---------------- Routes Through ValidateCookieSignature() ----------------

// TestUnitValidateCookieSignatureLengthInvalid - Verify that if the signature from
// the cookie is too short, an appropriate error is thrown
func TestUnitValidateCookieSignatureLengthInvalid(t *testing.T) {

	initConfig()

	Convey("Given the cookie signature is less than the desired length", t, func() {

		sig := strings.Repeat("a", cookieValueLength-1)

		Convey("When I initialise the Store and try to validate it, provided there are no Redis errors", func() {

			connection := &mockState.Connection{}
			connection.On("Del", "").Return(redis.NewIntResult(0, nil))

			c := &Cache{connection: connection}

			s := NewStore(c)
			err := s.validateSessionID(sig)

			Convey("Then an approriate error should be returned", func() {

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "Cookie signature is less than the desired cookie length")
			})
		})
	})

	cleanupConfig()
}

// TestUnitValidateSessionIDHappyPath - Verify that no errors are thrown when
// following the validate session ID 'happy path'
func TestUnitValidateSessionIDHappyPath(t *testing.T) {

	initConfig()

	Convey("Given the session ID is valid", t, func() {

		id := strings.Repeat("a", signatureStart)
		signatureByte := encoding.GenerateSha1Sum([]byte(id + "hello"))
		signature := encoding.EncodeBase64(signatureByte[:])

		sessionID := id + signature[0:signatureLength]

		Convey("When I initialise the Store and try to validate it", func() {

			s := NewStore(nil)
			err := s.validateSessionID(sessionID)

			Convey("Then no errors should be returned", func() {

				So(err, ShouldBeNil)
			})
		})
	})

	cleanupConfig()
}

// ---------------- Routes Through decodeSession() ----------------

// TestUnitDecodeSessionBase64Invalid - Verify that if a cookie doesn't exist by
// the name the config specifies, a new blank cookie is returned
func TestUnitDecodeSessionBase64Invalid(t *testing.T) {

	initConfig()

	Convey("Given the session string isn't base64 encoded", t, func() {

		s := NewStore(nil)

		Convey("When the Store tries to decode it", func() {
			decodedSession, err := s.decodeSession("Hello")

			Convey("Then I should have a blank decoded session", func() {
				So(decodedSession, ShouldBeNil)

				Convey("And the error should be populated", func() {
					So(err, ShouldNotBeNil)
				})
			})

		})
	})

	cleanupConfig()
}

// TestUnitDecodeSessionMessagepackInvalid - Verify that if a cookie doesn't exist by
// the name the config specifies, a new blank cookie is returned
func TestUnitDecodeSessionMessagepackInvalid(t *testing.T) {

	initConfig()

	Convey("Given the session string isn't messagepack encoded", t, func() {

		s := NewStore(nil)

		Convey("When the Store tries to decode it", func() {

			decodedSession, err := s.decodeSession("SGVsbG8=")

			Convey("Then I should have a blank decoded session", func() {

				So(decodedSession, ShouldBeNil)

				Convey("And the error should be populated", func() {

					So(err, ShouldNotBeNil)
				})
			})
		})
	})

	cleanupConfig()
}

// ---------------- Routes Through Load() ----------------

// TestUnitLoadErrorInValidateSignature - Verify error trapping whilst validating a
// cookie signature
func TestUnitLoadErrorInValidateSignature(t *testing.T) {

	initConfig()

	Convey("Given I have a session ID less than the desired length", t, func() {

		sessionID := strings.Repeat("a", cookieValueLength-1)

		Convey("And Redis throws no further errors", func() {

			connection := &mockState.Connection{}
			connection.On("Del", "").Return(redis.NewIntResult(0, nil))

			cache := &Cache{connection: connection}

			Convey("When I attempt to load the session", func() {

				s := NewStore(cache)

				err := s.Load(sessionID)

				Convey("Then no errors need to be returned, but the session data should be empty", func() {

					So(err, ShouldBeNil)
					So(len(s.Data), ShouldEqual, 0)
				})
			})
		})
	})

	cleanupConfig()
}

// TestUnitLoadErrorRetrievingSession - Verify error trapping whilst retrieving session
// data from Redis
func TestUnitLoadErrorRetrievingSession(t *testing.T) {

	initConfig()

	Convey("Given I have a valid session ID", t, func() {

		id := strings.Repeat("a", signatureStart)

		signatureByte := encoding.GenerateSha1Sum([]byte(id + "hello"))
		signature := encoding.EncodeBase64(signatureByte[:])

		sessionID := id + signature[0:signatureLength]

		Convey("If Redis returns an error", func() {

			connection := &mockState.Connection{}
			connection.On("Get", id).Return(redis.NewStringResult("",
				errors.New("Error retrieving session data")))

			cache := &Cache{connection: connection}

			Convey("When I attempt to load the session", func() {

				s := NewStore(cache)

				err := s.Load(sessionID)

				Convey("Then an error should be thrown whilst decoding the session", func() {

					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "Error retrieving session data")
				})
			})
		})
	})

	cleanupConfig()
}

// TestUnitLoadErrorDecodingSession - Verify error trapping whilst decoding session
// data on load
func TestUnitLoadErrorDecodingSession(t *testing.T) {

	initConfig()

	Convey("Given I have a valid session ID", t, func() {

		id := strings.Repeat("a", signatureStart)

		signatureByte := encoding.GenerateSha1Sum([]byte(id + "hello"))
		signature := encoding.EncodeBase64(signatureByte[:])

		sessionID := id + signature[0:signatureLength]

		Convey("If Redis returns blank data", func() {

			connection := &mockState.Connection{}
			connection.On("Get", id).Return(redis.NewStringResult("", nil))

			cache := &Cache{connection: connection}

			Convey("When I attempt to load the session", func() {

				s := NewStore(cache)

				err := s.Load(sessionID)

				Convey("Then an error should be thrown whilst decoding the session", func() {

					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "EOF")
				})
			})
		})
	})

	cleanupConfig()
}
