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

func TestAssembleNbtString(t *testing.T) {
	tests := []struct {
		keys []string
		want string
	}{
		{[]string{}, ""},
		{[]string{"NoAI"}, "{NoAI:1b}"},
		{[]string{"NoAI", "Silent"}, "{NoAI:1b,Silent:1b}"},
	}

	for _, tt := range tests {
		got := AssembleNbtString(tt.keys)
		if got != tt.want {
			t.Errorf("AssembleNbtString(%v) = %q; want %q", tt.keys, got, tt.want)
		}
	}
}

func TestParseCommandAndSyntaxGuides(t *testing.T) {
	if err := LoadData(); err != nil {
		t.Fatalf("failed to load data: %v", err)
	}

	// 1. ParseCommand "/give @p"
	res := ParseCommand("/give @p")
	if len(res.Words) != 2 {
		t.Fatalf("expected 2 words, got %d", len(res.Words))
	}
	if res.Words[0].Text != "/give" || !res.Words[0].IsLiteral {
		t.Errorf("expected '/give' to be literal, got: %+v", res.Words[0])
	}
	if res.Words[1].Text != "@p" || res.Words[1].IsLiteral {
		t.Errorf("expected '@p' to be dynamic argument, got: %+v", res.Words[1])
	}
	if res.CurrentNodeName != "targets" {
		t.Errorf("expected active node name to be 'targets', got %q", res.CurrentNodeName)
	}

	// 2. ParseCommand "/summon zombie ~ ~5 ~" (coordinate combination test)
	res2 := ParseCommand("/summon zombie ~ ~5 ~")
	if len(res2.Words) != 3 {
		t.Fatalf("expected 3 words (summon, zombie, coordinates), got %d: %+v", len(res2.Words), res2.Words)
	}
	if res2.Words[2].Text != "~ ~5 ~" {
		t.Errorf("expected '~ ~5 ~' to be combined coordinate word, got %q", res2.Words[2].Text)
	}

	// 3. GetSyntaxGuides for /give
	resGive := ParseCommand("/give")
	guides := GetSyntaxGuides(resGive.CurrentNode, "/give")
	if len(guides) == 0 {
		t.Error("expected syntax guides for /give, got 0")
	}

	// Verify one of the guides matches the minecraft give command pattern
	found := false
	for _, g := range guides {
		if g == "/give <targets> <item>" || g == "/give <targets> <item> <count>" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected standard give guide in results: %v", guides)
	}

	// 4. Item component suggestions test
	resComp := ParseCommand("/give @p diamond_sword[da")
	if resComp.CurrentNodeName != "components" {
		t.Errorf("expected active node name to be 'components', got %q", resComp.CurrentNodeName)
	}
	if resComp.CurrentParser != "minecraft:item_component" {
		t.Errorf("expected parser to be 'minecraft:item_component', got %q", resComp.CurrentParser)
	}
	foundDamage := false
	for _, s := range resComp.Suggestions {
		if s == "diamond_sword[minecraft:damage" {
			foundDamage = true
			break
		}
	}
	if !foundDamage {
		t.Errorf("expected to find 'diamond_sword[minecraft:damage' in suggestions for 'diamond_sword[da', got: %v", resComp.Suggestions)
	}

	// 5. Summon (Entity) suggestions test
	resSummon := ParseCommand("/summon zo")
	foundZombie := false
	for _, s := range resSummon.Suggestions {
		if s == "zombie" || s == "minecraft:zombie" {
			foundZombie = true
			break
		}
	}
	if !foundZombie {
		t.Errorf("expected to find 'zombie' or 'minecraft:zombie' in summon suggestions, got: %v", resSummon.Suggestions)
	}

	// 6. Effect suggestions test
	resEffect := ParseCommand("/effect give @p spe")
	foundSpeed := false
	for _, s := range resEffect.Suggestions {
		if s == "minecraft:speed" || s == "speed" {
			foundSpeed = true
			break
		}
	}
	if !foundSpeed {
		t.Errorf("expected to find 'minecraft:speed' or 'speed' in effect suggestions, got: %v", resEffect.Suggestions)
	}

	// 7. Coordinate incompleteness test
	resCoordErr := ParseCommand("/summon zombie ~ ~ {NoAI:1b}")
	foundCoordError := false
	for _, w := range resCoordErr.Words {
		if w.Text == "{NoAI:1b}" && w.IsError {
			foundCoordError = true
			break
		}
	}
	if !foundCoordError {
		t.Errorf("expected '{NoAI:1b}' to be marked as error due to incomplete coordinates, words: %+v", resCoordErr.Words)
	}

	// 8. Incomplete coordinate at the end of input should NOT be an error
	resCoordOk := ParseCommand("/summon zombie ~ ~")
	if resCoordOk.ErrorIdx != -1 {
		t.Errorf("expected no error for '/summon zombie ~ ~' (user still typing), got error at %d, words: %+v", resCoordOk.ErrorIdx, resCoordOk.Words)
	}
}



