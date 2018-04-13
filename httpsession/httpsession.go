package httpsession

import (
	"context"
	"net/http"
	"os"
	"strconv"

	"github.com/companieshouse/chs.go/log"
	session "github.com/companieshouse/go-session-handler/session"
	"github.com/companieshouse/go-session-handler/state"
	"github.com/justinas/alice"
	redis "gopkg.in/redis.v5"
)

// Type for creating context keys
type ContextKey string

// Set the context key for the session
var ContextKeySession = ContextKey("session")

// Register will append an HTTP handler to an Alice chain, whereby the stored
// session will be loaded and stored on the request context
func Register(c alice.Chain) alice.Chain {
	return c.Append(func(h http.Handler) http.Handler { return handler(h) })
}

// loadSession initialises a Store using config and cache structs, loads the
// session, and stores it on the request context to access later
func handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		// Init all config
		config := &state.StoreConfig{
			CookieName:        os.Getenv("COOKIE_NAME"),
			CookieSecret:      os.Getenv("COOKIE_SECRET"),
			DefaultExpiration: os.Getenv("DEFAULT_EXPIRATION"),
		}

		redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
		if err != nil {
			log.Error(err)
			redisDB = 0
		}

		cache, _ := state.NewCache(&redis.Options{
			Addr:     os.Getenv("REDIS_SERVER"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       redisDB,
		})

		s := state.NewStore(cache, config)

		// Pull session ID from the cookie on the request
		sessionID := getSessionIDFromRequest(config.CookieName, req)
		var sessionData session.SessionData

		// If session is stored, retrieve it from Redis
		if sessionID != "" {

			if err := s.Load(sessionID); err == nil {
				sessionData = s.Data
			}
		}

		ctx := context.WithValue(context.Background(), ContextKeySession, sessionData)
		req = req.WithContext(ctx)
		h.ServeHTTP(w, req)

		// Upon returning, store the updated session
		s.Data = sessionData
		s.Store()

		//TODO: Update the session cookie here using setSessionIDOnRequest
	})
}

//getSessionIDFromRequest will attempt to pull the Cookie from the request. If err
//is not nil, it will create a new Cookie and return that instead.
func getSessionIDFromRequest(cookieName string, req *http.Request) string {

	cookie, err := req.Cookie(cookieName)
	if err != nil {
		log.InfoR(req, err.Error())
		cookie = &http.Cookie{}
	}

	return cookie.Value
}

func setSessionIDOnRequest(w http.ResponseWriter) {
	cookie := &http.Cookie{} //TODO: construct cookie to store on the response
	http.SetCookie(w, cookie)
}
