build:
	rm -f -r .orca
	go mod tidy
	go mod vendor
	go build .
	go test ./lib/... -v -count=1

test:
	go test ./lib/... -v -count=1