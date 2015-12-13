all: build

src/github.com/robertkrimen/otto:
	go get github.com/robertkrimen/otto

src/gopkg.in/redis.v3:
	go get gopkg.in/redis.v3

src/github.com/gdamore/mangos:
	go get -u github.com/gdamore/mangos

build:  cmd

cmd:    src/donpdonp/gluon
	go build donpdonp/gluon/cmd

