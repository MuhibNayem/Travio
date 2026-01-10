module github.com/MuhibNayem/Travio/server/services/gateway

go 1.25.3

require (
	github.com/MuhibNayem/Travio/server/api v0.0.0
	github.com/MuhibNayem/Travio/server/pkg v0.0.0
	github.com/MuhibNayem/Travio/server/pkg/entitlement v0.0.0-20260107231921-4b2f95ebcb2e
	github.com/go-chi/chi/v5 v5.0.12
	github.com/go-chi/cors v1.2.1
	github.com/go-chi/render v1.0.3
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/uuid v1.6.0
	github.com/redis/go-redis/v9 v9.17.2
	github.com/sony/gobreaker v1.0.0
	google.golang.org/grpc v1.78.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/ajg/form v1.5.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/MuhibNayem/Travio/server/api => ../../api
	github.com/MuhibNayem/Travio/server/pkg => ../../pkg
)
