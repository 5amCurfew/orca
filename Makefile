build:
	rm -f -r .orca
	go mod tidy
	go mod vendor
	go build .