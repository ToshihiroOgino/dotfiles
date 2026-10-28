package tmuxctl

import (
	"fmt"
	"strings"
)

// ClientFormat is the list-clients format that ClientTable parses.
const ClientFormat = "#{client_name}\t#{client_key_table}"

// TableFunc reports the key table that a tmux client is in.
type TableFunc func(client string) (string, error)

// SpawnTable looks the table up by running tmux for every call.
func SpawnTable(c Client) TableFunc {
	return func(client string) (string, error) {
		out, err := c.Run("list-clients", "-F", ClientFormat)
		if err != nil {
			return "", err
		}
		return TableOf(out, client)
	}
}

// TableOf finds client's key table in list-clients output, or fails.
func TableOf(out, client string) (string, error) {
	if t, ok := ClientTable(out, client); ok {
		return t, nil
	}
	return "", fmt.Errorf("tmux client %q not found", client)
}

// ClientTable finds client's key table in `list-clients -F
// "#{client_name}\t#{client_key_table}"` output.
func ClientTable(out, client string) (string, bool) {
	for _, line := range strings.Split(out, "\n") {
		if name, table, ok := strings.Cut(strings.TrimSpace(line), "\t"); ok && name == client {
			return table, true
		}
	}
	return "", false
}
