all: build

build: gluon cmd

gluon: src/donpdonp/gluon/*.go src/donpdonp/gluon/**/*.go
	go build donpdonp/gluon

cmd: src/donpdonp/gluon/**/*.go
	go build donpdonp/gluon/cmd

