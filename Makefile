install:
	go get github.com/op/go-logging
	go get gopkg.in/redis.v3

build:
	go build

test:
	go test -v
