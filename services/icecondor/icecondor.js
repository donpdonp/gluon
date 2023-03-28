// node
var net = require("net")
var uuid = require('node-uuid')
var wsock = require('websock');
var Url = require('url');
var dayjs = require('dayjs')

var settings = require('./settings');

// npm
var redisLib = require("redis"),
    redisSub = redisLib.createClient(),
    redisPub = redisLib.createClient()

var pubsub_channel = 'gluon'
var my_uuid = uuid.v4()

var uri = Url.parse(settings.api)
var stream_id
var usercache = {}
var opened

// redis
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

// websocket go
wsbuild(uri)

function wsbuild(uri) {
  console.log('websocket connecting', uri.href)
  var ws = new wsock.connect(uri, {agent:{rejectUnauthorized:false}})
  ws.on('open', function() {
    redis_pub({method: "icecondor.open"})
    console.log('icecondor connected')
    opened = new Date()
  })
  
  ws.on('message', function(data) {
    try {
      var msg = JSON.parse(data)
	    console.log(JSON.stringify(msg))
      apigo(ws, msg)
    } catch(e) {
      console.log("json err: "+data+e)
    }
  })
  
  ws.on('error', function(data) {
    console.log('ws error')
    console.error(data)
  })
  
  ws.on('close', function() {
    redis_pub({method: "icecondor.closed"})
    var minutes = ((new Date()).getTime() - opened.getTime())/1000/60
    console.log("closed. duration "+minutes.toFixed(1)+"min")
    wsbuild(uri) // what could go wrong
  })
}

function apigo(ws, msg) {
  if(msg.method == "hello") {
    var m = { id: "123",
              method: "auth.session",
              params: {device_key: settings.key}}
	  console.log(JSON.stringify(m))
    ws.send(JSON.stringify(m))
  }
  if(msg.result) {
    if(msg.id == "123") {
      var m = { id: "456",
                method: "stream.follow",
                params: {type: "location", follow: true, order: 'newest'}}
      console.log(JSON.stringify(m))
      ws.send(JSON.stringify(m))
    } else if (msg.id == "456") {
      console.log('fw', JSON.stringify(msg.result))
      stream_id = msg.result.stream_id
      var added = msg.result.added[0]
      usercache[added.id] = added.username
    } else if (stream_id && msg.id == stream_id) {
      var username = usercache[msg.result.user_id]

      var now = new Date()
      var ldate = new Date(msg.result.date)
      var minago = (now.getTime() - ldate.getTime())/1000/60
      if (minago < 60) { 
       if (minago < 2) { 
        ago = (minago*60).toFixed(0)+ " sec" 
       } else {
        ago = minago.toFixed(1)+ " min" 
       }
      } else { ago = (minago/60).toFixed(1)+" hours" }

      var rdate = new Date(msg.result.received_at)
      var rminago = (rdate.getTime() - ldate.getTime())/1000/60
      if (msg.result.latitude) {
        console.log(dayjs(now).format('HH:mm:ss'), 
                    username, msg.result.latitude.toFixed(5),
                    msg.result.longitude.toFixed(5),
                    ago, "ago.", rminago.toFixed(1), "min delay")
        if (minago < 60*24) {
          redis_pub({method: "icecondor.location",
                    params: {username: username,
                             latitude: msg.result.latitude,
                             longitude: msg.result.longitude,
                             date: msg.result.date,
                             received_at: msg.result.received_at,
                             accuracy: parseFloat(msg.result.accuracy)
                           }})
        } else {
          console.log(username, 'too old')
        }
      } else {
        console.log(username, 'cloaked')
      }
    } else {
      console.log('unknown response:', msg)
    }
  }
}


function redis_pub(msg){
  msg["from"] = my_uuid
  msg["id"] = msg["id"] || uuid.v4()
  msg["key"] = settings.gluon_key
  var json = JSON.stringify(msg)
//  console.log('redis>', json)
  redisPub.publish(pubsub_channel, json)
}

function dispatch(payload) {
  var cmd = payload.method.split('.')[1]
  if(cmd == 'delete') {
  }
}

