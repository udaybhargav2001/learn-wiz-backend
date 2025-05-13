gen:
	protoc --proto_path=internal/api/proto \
		--go_out=internal/api/pb --go_opt=paths=source_relative \
		--go-grpc_out=internal/api/pb --go-grpc_opt=paths=source_relative \
		internal/api/proto/*.proto
server:
	nodemon --watch './**/*.go' --signal SIGTERM --exec 'go' run main.go