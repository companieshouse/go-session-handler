package state

import (
	"errors"
	"testing"
	"time"

	mockState "github.com/companieshouse/go-session-handler/state/state_mocks"
	. "github.com/smartystreets/goconvey/convey"

	redis "gopkg.in/redis.v5"
)

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

			err := s.setSession("")

			Convey("Then I expect the error to be caught and returned", func() {

				So(err, ShouldNotBeNil)
				So("Unsuccessful save", ShouldEqual, err.Error())
			})
		})
	})
}
