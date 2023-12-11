build:
	go mod tidy
	go mod vendor
	go build .