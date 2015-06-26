all: build/ mruby/.git mruby/build/host/lib/libmruby.a gluon

CFLAGS=-std=c99 -g
CPPFLAGS=-I mruby/include -I mruby/build/mrbgems
LDFLAGS=-L mruby/build/host/lib -lmruby -lm -lhiredis -lcurl -lpcre

mruby/.git:
	git clone https://github.com/mruby/mruby.git
	cp build_config.rb mruby/

mruby/build/host/lib/libmruby.a:
	make -C mruby
	# Hack: Freshen parson lib
	curl https://raw.githubusercontent.com/kgabis/parson/master/parson.h > mruby/build/mrbgems/mruby-json/src/parson.h
	curl https://raw.githubusercontent.com/kgabis/parson/master/parson.c > mruby/build/mrbgems/mruby-json/src/parson.c
	make -C mruby

build/:
	mkdir build

build/%.o: src/%.c
	$(CC) -c $(CPPFLAGS) -o $@ $< $(CFLAGS) 

gluon: build/main.o build/eval_mruby.o
	gcc -o gluon $^ mruby/build/host/mrbgems/mruby-json/src/parson.o $(LDFLAGS)

