package rcon

import (
	"strings"
)

// Player represents a registered player on the server.
type Player struct {
	Name   string
	Online bool
}

// CmdPlayers returns all registered players on the server.
func (r *RCON) CmdPlayers() (players []Player, err error) {
	resp, err := r.Execute("/players")
	if err != nil {
		return
	}
	lines := strings.Split(resp.Body, "\n")
	for i, line := range lines {
		if i == 0 {
			// First line is header with total players listed.
			continue
		}
		if len(strings.TrimSpace(line)) == 0 {
			// Last line is just a return. Do not want.
			continue
		}

		var name string
		var online bool
		if strings.HasSuffix(line, "(online)") {
			nameStatusSplit := strings.Split(line, "(")
			name = strings.TrimSpace(nameStatusSplit[0])
			online = true
		} else {
			name = strings.TrimSpace(line)
		}
		players = append(players, Player{Name: name, Online: online})
	}
	return
}

// CmdAdmins returns all registered admin players on the server.
func (r *RCON) CmdAdmins() (players []Player, err error) {
	resp, err := r.Execute("/admins")
	if err != nil {
		return
	}
	lines := strings.Split(resp.Body, "\n")
	for _, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			// Last line is just a return. Do not want.
			continue
		}

		var name string
		var online bool
		if strings.HasSuffix(line, "(online)") {
			nameStatusSplit := strings.Split(line, "(")
			name = strings.TrimSpace(nameStatusSplit[0])
			online = true
		} else {
			name = strings.TrimSpace(line)
		}
		players = append(players, Player{Name: name, Online: online})
	}
	return
}
