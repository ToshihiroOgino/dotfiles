package main

import "testing"

func TestNextState(t *testing.T) {
	tests := []struct {
		requested string
		status    string
		want      string
		changed   bool
	}{
		{"working", "0|0|1|", "working", true},
		{"waiting", "0|0|1|working", "waiting", true},
		{"waiting", "0|0|1|done", "waiting", false},
		{"done", "1|1|1|working", "seen", true},
		{"done", "1|1|2|working", "seen", true},
		{"done", "1|1|0|working", "done", true},
		{"done", "1|0|1|working", "done", true},
		{"seen", "1|1|1|done\r\n", "seen", true},
		{"seen", "1|1|1|working", "seen", false},
	}
	for _, tt := range tests {
		got, changed := nextState(tt.requested, parseStatus(tt.status))
		if got != tt.want || changed != tt.changed {
			t.Errorf("nextState(%q, %q) = %q, %v; want %q, %v",
				tt.requested, tt.status, got, changed, tt.want, tt.changed)
		}
	}
}
