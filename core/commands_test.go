package core

import "testing"

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "ab", 1},
		{"abc", "ac", 1},
		{"kitten", "sitting", 3},
		{"こんにちは", "こにちは", 1}, // test unicode character
	}

	for _, tt := range tests {
		got := Levenshtein(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("Levenshtein(%q, %q) = %d; want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestGiveCommand(t *testing.T) {
	cmd := GiveCommand{
		Selector:   "@p",
		Item:       "minecraft:diamond_sword",
		Components: "damage=10",
		Amount:     "2",
	}

	want := "give @p minecraft:diamond_sword[damage=10] 2"
	got := cmd.Build()
	if got != want {
		t.Errorf("GiveCommand.Build() = %q; want %q", got, want)
	}
}

func TestTeleportCommand(t *testing.T) {
	cmd := TeleportCommand{
		Type:        "2",
		Targets:     "@a",
		Destination: "@s",
	}

	want := "teleport @a @s"
	got := cmd.Build()
	if got != want {
		t.Errorf("TeleportCommand.Build() = %q; want %q", got, want)
	}
}

func TestIsValidPosition(t *testing.T) {
	tests := []struct {
		pos  string
		want bool
	}{
		{"~ ~ ~", true},
		{"0 64 0", true},
		{"~-5 ~10 ~", true},
		{"^ ^ ^3", true},
		{"~ ~", false},
		{"~ ~ ~ ~", false},
		{"abc 64 0", false},
	}

	for _, tt := range tests {
		got := IsValidPosition(tt.pos)
		if got != tt.want {
			t.Errorf("IsValidPosition(%q) = %t; want %t", tt.pos, got, tt.want)
		}
	}
}

func TestSummonCommand(t *testing.T) {
	cmd := SummonCommand{
		Entity:   "zombie",
		Position: "~ ~5 ~",
		NBT:      "{IsBaby:1b}",
	}

	want := "summon minecraft:zombie ~ ~5 ~ {IsBaby:1b}"
	got := cmd.Build()
	if got != want {
		t.Errorf("SummonCommand.Build() = %q; want %q", got, want)
	}
}

