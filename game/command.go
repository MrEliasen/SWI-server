package game

import (
	"strings"

	"github.com/mreliasen/swi-server/internal/responses"
)

type Command struct {
	CommandKey    string
	Args          []string
	Description   string
	Example       string
	AllowInGame   bool
	AllowAuthed   bool
	AllowUnAuthed bool
	AdminCommand  bool
	Call          func(c *Client, args []string)
	Help          func(c *Client)
}

func ExecuteCommand(c *Client, msg string) {
	if msg == "" {
		return
	}

	args := strings.Fields(msg)
	cmdKey := strings.ToLower(args[0])
	args = args[1:]
	isUnAuthed := !c.Authenticated
	isAuthed := c.Authenticated
	isInGame := c.Player != nil
	isAdmin := false
	help := false

	if isInGame {
		isAdmin = c.Player.IsAdmin
	}

	if len(args) == 1 {
		if args[0] == "help" {
			help = true
		}
	}

	if alias, ok := CommandAliases[cmdKey]; ok {
		cmdKey = alias
	}

	var cmdToRun *Command

	if command, ok := CommandsList[cmdKey]; ok {
		if (isInGame && command.AllowInGame) || (isAuthed && command.AllowAuthed) || (isUnAuthed && command.AllowUnAuthed) {
			cmdToRun = command
		}
	}

	if cmdToRun == nil && isInGame {
		if cmd, found := c.Player.PlayerGetBuildingCommand(cmdKey); found {
			cmdToRun = cmd
		}
	}

	if cmdToRun != nil {
		if !cmdToRun.AdminCommand || (cmdToRun.AdminCommand && isAdmin) {
			if help {
				cmdToRun.Help(c)
				return
			}

			cmdToRun.Call(c, args)
			return
		}
	}

	c.SendEvent(&responses.Generic{
		Status: responses.ResponseStatus_RESPONSE_STATUS_ERROR,
		Messages: []string{
			"Unknown command",
		},
	})
}
