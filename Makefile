deps:
	go get github.com/docker/docker/api
	go get github.com/docker/docker/client

engine-check-arm:
	GOOS=linux GOARCH=arm go build -o engine-check-arm rmazur.io/engine-check/cmd/engine-check
