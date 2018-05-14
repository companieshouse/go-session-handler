# go-session-handler

Go implementation of CHS session handling. It deals with two main areas:
- Loading of the session from the backend store
- Storing of the session in the backend store

The backend store is Redis.

## Requirements

In order to build this library locally you will need the following:
- [Go](https://golang.org/)
- [Git](https://git-scm.com/downloads)

## Getting started

The library is built using the following commands:
```
go get ./...
go build
```

## Packages
The library contains the the following packages:
- state
- encoding
- httpsession
- session

#### State
The `state` package handles the loading and storing of the session within the cache. The `store.go` has the functions to deal with
loading/storing, whilst `cache.go` deals provides an interface for connecting to the cache (in theory this can be replaced with
another cache that isn't Redis).

#### Encoding
The `encoding` package wraps a few different encoding libraries to provide standard encoding for our sessions. It provides functions
for encoding and decoding both [base64](https://golang.org/pkg/encoding/base64/) and [messagepack](https://github.com/vmihailenco/msgpack) encodings.

#### HttpSession
The `httpsession` package gives the user the ability to register with an [alice chain](https://github.com/justinas/alice) and provide a
Handler.

#### session
The `session` package provides some useful helper functions to retrieve commonly used Session data from the stored Session map.  

## Testing
The library can be tested by running the following in the command line (in the `go-session-handler` directory):
```
goconvey
```

Note: this library is not a standalone service, and can only be used within services or other libraries.

## Environment Variables

The following environment variables are required when integrating the session handler into any Go service.
Note: this library uses `gofigure` to manage environment variables. These variables must not be overridden by applications using the library.

Key | Description | Scope | Mandatory
----|-------------|-------|-----------
COOKIE_SECRET | The shared secret used in validating/calculating the session cookie signature | State | Y
COOKIE_NAME | The name of the cookie from which to retrieve the session ID | HttpSession | Y
DEFAULT_SESSION_EXPIRATION | Default session expiration in seconds | State | Y
REDIS_SERVER | Server address for the Redis database | HttpSession | Y
REDIS_DB | The Redis database number (integer) | HttpSession | Y
REDIS_PASSWORD | Password to access the Redis database | HttpSession | Y


## Example library usage

To use the main component of this library, the State package, add the following to the relevant package import:
- `"github.com/companieshouse/go-session-handler/state"`
