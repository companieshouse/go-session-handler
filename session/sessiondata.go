package httpsession

import (
	"time"

	goauth2 "golang.org/x/oauth2"
)

type SessionData map[string]interface{}

func (data *SessionData) getAccessToken() string {
	signinInfo := (*data)["signin_info"].(map[string]interface{})
	accessTokenMap := (signinInfo)["access_token"].(map[string]interface{})
	return (accessTokenMap)["access_token"].(string)
}

func (data *SessionData) getRefreshToken() string {
	signinInfo := (*data)["signin_info"].(map[string]interface{})
	accessTokenMap := (signinInfo)["access_token"].(map[string]interface{})
	return (accessTokenMap)["refresh_token"].(string)
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

func (s *SessionData) GetOauth2Token() *goauth2.Token {
	if s.isSignedIn() {
		tok := &goauth2.Token{AccessToken: s.getAccessToken(),
			RefreshToken: s.getRefreshToken(),
			Expiry:       time.Now(), //Replace with actual session expiry
		}

		return tok
	}

	return nil
}
