module github.com/MuhibNayem/Travio/server/services/queue

go 1.25.3

require (
	github.com/MuhibNayem/Travio/server/api v0.0.0
	github.com/MuhibNayem/Travio/server/pkg v0.0.0
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/redis/go-redis/v9 v9.17.2
	google.golang.org/grpc v1.78.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace (
	github.com/MuhibNayem/Travio/server/api => ../../api
	github.com/MuhibNayem/Travio/server/pkg => ../../pkg
)
