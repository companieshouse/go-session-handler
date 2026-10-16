package config

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitGet(t *testing.T) {

	Convey("Config tests", t, func() {
		// cfg is nil
		config := Get()
		So(config, ShouldNotBeNil)

		// cfg is not nil
		config = Get()
		So(config, ShouldNotBeNil)
	})
}
