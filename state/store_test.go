package state

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/companieshouse/go-session-handler/encoding"
	mockState "github.com/companieshouse/go-session-handler/state/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"

	redis "gopkg.in/redis.v5"
)

func getStoreConfig() *StoreConfig {
	return &StoreConfig{
		DefaultExpiration: "60",
		CookieName:        "TEST",
		CookieSecret:      strings.Repeat("b", signatureLength),
	}
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

			err := s.storeSession("")

			Convey("Then I expect the error to be caught and returned", func() {

				So(err, ShouldNotBeNil)
				So("Unsuccessful save", ShouldEqual, err.Error())
			})
		})
	})
}

// ------------------- Routes Through GetSession() -------------------

// TestGetSessionErrorPath - Verify error trapping if any errors are returned
// from redis
func TestGetSessionErrorPath(t *testing.T) {

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

// TestGetSessionHappyPath - Verify no errors are returned when following the
// GetSession 'happy path'
func TestGetSessionHappyPath(t *testing.T) {

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

// TestStoreErrorInValidateSession - Verify session data is cleared if there's an
// issue when validating the session data
func TestStoreErrorInValidateSession(t *testing.T) {

	Convey("Given I create a store with no data", t, func() {

		s := NewStore(nil, getStoreConfig())

		Convey("When I store the session", func() {

			err := s.Store()

			Convey("Then I expect no errors but an empty session map", func() {

				So(err, ShouldBeNil)
				So(s.sessionDataIsEmpty(), ShouldBeTrue)
			})
		})
	})
}

// TestStoreErrorInSetSession - Verify error trapping is enforced if there's an
// issue when saving the session data
func TestStoreErrorInSetSession(t *testing.T) {

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

			s := NewStore(c, getStoreConfig())
			s.Data = data

			Convey("When I store the session", func() {

				err := s.Store()

				Convey("Then I expect the error to be caught and returned", func() {

					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "Error saving session data")
				})
			})
		})
}

// TestStoreHappyPath - Verify no errors are returned from Store if the happy
// path is followed
func TestStoreHappyPath(t *testing.T) {

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

			s := NewStore(c, getStoreConfig())
			s.Data = data

			Convey("When I store the session", func() {

				err := s.Store()

				Convey("Then I expect no errors to be returned", func() {

					So(err, ShouldBeNil)
				})
			})
		})
}

// ------------------- Routes Through validateExpiration() -------------------

