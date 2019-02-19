var express 	= require('express');
var app 		= express();
var server 		= require('http').createServer(app);
var io 			= require('socket.io').listen(server);
var redis 		= require('socket.io-redis');
var velox 		= require('veloxjs');
var port 		= process.env.PORT || 3000;

io.adapter(redis({ host: 'redis', port: 6379 }));

server.listen(port, function() {
  console.log('Server listening at port %d', port);
});

app.all('/', function(req, res) {
	res.header('Access-Control-Allow-Origin', '*');
	res.json({'success': true});
});

app.all('/emit', function(req, res) {

	res.header("Access-Control-Allow-Origin", "*");

	var room  	= req.query.room;
	var event 	= req.query.event;
	var data  	= req.query.data;

	try {
		data = JSON.parse(data);
	} catch (e) {
		data = req.query.data;
	}

	if (room) {
		console.log(room, event, data);
		io.sockets.in(room).emit(event, data);
	} else {
		console.log(event, data);
		io.sockets.emit(event, data);
	}

	res.json({'success': true});

});

var config = {};
var searchProviders = {};

var foo   = {};
var v     = velox("http://cloudtorrent:3000/sync", foo);

v.onupdate = function(object) {

	// console.log('onupdate', object);

	config = object.Config;
	searchProviders = object.SearchProviders;

	io.sockets.emit('update', object);
	io.sockets.emit('downloads', object.Downloads);
	io.sockets.emit('stats', object.Stats);
	io.sockets.emit('torrents', object.Torrents);
	io.sockets.emit('users', object.Users);
	
};

v.onconnect = function(object) {
	console.log('onconnect');
};

io.on('connection', function(socket) {

	console.log('a connection has been created: ' + io.engine.clientsCount);

	socket.on('subscribe', function(channel) { 
		console.log('joining channel', channel);
		socket.join(channel);
		socket.emit(channel + '|subscribe', 1);
	});

	socket.on('unsubscribe', function(channel) {  
		console.log('leaving channel', channel);
		socket.emit(channel + '|unsubscribe', 1);
	});

	socket.on('channel', function(data) { 
		io.sockets.in(data.channel).emit(data.channel + '|' + data.event, data.data);
	});

	socket.on('disconnect', function() {
		console.log('a connection has been terminated: ' + io.engine.clientsCount);
	});

	socket.emit('config', config);
	socket.emit('searchProviders', searchProviders);

});