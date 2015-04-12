# gluon

Gluon is a container for running multiple language runtimes in a single process. Message passing json over pub/sub allows for command and control. This is a place to manage the execution of untrusted scripts.

# Setup

```
$ git clone https://github.com/donpdonp/gluon.git
$ cd gluon
$ make
(builds prerequisite mruby)
$ cp config.json.sample config.json
$ ./gluon
new machine #0 allocated for admin
```

# messages

Push these structures as JSON strings into the 'neur0n' redis channel (defined in config.json)

```json
{"type":"vm.add","name":"bobo4","url":"https://gist/funny-responses.rb"}
{"type":"irc.connect","server":"irc.freenode.net","nick":"n0bot"}'
{"type":"irc.join","network":"freenode","channel":"#pdxbots"}'
{"type":"irc.privmsg","network":"freenode","channel":"#pdxbots", "message":"I am here."}
```

# Supported language/runtimes

* ruby/mruby (done)
* javascript/v8 (planned)

