all: mruby/.git mruby/build neur0n

CPPFLAGS=-I mruby/include -I mruby/build/mrbgems
LDFLAGS=-L mruby/build/host/lib -lmruby -lm -lhiredis

mruby/.git:
	git clone https://github.com/mruby/mruby.git

mruby/build:
	make -C mruby

neur0n: src/*.o
	gcc -o neur0n src/*.o mruby/build/host/mrbgems/mruby-json/src/parson.o $(LDFLAGS)

