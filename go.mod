module github.com/lmas/d0020e_code

go 1.23
require (
	github.com/gorilla/websocket v1.5.3
	github.com/influxdata/influxdb-client-go/v2 v2.14.0
	github.com/sdoque/mbaigo v0.0.0-20241019053937-4e5abf6a2df4
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/influxdata/line-protocol v0.0.0-20200327222509-2487e7298839 // indirect
	github.com/oapi-codegen/runtime v1.0.0 // indirect
	golang.org/x/net v0.36.0 // indirect
)

// Replaces this library with a patched version
replace github.com/sdoque/mbaigo v0.0.0-20241019053937-4e5abf6a2df4 => github.com/lmas/mbaigo v0.0.0-20250123014631-ad869265483c
