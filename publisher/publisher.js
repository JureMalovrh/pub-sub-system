
const WebSocket = require('ws');
const path = require('path')
const config = require(path.resolve('config.json'))


const server = new WebSocket.Server({ host: config.host, port: config.port });
console.log('Socket server listening on', config.host, config.port)

server.broadcast = (message, sender) => {
    for(listener of server.clients) {
        if (listener !== sender && listener.readyState === WebSocket.OPEN) {
            listener.send(message);
        }
    }
}

server.on('connection', (socket) => {
    console.log('New client connected')
    socket.send('Successfully connected to publisher');

    socket.on('message', (message) => {
        console.log('received: %s', message);
        server.broadcast(message, socket);
    });

    socket.on('close', () => {
        console.log('Closed connection')
    });

    socket.on('disconect', () => {
        console.log('Disconected connection')
    });

    socket.on('error', (err) => {
        console.log('Error', err)
    });
});

server.on('error', (err) => {
    console.log('Error occured when running', err)
})