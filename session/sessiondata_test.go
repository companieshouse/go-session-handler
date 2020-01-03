package session

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func initConfig() {
	os.Setenv("DEFAULT_SESSION_EXPIRATION", "5")
}

func cleanupConfig() {
	os.Unsetenv("DEFAULT_SESSION_EXPIRATION")
}

// TestUnitGetAccessToken verifies that the session data access token is returned
// correctly
func TestUnitGetAccessToken(t *testing.T) {

	Convey("Given I have session data with an access token", t, func() {

		accessToken := "Foo"

		var sessionData Session

		session := map[string]interface{}{
			"signin_info": map[string]interface{}{
				"access_token": map[string]interface{}{
					"access_token": accessToken,
				},
			},
		}
		sessionData = session

		Convey("When I call GetAccessToken", func() {

			output := sessionData.GetAccessToken()

			Convey("Then the access token should be returned", func() {

				So(output, ShouldEqual, accessToken)
			})
		})
	})
}

// TestUnitGetRefreshToken verifies that the session data refresh token is returned
// correctly
func TestUnitGetRefreshToken(t *testing.T) {

	Convey("Given I have session data with a refresh token", t, func() {

		refreshToken := "Bar"

		var sessionData Session

		session := map[string]interface{}{
			"signin_info": map[string]interface{}{
				"access_token": map[string]interface{}{
					"refresh_token": refreshToken,
				},
			},
		}
		sessionData = session

		Convey("When I call getRefreshToken", func() {

			output := sessionData.getRefreshToken()

			Convey("Then the refresh token should be returned", func() {

				So(output, ShouldEqual, refreshToken)
			})
		})
	})
}

// TestUnitSetAccessToken verifies that the session data access token is stored
// correctly
func TestUnitSetAccessToken(t *testing.T) {

	Convey("Given I have session data with an old access token", t, func() {

		oldAccessToken := "Foo"

		var sessionData Session = map[string]interface{}{
			"signin_info": map[string]interface{}{
				"access_token": map[string]interface{}{
					"access_token": oldAccessToken,
				},
			},
		}

		Convey("When I call SetAccessToken with a new token", func() {

			newAccessToken := "Bar"
			sessionData.SetAccessToken(newAccessToken)

			Convey("Then the access token should be updated", func() {

				So(sessionData.GetAccessToken(), ShouldEqual, newAccessToken)
			})
		})
	})
}

// TestUnitSetRefreshToken verifies that the session data refresh token is stored
// correctly
func TestUnitSetRefreshToken(t *testing.T) {

	Convey("Given I have session data with an old refresh token", t, func() {

		oldRefreshToken := "Foo"

		var sessionData Session = map[string]interface{}{
			"signin_info": map[string]interface{}{
				"access_token": map[string]interface{}{
					"refresh_token": oldRefreshToken,
				},
			},
		}

		Convey("When I call SetRefreshToken with a new token", func() {

			newRefreshToken := "Bar"
			sessionData.SetRefreshToken(newRefreshToken)

			Convey("Then the refresh token should be updated", func() {

				So(sessionData.getRefreshToken(), ShouldEqual, newRefreshToken)
			})
		})
	})
}

// TestUnitGetOauth2TokenUserSignedIn verifies that an oauth2 token is returned when
// a user is signed in
func TestUnitGetOauth2TokenUserSignedIn(t *testing.T) {

	Convey("Given I have session data for a signed-in session", t, func() {
		accessToken := "Foo"
		refreshToken := "Bar"
		expiry := uint32(12345)

		var sessionData Session = map[string]interface{}{
			"expires": expiry,
			"signin_info": map[string]interface{}{
				"signed_in": int8(1),
				"access_token": map[string]interface{}{
					"access_token":  accessToken,
					"refresh_token": refreshToken,
				},
			},
		}

		Convey("When I call GetOauth2Token", func() {

			tok := sessionData.GetOauth2Token()

			Convey("Then an oauth2 token should be retuned", func() {

				So(tok, ShouldNotBeNil)

				Convey("With the correct values", func() {

					So(tok.AccessToken, ShouldEqual, accessToken)
					So(tok.RefreshToken, ShouldEqual, refreshToken)
					So(tok.Expiry, ShouldEqual, time.Unix(int64(expiry), 0))
				})
			})
		})
	})
}

// TestUnitGetOauth2TokenNotUserSignedIn verifies that nothing is returned when
// a user is signed in
func TestUnitGetOauth2TokenNotUserSignedIn(t *testing.T) {

	Convey("Given I have session data for a non-signed-in session", t, func() {

		var sessionData Session = map[string]interface{}{
			"signin_info": map[string]interface{}{
				"signed_in": int8(0),
			},
		}

		Convey("When I call GetOauth2Token", func() {

			tok := sessionData.GetOauth2Token()

			Convey("Then nothing should be retuned", func() {

				So(tok, ShouldBeNil)
			})
		})
	})
}

// TestUnitIsSignedInEmptySessionDataMap verifies that false is returned when
// checking if an empty session is signed in
func TestUnitIsSignedInEmptySessionDataMap(t *testing.T) {

	Convey("Given I have an empty session map", t, func() {

		var sessionData Session = map[string]interface{}{}

		Convey("When I call isSignedIn", func() {

			signedIn := sessionData.isSignedIn()

			Convey("Then I should return false", func() {

				So(signedIn, ShouldBeFalse)
			})
		})
	})
}

// TestUnitGetExpirationHappyPath verifies that expiration is returned successfully
func TestUnitGetExpirationHappyPath(t *testing.T) {

	Convey("Given I have some session data with an 'expires_in' token", t, func() {

		expiresIn := uint16(123)

		var sessionData Session = map[string]interface{}{
			"signin_info": map[string]interface{}{
				"access_token": map[string]interface{}{
					"expires_in": expiresIn,
				},
			},
		}

		Convey("When I call GetExpiration", func() {

			expiration := sessionData.GetExpiration()

			Convey("Then expiration should be returned", func() {

				So(expiration, ShouldEqual, uint64(expiresIn))
			})
		})
	})
}

// TestUnitGetExpirationNonePresent verifies that when expiration is not present on
// the session, 0 is returned
func TestUnitGetExpirationNonePresent(t *testing.T) {

	Convey("Given I have some session data with no 'expires_in' token", t, func() {

		var sessionData Session = map[string]interface{}{
			"signin_info": map[string]interface{}{
				"access_token": map[string]interface{}{},
			},
		}

		Convey("When I call GetExpiration", func() {

			expiration := sessionData.GetExpiration()

			Convey("Then 0 should be returned", func() {

				So(expiration, ShouldEqual, uint64(0))
			})
		})
	})
}

// TestUnitRefreshExpiration verifies that once refreshed, expiration is not nil
func TestUnitRefreshExpiration(t *testing.T) {
	initConfig()

	Convey("Given I have some session data", t, func() {

		var sessionData Session = map[string]interface{}{
			"signin_info": map[string]interface{}{
				"access_token": map[string]interface{}{},
			},
			"expires": 5,
		}

		Convey("When I call RefreshExpiration", func() {

			sessionData.RefreshExpiration()

			Convey("Then 'expires' should be set", func() {

				So(sessionData.getExpiry(), ShouldNotBeNil)
			})
		})
	})

	cleanupConfig()
}
