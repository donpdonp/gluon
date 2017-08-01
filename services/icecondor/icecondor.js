// node
var net = require("net")
var uuid = require('node-uuid')
var wsock = require('websock');
var Url = require('url');

var settings = require('./settings');

// npm
var redisLib = require("redis"),
    redisSub = redisLib.createClient(),
    redisPub = redisLib.createClient()

var pubsub_channel = 'gluon'
var my_uuid = uuid.v4()

var uri = Url.parse(settings.api)

var ws = new wsock.connect(uri, {agent:{rejectUnauthorized:false}})
ws.on('open', function() {
  redis_pub({method: "icecondor.open"})
  console.log('icecondor connected')
})
ws.on('message', function(data) {
  //process.stdout.write("ic: "+data)
  var msg = JSON.parse(data)
  if(msg.method == "hello") {
    var m = { id: "123",
              method: "auth.session",
              params: {device_key: settings.key}}
    ws.send(JSON.stringify(m))
  }
  if(msg.result) {
    if(msg.id == "123") {
      var m = { id: "456",
                method: "stream.follow",
                params: {type: "location", follow: true}}
      ws.send(JSON.stringify(m))
    } else {
      console.log(msg.result.id, msg.result.latitude, msg.result.longitude)
      redis_pub({method: "icecondor.location",
                params: {user_id: msg.result.user_id,
                         latitude: msg.result.latitude,
                         longitude: msg.result.longitude,
                         date: msg.result.date,
                         accuracy: msg.result.accuracy
                       }})
    }
  }
});
ws.on('error', function(data) {
  console.error(data)
});
ws.on('close', function() {
  redis_pub({method: "icecondor.closed"})
  console.log("closed")
});



redisSub.on("subscribe", function (channel, count) {
  console.log("redis subscribe "+channel)
})

redisSub.on("message", function (channel, message) {
  try {
    var payload = JSON.parse(message)
    if(payload.method && payload.method.match(/^db\./)) {
      dispatch(payload)
    }
  } catch(err) {
      console.log("redis json err %s", err);
  }
})

redisSub.subscribe(pubsub_channel)

function redis_pub(msg){
  msg["from"] = my_uuid
  msg["id"] = msg["id"] || uuid.v4()
  msg["key"] = settings.gluon_key
  var json = JSON.stringify(msg)
  console.log('redis>', json)
  redisPub.publish(pubsub_channel, json)
}

function dispatch(payload) {
  var cmd = payload.method.split('.')[1]
  if(cmd == 'delete') {
  }
}

