.PHONY: build clean deploy

build:
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/users_create users/create.go
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/users_get users/get.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
