package main

type commandID int

const (
	CMD_REGISTER commandID = iota
	CMD_LOGIN
	CMD_LOGOUT
	CMD_NICK
	CMD_JOIN
	CMD_ROOMS
	CMD_MSG
	CMD_QUIT
	CMD_USERS
	CMD_DM
	CMD_STATUS
	CMD_HELP
	CMD_HISTORY
)

type commands struct {
	id     commandID
	client *client
	args   []string
}
