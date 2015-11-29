all: src/github.com/robertkrimen/otto src/gopkg.in/redis.v3 src/github.com/gdamore/mangos build

src/github.com/robertkrimen/otto:
	go get github.com/robertkrimen/otto

src/gopkg.in/redis.v3:
	go get gopkg.in/redis.v3

src/github.com/gdamore/mangos:
	go get -u github.com/gdamore/mangos

build:
	go build

run:
	go run gluon.go
