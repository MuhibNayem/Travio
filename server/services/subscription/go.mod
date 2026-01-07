module github.com/MuhibNayem/Travio/server/services/subscription

go 1.25.3

require (
	github.com/MuhibNayem/Travio/server/api v0.0.0-00010101000000-000000000000
	github.com/MuhibNayem/Travio/server/pkg v0.0.0-20260107213724-9e1bd5afa1a6
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	google.golang.org/grpc v1.78.0
)

require (
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/MuhibNayem/Travio/server/api => ../../api
