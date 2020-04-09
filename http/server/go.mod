module github.com/keys-pub/keysd/http/server

go 1.12

require (
	github.com/gorilla/websocket v1.4.2
	github.com/keys-pub/keys v0.0.0-20200409193232-85fe44951a30
	github.com/keys-pub/keysd/firestore v0.0.0-20200402183018-a85eceb453b1
	github.com/keys-pub/keysd/http/api v0.0.0-20200402183018-a85eceb453b1
	github.com/labstack/echo/v4 v4.1.11
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.4.0
	github.com/valyala/fasttemplate v1.1.0 // indirect
	google.golang.org/api v0.20.0
)

// replace github.com/keys-pub/keys => ../../../keys

// replace github.com/keys-pub/keysd/http/api => ../api

// replace github.com/keys-pub/keysd/firestore => ../../firestore
