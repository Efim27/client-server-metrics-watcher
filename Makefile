doc:
	godoc -all . -http=:8089

swag:
	swag init --output ../../swagger/

pprof:
	go tool pprof -http=":9090" -seconds=30 http://localhost:8090/debug/pprof/profile

pprof.heap:
	go tool pprof -http=":9090" -seconds=30 http://localhost:8080/debug/pprof/heap

pprof.diff:
	go tool pprof -top -diff_base=./profiles/base.pprof ./profiles/result.pprof

test:
	go test ./...

test.integration:
	go test -tags=integration ./...

test.cover:
	go test -tags=integration ./... -coverprofile cover.out
	go tool cover -func cover.out

vet:
	go vet .\internal\agent\...

lint:
	go run .\cmd\staticlint\main.go .\internal\...

protoc:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/metrics.proto