// TestValidateExpirationSessionHasExpired - Verify that when a session has
// expired we throw an error
func TestValidateExpirationSessionHasExpired(t *testing.T) {

	Convey("Given I have an expired session", t, func() {

		s := NewStore(nil, getStoreConfig())

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
}

// TestValidateExpirationNoExpirationSet - Verify that when 'expires' isn't set
// on the store, it is set in validateExpiration
func TestValidateExpirationNoExpirationSet(t *testing.T) {

	Convey("Given I have an session store with expires set to 0", t, func() {

		s := NewStore(nil, getStoreConfig())

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
}

// ------------------- Routes Through Delete() -------------------

// TestDeleteErrorPath - Verify error trapping is enforced if there's an
// issue when deleting session data
func TestDeleteErrorPath(t *testing.T) {

	Convey("Given a Redis error is thrown when deleting session data", t, func() {

		connection := &mockState.Connection{}
		connection.On("Del", "abc").
			Return(redis.NewIntResult(0, errors.New("Unsuccessful Delete")))

		Convey("When I initialise the Store and try to delete it", func() {

			cache := &Cache{connection: connection}

			s := NewStore(cache, getStoreConfig())

			test := "abc"

			err := s.Delete(&test)

			Convey("Then the error should be caught and returned", func() {

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "Unsuccessful Delete")
			})
		})
	})
}

// TestDeleteHappyPath - Verify no errors are returned when following the 'happy
// path' whilst deleting session data
func TestDeleteHappyPath(t *testing.T) {

	Convey("Given a the happy path is followed when deleting session data", t, func() {

		connection := &mockState.Connection{}
		connection.On("Del", "abc").
			Return(redis.NewIntResult(0, nil))

		Convey("When I initialise the Store and try to delete it", func() {

			cache := &Cache{connection: connection}

			s := NewStore(cache, getStoreConfig())

			test := "abc"

			err := s.Delete(&test)

			Convey("No errors should be returned", func() {

				So(err, ShouldBeNil)
			})
		})
	})
}

// ------------------- Routes Through Clear() -------------------

// TestClearErrorPath - Verify error trapping is enforced if there's an
// issue when clearing session data
func TestClearErrorPath(t *testing.T) {

	Convey("Given a Redis error is thrown when deleting session data", t, func() {

		connection := &mockState.Connection{}
		connection.On("Del", "abc").
			Return(redis.NewIntResult(0, errors.New("Unsuccessful Delete")))

		Convey("When I initialise the Store and try to clear it", func() {

			cache := &Cache{connection: connection}

			s := NewStore(cache, getStoreConfig())

			s.ID = "abc"

			err := s.Clear()

			Convey("Then the error should be caught and returned and ID should remain unchanged", func() {

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "Unsuccessful Delete")
				So(s.ID, ShouldEqual, "abc")
			})
		})
	})
}

// TestClearHappyPath - Verify no errors are returned from the Clear() happy path
func TestClearHappyPath(t *testing.T) {

	Convey("Given no errors are thrown when deleting session data", t, func() {

		connection := &mockState.Connection{}
		connection.On("Del", "abc").Return(redis.NewIntResult(0, nil))

		Convey("When I initialise the Store and try to clear it", func() {

			cache := &Cache{connection: connection}

			s := NewStore(cache, getStoreConfig())

			s.ID = "abc"
			s.Data = map[string]interface{}{
				"test": "Hello, world!",
			}

			err := s.Clear()

			Convey("Then no error should be returned, data should be nil, and the token should be refreshed",
				func() {

					So(err, ShouldBeNil)
					So(s.ID, ShouldNotEqual, "abc")
					So(s.Data, ShouldBeNil)
				})
		})
	})
}

// ---------------- Routes Through ValidateCookieSignature() ----------------

// TestValidateCookieSignatureLengthInvalid - Verify that if the signature from
// the cookie is too short, an appropriate error is thrown
func TestValidateCookieSignatureLengthInvalid(t *testing.T) {

	Convey("Given the cookie signature is less than the desired length", t, func() {

		sig := strings.Repeat("a", cookieValueLength-1)

		Convey("When I initialise the Store and try to validate it, provided there are no Redis errors", func() {

			connection := &mockState.Connection{}
			connection.On("Del", "").Return(redis.NewIntResult(0, nil))

			c := &Cache{connection: connection}

			s := NewStore(c, getStoreConfig())
			err := s.validateSessionID(sig)

			Convey("Then an approriate error should be returned", func() {

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "Cookie signature is less than the desired cookie length")
			})
		})
	})
}

// TestValidateSessionIDHappyPath - Verify that no errors are thrown when
// following the validate session ID 'happy path'
func TestValidateSessionIDHappyPath(t *testing.T) {

	Convey("Given the session ID is valid", t, func() {

		config := getStoreConfig()

		id := strings.Repeat("a", signatureStart)
		signatureByte := encoding.GenerateSha1Sum([]byte(id + config.CookieSecret))
		signature := encoding.EncodeBase64(signatureByte[:])

		sessionID := id + signature[0:signatureLength]

		Convey("When I initialise the Store and try to validate it", func() {

			s := NewStore(nil, config)
			err := s.validateSessionID(sessionID)

			Convey("Then no errors should be returned", func() {

				So(err, ShouldBeNil)
			})
		})
	})
}

// ---------------- Routes Through decodeSession() ----------------

// TestDecodeSessionBase64Invalid - Verify that if a cookie doesn't exist by
// the name the config specifies, a new blank cookie is returned
func TestDecodeSessionBase64Invalid(t *testing.T) {

	Convey("Given the session string isn't base64 encoded", t, func() {

		s := NewStore(nil, getStoreConfig())

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
}

// TestDecodeSessionMessagepackInvalid - Verify that if a cookie doesn't exist by
// the name the config specifies, a new blank cookie is returned
func TestDecodeSessionMessagepackInvalid(t *testing.T) {

	Convey("Given the session string isn't messagepack encoded", t, func() {

		s := NewStore(nil, getStoreConfig())

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
}

// ---------------- Routes Through Load() ----------------

// TestLoadErrorInValidateSignature - Verify error trapping whilst validating a
// cookie signature
func TestLoadErrorInValidateSignature(t *testing.T) {

	Convey("Given I have a session ID less than the desired length", t, func() {

		sessionID := strings.Repeat("a", cookieValueLength-1)

		Convey("And Redis throws no further errors", func() {

			connection := &mockState.Connection{}
			connection.On("Del", "").Return(redis.NewIntResult(0, nil))

			cache := &Cache{connection: connection}

			Convey("When I attempt to load the session", func() {

				config := &StoreConfig{}

				s := NewStore(cache, config)

				err := s.Load(sessionID)

				Convey("Then no errors need to be returned, but the session data should be empty", func() {

					So(err, ShouldBeNil)
					So(s.sessionDataIsEmpty(), ShouldBeTrue)
				})
			})
		})
	})
}

// TestLoadErrorRetrievingSession - Verify error trapping whilst retrieving session
// data from Redis
func TestLoadErrorRetrievingSession(t *testing.T) {

	Convey("Given I have a valid session ID", t, func() {

		config := &StoreConfig{CookieSecret: strings.Repeat("b", signatureLength)}

		id := strings.Repeat("a", signatureStart)

		signatureByte := encoding.GenerateSha1Sum([]byte(id + config.CookieSecret))
		signature := encoding.EncodeBase64(signatureByte[:])

		sessionID := id + signature[0:signatureLength]

		Convey("If Redis returns an error", func() {

			connection := &mockState.Connection{}
			connection.On("Get", id).Return(redis.NewStringResult("",
				errors.New("Error retrieving session data")))

			cache := &Cache{connection: connection}

			Convey("When I attempt to load the session", func() {

				s := NewStore(cache, config)

				err := s.Load(sessionID)

				Convey("Then an error should be thrown whilst decoding the session", func() {

					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "Error retrieving session data")
				})
			})
		})
	})
}

// TestLoadErrorDecodingSession - Verify error trapping whilst decoding session
// data on load
func TestLoadErrorDecodingSession(t *testing.T) {

	Convey("Given I have a valid session ID", t, func() {

		config := &StoreConfig{CookieSecret: strings.Repeat("b", signatureLength)}

		id := strings.Repeat("a", signatureStart)

		signatureByte := encoding.GenerateSha1Sum([]byte(id + config.CookieSecret))
		signature := encoding.EncodeBase64(signatureByte[:])

		sessionID := id + signature[0:signatureLength]

		Convey("If Redis returns blank data", func() {

			connection := &mockState.Connection{}
			connection.On("Get", id).Return(redis.NewStringResult("", nil))

			cache := &Cache{connection: connection}

			Convey("When I attempt to load the session", func() {

				s := NewStore(cache, config)

				err := s.Load(sessionID)

				Convey("Then an error should be thrown whilst decoding the session", func() {

					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "EOF")
				})
			})
		})
	})
}
