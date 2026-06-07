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

func TestGetApplicableOptionsSelect(t *testing.T) {
	// 1. カエル (frog) の場合の variant テスト
	frogOpts := GetApplicableOptions("frog")
	var frogVariantFound bool
	for _, opt := range frogOpts {
		if opt.Key == "variant" {
			frogVariantFound = true
			if opt.Type != "select" {
				t.Errorf("frog variant Type = %q; want 'select'", opt.Type)
			}
			if len(opt.Choices) != 3 {
				t.Errorf("frog variant Choices count = %d; want 3", len(opt.Choices))
			}
		}
	}
	if !frogVariantFound {
		t.Error("frog variant NBT option not found")
	}

	// 2. オオカミ (wolf) の場合の CollarColor テスト
	wolfOpts := GetApplicableOptions("wolf")
	var collarColorFound bool
	for _, opt := range wolfOpts {
		if opt.Key == "CollarColor" {
			collarColorFound = true
			if opt.Type != "select" {
				t.Errorf("wolf CollarColor Type = %q; want 'select'", opt.Type)
			}
			if len(opt.Choices) != 16 {
				t.Errorf("wolf CollarColor Choices count = %d; want 16", len(opt.Choices))
			}
		}
	}
	if !collarColorFound {
		t.Error("wolf CollarColor NBT option not found")
	}
}
