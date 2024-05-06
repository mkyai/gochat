# GO-CHAT
Realtime chat application using websockets and golang. 

## Tech Stack
- Golang
- Websockets
- MongoDB
- Crypto
- JWT

## Features
- Realtime chat
- User authentication
- User registration
- User search
- Channel creation
- Channel listing

## Installation
1. Clone the repository
2. Run `go mod tidy`
3. Run application at port 3020
```bash
go run .
```

## Usage
1. Register a new user
```bash
curl -X POST http://localhost:3020/signup -d '{"username": "test" ,"name": "test" ,"password": "test"}'
```
2. Login with the registered user
```bash
curl -X POST http://localhost:3020/login -d '{"username": "test" ,"password": "test"}'
```
3. List all users
```bash
curl -X GET http://localhost:3020/users
```
4. Create a new channel
```bash
curl -X POST http://localhost:3020/channels -d '{"user": "__userID__"}' -H "Authorization: Bearer __JWT__"
```
5. List all channels
```bash
curl -X GET http://localhost:3020/channels -H "Authorization: Bearer __JWT__"
```
6. List all messages in a channel
```bash
curl -X GET http://localhost:3020/messages/{channelID} -H "Authorization: Bearer __JWT__"
```

7. Start the chat [WEBSOCKET]
```bash
connect at ws://localhost:3020/{channelID}?token=__JWT__
```

## TODO
- [ ] Verify Dockerfile
- [ ] Kafka integration
- [ ] Redis integration
- [ ] Frontend Templates
- [ ] Group chat
- [ ] Message encryption
- [ ] CI/CD


