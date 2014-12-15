all: mruby/.git mruby/build neur0n

CFLAGS=-std=c99 -g
CPPFLAGS=-I mruby/include -I mruby/build/mrbgems
LDFLAGS=-L mruby/build/host/lib -lmruby -lm -lhiredis -lcurl

mruby/.git:
	git clone https://github.com/mruby/mruby.git

mruby/build:
	make -C mruby

neur0n: src/main.o src/eval_mruby.o
	gcc -o neur0n $^ mruby/build/host/mrbgems/mruby-json/src/parson.o $(LDFLAGS)

