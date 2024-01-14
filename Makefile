build:
	rm -r -f logs
	go mod tidy
	go mod vendor
	go build .