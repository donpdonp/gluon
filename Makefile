all: mruby/.git mruby/build neur0n

mruby/.git:
	git clone https://github.com/mruby/mruby.git

mruby/build:
	make -C mruby

src/main.o: src/main.c
	gcc src/main.c -g -c -o src/main.o -I mruby/include -I mruby/build/mrbgems

neur0n: src/main.o
	gcc -o neur0n src/main.o mruby/build/host/mrbgems/mruby-json/src/parson.o -L mruby/build/host/lib -lmruby -lm -lhiredis

