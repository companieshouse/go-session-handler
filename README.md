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
go get
go build
```

Note: this library is not a standalone service, and can only be used within services or other libraries.

## Environment Variables

Key | Description | Scope | Mandatory | Default
----|-------------|-------|-----------|--------
COOKIE_SECRET | The shared secret used in validating/calculating the session cookie signature | State | X |


## Example library usage

To use the main component of this library, the State package, add the following to the relevant package import:
- `"github.com/companieshouse/go-session-handler/state"`
