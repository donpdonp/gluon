.PHONY: all test

all: build

build: gluon cmd
v8: gov8 rubyv8

gluon: src/gluon/*.go src/gluon/**/*.go
	cd src/gluon; go build -o ../../bin/

cmd: src/gluon/**/*.go
	cd src/gluon; go build -o ../../bin/ donpdonp/gluon/cmd

test:
	GOPATH=`pwd` go test donpdonp/gluon

format:
	gofmt -w src/gluon
