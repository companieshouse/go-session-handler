package session

import (
	"time"

	goauth2 "golang.org/x/oauth2"
)

type SessionData map[string]interface{}

func (data *SessionData) GetAccessToken() string {
	signinInfo := (*data)["signin_info"].(map[string]interface{})
	accessTokenMap := (signinInfo)["access_token"].(map[string]interface{})
	return (accessTokenMap)["access_token"].(string)
}

func (data *SessionData) getRefreshToken() string {
	signinInfo := (*data)["signin_info"].(map[string]interface{})
	accessTokenMap := (signinInfo)["access_token"].(map[string]interface{})
	return (accessTokenMap)["refresh_token"].(string)
}

func (data *SessionData) getExpiry() time.Time {
	expiry := (*data)["expires"].(uint32)
	return time.Unix(int64(expiry), 0)
}

func (data *SessionData) isSignedIn() bool {
	signinInfo := (*data)["signin_info"].(map[string]interface{})
	signedInFlag := signinInfo["signed_in"]
	signedIn, ok := signedInFlag.(int8)
	return ok && signedIn == 1
}

func (data *SessionData) SetAccessToken(accessToken string) {
	signinInfo := (*data)["signin_info"].(map[string]interface{})
	accessTokenMap := signinInfo["access_token"].(map[string]interface{})
	accessTokenMap["access_token"] = accessToken
}

func (data *SessionData) GetExpiration() uint64 {
	signinInfo := (*data)["signin_info"].(map[string]interface{})
	accessTokenMap := (signinInfo)["access_token"].(map[string]interface{})
	expiration, ok := (accessTokenMap)["expires_in"].(uint16)
	if !ok {
		return uint64(0)
	}
	return uint64(expiration)
}

func (s *SessionData) GetOauth2Token() *goauth2.Token {
	if s.isSignedIn() {
		tok := &goauth2.Token{AccessToken: s.GetAccessToken(),
			RefreshToken: s.getRefreshToken(),
			Expiry:       s.getExpiry(),
		}

		return tok
	}

	return nil
}
