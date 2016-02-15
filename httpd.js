#!/usr/bin/env node

var http = require('http');
http.createServer(function (req, res) {
    res.writeHead(200, {'Content-Type': 'text/plain'});
    res.end('Hello World from node.js/012345678901234567890123456789012345678901234567890123456789012345678');
}).listen(8080);
