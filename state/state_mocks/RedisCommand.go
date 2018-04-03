// Code generated by mockery v1.0.0
package mocks

import mock "github.com/stretchr/testify/mock"
import redis "gopkg.in/redis.v5"

import time "time"

// RedisCommand is an autogenerated mock type for the RedisCommand type
type RedisCommand struct {
	mock.Mock
}

// GetSessionData provides a mock function with given fields: key
func (_m *RedisCommand) GetSessionData(key string) (string, error) {
	ret := _m.Called(key)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetSessionData provides a mock function with given fields: key, value, expiration
func (_m *RedisCommand) SetSessionData(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	ret := _m.Called(key, value, expiration)

	var r0 *redis.StatusCmd
	if rf, ok := ret.Get(0).(func(string, interface{}, time.Duration) *redis.StatusCmd); ok {
		r0 = rf(key, value, expiration)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*redis.StatusCmd)
		}
	}

	return r0
}
