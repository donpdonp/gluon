all: mruby/.git mruby/build neur0n

mruby/.git:
	git clone https://github.com/mruby/mruby.git

mruby/build:
	make -C mruby

neur0n: src/main.c
	gcc src/main.c -o neur0n -I mruby/include -L mruby/build/host/lib -lmruby -lm -lhiredis

