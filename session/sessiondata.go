package session

import (
	"os"
	"strconv"
	"time"

	goauth2 "golang.org/x/oauth2"
)

// SessionData is a map respresentation of the session data
type SessionData map[string]interface{}

// GetAccessToken retrieves the access token from the session data
func (data *SessionData) GetAccessToken() string {
	signinInfo := (*data)["signin_info"].(map[string]interface{})
	accessTokenMap := (signinInfo)["access_token"].(map[string]interface{})
	return (accessTokenMap)["access_token"].(string)
}

// getRefreshToken retrieves the refresh token from the session data
func (data *SessionData) getRefreshToken() string {
	signinInfo := (*data)["signin_info"].(map[string]interface{})
	accessTokenMap := (signinInfo)["access_token"].(map[string]interface{})
	return (accessTokenMap)["refresh_token"].(string)
}

// getExpiry retrieves the 'expires' value from the session data and converts it
// to a time
func (data *SessionData) getExpiry() time.Time {
	expiry := (*data)["expires"].(uint32)
	return time.Unix(int64(expiry), 0)
}

// isSignedIn checks whether a user is signed in given the session data. Returns
// a boolean
func (data *SessionData) isSignedIn() bool {
	signinInfo, ok := (*data)["signin_info"].(map[string]interface{})
	if !ok {
		return false
	}
	signedInFlag := signinInfo["signed_in"]
	signedIn, ok := signedInFlag.(int8)
	return ok && signedIn == 1
}

// SetAccessToken sets the access token on the session data map
func (data *SessionData) SetAccessToken(accessToken string) {
	signinInfo := (*data)["signin_info"].(map[string]interface{})
	accessTokenMap := signinInfo["access_token"].(map[string]interface{})
	accessTokenMap["access_token"] = accessToken
}

// GetExpiration returns the expiration period from the session data
func (data *SessionData) GetExpiration() uint64 {
	signinInfo := (*data)["signin_info"].(map[string]interface{})
	accessTokenMap := (signinInfo)["access_token"].(map[string]interface{})
	expiration, ok := (accessTokenMap)["expires_in"].(uint16)
	if !ok {
		return uint64(0)
	}
	return uint64(expiration)
}

// RefreshExpiration updates the 'expires' value on the session to the current
// time plus the expiration period
func (data *SessionData) RefreshExpiration() {
	expiration := data.GetExpiration()
	if expiration == uint64(0) {
		expiration, _ = strconv.ParseUint(os.Getenv("DEFAULT_EXPIRATION"), 0, 64)
	}

	(*data)["expires"] = uint32(uint64(time.Now().Unix()) + expiration)
}

// GetOauth2Token returns an oauth2 token derived from the session data. Returns
// nil if the user is not yet signed in
func (data *SessionData) GetOauth2Token() *goauth2.Token {
	if data.isSignedIn() {
		tok := &goauth2.Token{AccessToken: data.GetAccessToken(),
			RefreshToken: data.getRefreshToken(),
			Expiry:       data.getExpiry(),
		}

		return tok
	}

	return nil
}
