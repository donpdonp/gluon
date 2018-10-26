all: build

build: gluon cmd
v8: gov8 rubyv8

gluon: src/donpdonp/gluon/*.go src/donpdonp/gluon/**/*.go
	GOPATH=`pwd` go build donpdonp/gluon

cmd: src/donpdonp/gluon/**/*.go
	GOPATH=`pwd` go build donpdonp/gluon/cmd

gov8:
	go get github.com/augustoroman/v8
	#git clone https://github.com/augustoroman/v8

rubyv8:
	echo 'source "https://rubygems.org"' > Gemfile
	echo 'gem "libv8", "~> 6.7", ">= 6.7.288.46.1"' >> Gemfile
	bundle install --path gems
	tar xCf gems ./gems/ruby/2.5.0/cache/libv8-6.7.288.46.1-x86_64-linux.gem
	tar xCf gems ./gems/data.tar.gz
	ln -fs $(shell pwd)/gems/vendor/v8/include ./src/github.com/augustoroman/v8/include
	ln -fs $(shell pwd)/gems/vendor/v8/out.gn/libv8/obj ./src/github.com/augustoroman/v8/libv8
