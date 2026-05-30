package cmd

import "testing"

func TestMergeNbtStrings(t *testing.T) {
	tests := []struct {
		assembled string
		custom    string
		want      string
	}{
		{"", "", ""},
		{"{NoAI:1b}", "", "{NoAI:1b}"},
		{"", "Attributes:[{}]", "{Attributes:[{}]}"},
		{"", "{Attributes:[{}]}", "{Attributes:[{}]}"},
		{"{NoAI:1b,Silent:1b}", "Attributes:[{}]", "{NoAI:1b,Silent:1b,Attributes:[{}]}"},
		{"{NoAI:1b}", "{Attributes:[{}]}", "{NoAI:1b,Attributes:[{}]}"},
	}

	for _, tt := range tests {
		got := MergeNbtStrings(tt.assembled, tt.custom)
		if got != tt.want {
			t.Errorf("MergeNbtStrings(%q, %q) = %q; want %q", tt.assembled, tt.custom, got, tt.want)
		}
	}
}
