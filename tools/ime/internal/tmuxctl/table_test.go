package tmuxctl

import "testing"

func TestClientTable(t *testing.T) {
	out := "/dev/ttys001\troot\n/dev/ttys004\tprefix\n\n"
	cases := []struct {
		client string
		table  string
		ok     bool
	}{
		{"/dev/ttys001", "root", true},
		{"/dev/ttys004", "prefix", true},
		{"/dev/ttys009", "", false},
	}
	for _, c := range cases {
		table, ok := ClientTable(out, c.client)
		if table != c.table || ok != c.ok {
			t.Errorf("ClientTable(%q) = %q, %v; want %q, %v", c.client, table, ok, c.table, c.ok)
		}
	}
}
