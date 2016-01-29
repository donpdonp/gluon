## CPP spike (Oct 2015)

* redis pub/sub
* mruby

## Go spike (Dec 2015)

* nanomsg/mango (dropped)
* redis pub/sub
* otto javascript

Otto VM container

### JS API


go function is called on every pub/sub msg. filter by msg.method

```js

function go(msg) {
  var word_match = /^keyword$/.exec(msg.params.message)

  if (msg.method == "clocktower") {
    var time = new Date(Date.parse(msg.params.time))
  } // clocktower called every minute
}
```
