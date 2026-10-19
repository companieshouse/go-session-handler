package httpsession

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/companieshouse/go-session-handler/session"
	"github.com/companieshouse/go-session-handler/state"
	"github.com/justinas/alice"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitRegister(t *testing.T) {
	Convey("Test Register", t, func() {
		var chain = alice.Chain{}
		chain = Register(chain)
		So(chain, ShouldNotBeNil)

		var h http.Handler
		chain.Then(h)
		So(chain, ShouldNotBeNil)
	})
}

func TestUnitHandler(t *testing.T) {
	Convey("Test Handler", t, func() {
		var h http.Handler
		h = handler(h)
		So(h, ShouldNotBeNil)
	})
}

// ---------------- Routes Through getSessionIDFromRequest() ----------------

// TestUnitGetSessionIDFromRequestInvalid - Verify that if a cookie doesn't exist by
// the name the config specifies, a blank session ID is returned
func TestUnitGetSessionIDFromRequestInvalid(t *testing.T) {

	Convey("Given the cookie by the name TEST doesn't exist", t, func() {

		req, _ := http.NewRequest("GET", "teststuff", nil)

		cookie := &http.Cookie{}

		cookie.Name = "NOT_TEST"
		cookie.Value = "Foo"

		req.AddCookie(cookie)

		Convey("When I try to get the session ID from the cookie named 'TEST' on the request", func() {
			sessionID := getSessionIDFromRequest("TEST", req)

			Convey("Then the session ID should be blank", func() {
				So(sessionID, ShouldEqual, "")
			})
		})
	})
}

// TestUnitGetSessionIDFromRequestHappyPath - Verify that if a cookie does exist by
// the name the config specifies, a valid ID is returned
func TestUnitGetSessionIDFromRequestHappyPath(t *testing.T) {

	Convey("Given the cookie by the name TEST exists", t, func() {

		req, _ := http.NewRequest("GET", "teststuff", nil)

		cookie := &http.Cookie{}

		cookie.Name = "TEST"
		cookie.Value = "Foo"

		req.AddCookie(cookie)

		Convey("When I try to get the session ID from the cookie named 'TEST' on the request", func() {
			sessionID := getSessionIDFromRequest("TEST", req)

			Convey("Then the session ID should be 'Foo'", func() {
				So(sessionID, ShouldEqual, "Foo")
			})
		})
	})
}

func TestUnitSetSessionIDOnResponse(t *testing.T) {
	Convey("Test setSessionIDOnResponse", t, func() {
		_ = os.Setenv("COOKIE_NAME", "test-cookie")
		w := httptest.NewRecorder()
		s := &state.Store{}
		setSessionIDOnResponse(w, s)
		So(w.Header().Get("Set-Cookie"), ShouldEqual, "test-cookie=2jmj7l5rSw0yVb/vlWAYkK/YBwk")
	})
}

func TestUnitGetSessionFromRequest(t *testing.T) {
	Convey("Test GetSessionFromRequest", t, func() {
		req, _ := http.NewRequest("GET", "test", nil)
		ctx := context.WithValue(req.Context(), "testkey", "testval")
		s := GetSessionFromRequest(req.WithContext(ctx))
		So(s, ShouldBeNil)

		var sess session.Session = map[string]interface{}{
			"signin_info": map[string]interface{}{
				"access_token": map[string]interface{}{
					"refresh_token": "refresh_token",
				},
			},
		}
		sess.SetAccessToken("testAccessToken")
		ctx = context.WithValue(context.Background(), ContextKeySession, &sess)

		s = GetSessionFromRequest(req.WithContext(ctx))
		So(s.GetAccessToken(), ShouldEqual, "testAccessToken")
	})
}
