all: build/ mruby/.git mruby/build/host/lib/libmruby.a gluon

CFLAGS=-std=c99 -g
CPPFLAGS=-I mruby/include -I mruby/build/mrbgems
LDFLAGS=-L mruby/build/host/lib -lmruby -lm -lhiredis -lcurl -lpcre

mruby/.git:
	git clone https://github.com/mruby/mruby.git
	cp build_config.rb mruby/

mruby/build/host/lib/libmruby.a:
	make -C mruby

build/:
	mkdir build

build/%.o: src/%.c
	$(CC) -c $(CPPFLAGS) -o $@ $< $(CFLAGS) 

gluon: build/main.o build/eval_mruby.o build/parson.o
	gcc -o gluon $^ $(LDFLAGS)

