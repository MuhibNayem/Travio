module github.com/MuhibNayem/Travio/server/services/inventory

go 1.25.3

require (
	github.com/MuhibNayem/Travio/server/api v0.0.0
	github.com/MuhibNayem/Travio/server/pkg v0.0.0
	github.com/gocql/gocql v1.7.0
	github.com/google/uuid v1.6.0
	github.com/redis/go-redis/v9 v9.17.2
	google.golang.org/grpc v1.78.0
)

replace (
	github.com/MuhibNayem/Travio/server/api => ../../api
	github.com/MuhibNayem/Travio/server/pkg => ../../pkg
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
)
