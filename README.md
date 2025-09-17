# TCP-Chat

## Practice Project using Golang 

Building a TCP chat server using Go, which enables clients to communicate with each other. This project starts with Go's “net” package that well supports TCP, as well using channels and goroutines

## Command

- Create a name for user, otherwise user remains anonymous.
```
/nick <name>
```
- User can join a room, or this will create a new room if not existed. Note: A user can join and stay in one room at a time.
```
/join <name>
```
- Display a list of all available rooms to join.
```
/room
```
- Send a message to everyone the room server.
```
/msg <msg>
```
- Exit the room chat.
```
/quit
```