'use strict';

const http = require('http');

const server = http.createServer(function (req, res) {
    res.writeHead(200, {'Content-Type': 'text/plain'});
    res.end(process.env.ECHO_MESSAGE || "")
});

const PORT = process.env.PORT || 8080
server.listen(PORT);

console.log('Server listening on port ' + PORT)