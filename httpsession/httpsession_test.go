package httpsession

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

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
