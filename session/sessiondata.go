package session

import (
	"strconv"
	"time"

	"github.com/companieshouse/chs.go/log"

	"github.com/companieshouse/go-session-handler/config"
	goauth2 "golang.org/x/oauth2"
)

// Session is a map respresentation of the session data
type Session map[string]interface{}

// GetAccessToken retrieves the access token from the session data
func (data *Session) GetAccessToken() string {
	signinInfo := (*data)["signin_info"].(map[string]interface{})
	accessTokenMap := (signinInfo)["access_token"].(map[string]interface{})
	return (accessTokenMap)["access_token"].(string)
}

// getRefreshToken retrieves the refresh token from the session data
func (data *Session) getRefreshToken() string {
	signinInfo := (*data)["signin_info"].(map[string]interface{})
	accessTokenMap := (signinInfo)["access_token"].(map[string]interface{})
	return (accessTokenMap)["refresh_token"].(string)
}

// getExpiry retrieves the 'expires' value from the session data and converts it
// to a time
func (data *Session) getExpiry() time.Time {
	expiry := (*data)["expires"].(uint32)
	return time.Unix(int64(expiry), 0)
}

// isSignedIn checks whether a user is signed in given the session data. Returns
// a boolean
func (data *Session) isSignedIn() bool {
	signinInfo, ok := (*data)["signin_info"].(map[string]interface{})
	if !ok {
		return false
	}
	signedInFlag := signinInfo["signed_in"]
	signedIn, ok := signedInFlag.(int8)
	return ok && signedIn == 1
}

// SetAccessToken sets the access token on the session data map
func (data *Session) SetAccessToken(accessToken string) {
	signinInfo := (*data)["signin_info"].(map[string]interface{})
	accessTokenMap := signinInfo["access_token"].(map[string]interface{})
	accessTokenMap["access_token"] = accessToken
}

// SetRefreshToken sets the refresh token on the session data map
func (data *Session) SetRefreshToken(refreshToken string) {
	signinInfo := (*data)["signin_info"].(map[string]interface{})
	accessTokenMap := signinInfo["access_token"].(map[string]interface{})
	accessTokenMap["refresh_token"] = refreshToken
}

// GetExpiration returns the expiration period from the session data
func (data *Session) GetExpiration() uint64 {
	signinInfo, ok := (*data)["signin_info"].(map[string]interface{})
	if !ok {
		log.Info("GetExpiration(): 'signin_info' not found - returning expiration of '0'")
		return uint64(0)
	}
	accessTokenMap, ok := (signinInfo)["access_token"].(map[string]interface{})
	if !ok {
		log.Info("GetExpiration(): 'access_token' not found - returning expiration of '0'")
		return uint64(0)
	}
	expiration, ok := (accessTokenMap)["expires_in"].(uint16)
	if !ok {
		log.Info("GetExpiration(): 'expires_in' not found - returning expiration of '0'")
		return uint64(0)
	}
	return uint64(expiration)
}

// RefreshExpiration updates the 'expires' value on the session to the current
// time plus the expiration period
func (data *Session) RefreshExpiration() error {
	var err error
	expiration := data.GetExpiration()
	if expiration == uint64(0) {
		expiration, err = strconv.ParseUint(config.Get().DefaultExpiration, 0, 64)
		if err != nil {
			return err
		}
	}

	(*data)["expires"] = uint32(uint64(time.Now().Unix()) + expiration)
	return nil
}

// GetOauth2Token returns an oauth2 token derived from the session data. Returns
// nil if the user is not yet signed in
func (data *Session) GetOauth2Token() *goauth2.Token {
	if data.isSignedIn() {
		tok := &goauth2.Token{AccessToken: data.GetAccessToken(),
			RefreshToken: data.getRefreshToken(),
			Expiry:       data.getExpiry(),
		}

		return tok
	}

	return nil
}
