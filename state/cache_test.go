package state

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/redis.v5"
)

func TestUnitNewCache(t *testing.T) {
	Convey("test new cache", t, func() {
		c := NewCache("address", 123, "password")
		So(c, ShouldNotBeNil)
	})
}

func TestUnitSetRedisClient(t *testing.T) {

	Convey("test set redis client", t, func() {
		cache := &Cache{}
		opts := redis.Options{}
		cache.setRedisClient(&opts)
		So(cache.connection, ShouldNotBeNil)

	})
}
