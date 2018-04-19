package session

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// TestGetAccessToken verifies that the session data access token is returned
// correctly
func TestGetAccessToken(t *testing.T) {

	Convey("Given I have session data with an access token", t, func() {

		accessToken := "Foo"

		var sessionData SessionData

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

// TestGetRefreshToken verifies that the session data refresh token is returned
// correctly
func TestGetRefreshToken(t *testing.T) {

	Convey("Given I have session data with a refresh token", t, func() {

		refreshToken := "Bar"

		var sessionData SessionData

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

// TestSetAccessToken verifies that the session data access token is stored
// correctly
func TestSetAccessToken(t *testing.T) {

	Convey("Given I have session data with an old access token", t, func() {

		oldAccessToken := "Foo"

		var sessionData SessionData

		session := map[string]interface{}{
			"signin_info": map[string]interface{}{
				"access_token": map[string]interface{}{
					"access_token": oldAccessToken,
				},
			},
		}
		sessionData = session

		Convey("When I call SetAccessToken with a new token", func() {

			newAccessToken := "Bar"
			sessionData.SetAccessToken(newAccessToken)

			Convey("Then the access token should be updated", func() {

				So(sessionData.GetAccessToken(), ShouldEqual, newAccessToken)
			})
		})
	})
}

// TestGetOauth2TokenUserSignedIn verifies that an oauth2 token is returned when
// a user is signed in
func TestGetOauth2TokenUserSignedIn(t *testing.T) {

	Convey("Given I have session data for a signed-in session", t, func() {
		accessToken := "Foo"
		refreshToken := "Bar"
		expiry := uint32(12345)

		var sessionData SessionData = map[string]interface{}{
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

// TestGetOauth2TokenUserNotSignedIn verifies that nothing is returned when
// a user is signed in
func TestGetOauth2TokenNotUserSignedIn(t *testing.T) {

	Convey("Given I have session data for a non-signed-in session", t, func() {

		var sessionData SessionData = map[string]interface{}{
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
