package main

type commandID int

const (
	CMD_NICK commandID = iota
	CMD_JOIN
	CMD_ROOMS
	CMD_MSG
	CMD_QUIT
	CMD_USERS
	CMD_DM
	CMD_STATUS
	CMD_HELP
)

type commands struct {
	id     commandID
	client *client
	args   []string
}
