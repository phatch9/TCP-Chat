# TCP-Chat

## Practice Project using Golang 

Building a TCP chat server using Go, which enables clients to communicate with each other. This project starts with Go's “net” package that well supports TCP, as well using channels and goroutines

## Try the server:
- Provide a module path when running
```
go mod init
```
- Cleans up go.mod (Optional):
```
go mod tidy
```
- Build the chat on localhost:
```
go build .
```

On a different terminal, run:
```
telnet localhost 8888
```
Chat server run successful output:
```
Trying ::1...
Connected to localhost.
Escape character is '^]'.
```
- Start with /nick yourname -


## Features

### Core Features
- **Multi-room support**: Create and join chat rooms instantly
- **Real-time messaging**: Broadcast messages to all users in a room
- **User management**: Set custom usernames (nicknames)
- **User status**: Set status like away, busy, online, etc.
- **Direct messaging**: Send private messages to individual users
- **Message history**: Retrieve recent messages from a room
- **Online user listing**: See all users in your current room with their status

### Technical Features  
- **Thread-safe concurrency**: Uses mutexes to prevent race conditions
- **Buffered command channel**: Handles high-load scenarios efficiently
- **Automatic cleanup**: Empty rooms are automatically removed
- **Graceful shutdown**: Proper signal handling for clean server shutdown
- **Connection resilience**: Handles client disconnections gracefully

## Command Lines

| Command | Usage | Description |
|---------|-------|-------------|
| `/nick` | `/nick <name>` | Set your username (2-20 characters) |
| `/join` | `/join <room>` | Join or create a chat room |
| `/rooms` | `/rooms` | List all available rooms |
| `/users` | `/users` | List all users in current room |
| `/msg` | `/msg <message>` | Send a message to the room |
| `/dm` | `/dm <user> <message>` | Send a private message to a user |
| `/status` | `/status <status>` | Set your availability status |
| `/history` | `/history [n]` | Show last n messages (default: 10) |
| `/help` | `/help` | Display all commands |
| `/quit` | `/quit` | Exit the chat |

## Architecture

```
┌─────────────────────────────────────┐
│         TCP Chat Server             │
├─────────────────────────────────────┤
│ main.go      - Entry point, listen  │
│ server.go    - Command handlers     │
│ client.go    - Client connection    │
│ room.go      - Room & messages      │
│ commands.go  - Command definitions  │
└─────────────────────────────────────┘
```

## Build & Run

### Prerequisites
- Go 1.25.0 or higher

### Build the Project
```bash
cd /Users/phch/Documents/TCP-Chat
go build .
```

### Start the Server
```bash
./TCP-Chat
```

Server will start on `localhost:8888`

### Connect Clients
In separate terminal(s):
```bash
telnet localhost 8888
```

## Example Session

**Terminal 1:**
```
$ telnet localhost 8888
/nick Alice
/join general
/msg Hello everyone!
/users
/dm bob Nice to meet you!
/history 5
/quit
```

**Terminal 2:**
```
$ telnet localhost 8888
/nick Bob
/join general
/msg Hi Alice!
```

## Security & Improvements Implemented

✅ **Mutex Protection** - All shared data structures protected with RWMutex  
✅ **Input Validation** - Commands validated for required arguments  
✅ **Error Handling** - Proper error messages for all operations  
✅ **Nick Validation** - Nick names limited to 2-20 characters  
✅ **Connection Handling** - Automatic cleanup on disconnect  
✅ **Timestamp Support** - All messages include timestamps  
✅ **Message History** - Auto-buffers last 100 messages per room  
✅ **Empty Room Cleanup** - Rooms deleted when last user leaves  
✅ **Buffered Channels** - Command channel handles high throughput  

## Performance Features

- Concurrent client handling with goroutines
- Lock-free reads where possible using RWMutex  
- Buffered command channel (100 buffer size)
- Efficient message history with fixed-size buffer
- Automatic resource cleanup

## Future Enhancements

- [ ] User authentication & persistence
- [ ] SSL/TLS encryption
- [ ] File sharing support
- [ ] User profiles & avatars
- [ ] Message search
- [ ] Room permissions
- [ ] Web UI client
- [ ] Database integration

---

**Built with Go using goroutines and channels for concurrent networking.**
