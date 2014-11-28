neur0n: src/main.c
	gcc src/main.c -o neur0n -I mruby/include -L mruby/build/host/lib -lmruby -lm -lhiredis

