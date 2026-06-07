package cmd

import (
	"cmdforge/core"
	"testing"
)

func TestGetHighlightAndSuggestionsGive(t *testing.T) {
	_ = core.LoadData() // Ensure data is loaded

	// Test case 1: Starting to type /give (incomplete)
	styles, suggestions, preview := getHighlightAndSuggestions("/gi")
	if len(styles) != 1 || styles[0].Text != "/gi" || styles[0].IsLiteral {
		t.Errorf("expected styles for '/gi' to be non-literal, got: %+v", styles)
	}
	var foundGive bool
	for _, s := range suggestions {
		if s == "/give" {
			foundGive = true
			break
		}
	}
	if !foundGive {
		t.Error("expected '/give' to be in suggestions for '/gi'")
	}
	if preview == "" {
		t.Error("expected syntax preview but got empty string")
	}

	// Test case 1b: Completed command name (should be literal)
	styles, _, _ = getHighlightAndSuggestions("/give")
	if len(styles) != 1 || styles[0].Text != "/give" || !styles[0].IsLiteral {
		t.Errorf("expected completed '/give' to be literal, got: %+v", styles)
	}

	// Test case 2: /give @p 
	// Trailing space means we are waiting for the next argument (item)
	styles, suggestions, _ = getHighlightAndSuggestions("/give @p ")
	if len(styles) != 2 {
		t.Fatalf("expected 2 styles, got %d", len(styles))
	}
	if styles[0].Text != "/give" || !styles[0].IsLiteral {
		t.Errorf("expected '/give' to be literal, got %+v", styles[0])
	}
	if styles[1].Text != "@p" || styles[1].IsLiteral || styles[1].ArgIndex != 0 {
		t.Errorf("expected '@p' to be dynamic arg 0, got %+v", styles[1])
	}
	// Suggestions should contain items
	if len(suggestions) == 0 {
		t.Error("expected item suggestions, got 0")
	}
}

func TestGetHighlightAndSuggestionsItem(t *testing.T) {
	_ = core.LoadData()

	// Test case 1: /item replace block ~ 
	styles, _, _ := getHighlightAndSuggestions("/item replace block ~ ")
	if len(styles) != 4 {
		t.Fatalf("expected 4 styles, got %d: %+v", len(styles), styles)
	}

	// /item (literal)
	if styles[0].Text != "/item" || !styles[0].IsLiteral {
		t.Errorf("expected '/item' to be literal, got: %+v", styles[0])
	}
	// replace (literal)
	if styles[1].Text != "replace" || !styles[1].IsLiteral {
		t.Errorf("expected 'replace' to be literal, got: %+v", styles[1])
	}
	// block (literal)
	if styles[2].Text != "block" || !styles[2].IsLiteral {
		t.Errorf("expected 'block' to be literal, got: %+v", styles[2])
	}
	// ~ (dynamic argument 0)
	if styles[3].Text != "~" || styles[3].IsLiteral || styles[3].ArgIndex != 0 {
		t.Errorf("expected '~' to be dynamic arg 0, got: %+v", styles[3])
	}
}

func TestCombineCoordinates(t *testing.T) {
	_ = core.LoadData()

	// Test 1: Coordinate combining
	styles, _, _ := getHighlightAndSuggestions("/summon zombie ~ ~5 ~ ")
	if len(styles) != 3 {
		t.Fatalf("expected 3 styles (command, entity, position), got %d: %+v", len(styles), styles)
	}

	if styles[0].Text != "/summon" || !styles[0].IsLiteral {
		t.Errorf("expected '/summon' to be literal, got: %+v", styles[0])
	}
	if styles[1].Text != "zombie" || styles[1].IsLiteral || styles[1].ArgIndex != 0 {
		t.Errorf("expected 'zombie' to be dynamic arg 0, got: %+v", styles[1])
	}
	if styles[2].Text != "~ ~5 ~" || styles[2].IsLiteral || styles[2].ArgIndex != 1 {
		t.Errorf("expected '~ ~5 ~' to be dynamic arg 1, got: %+v", styles[2])
	}
}

func TestPrefixlessCompletion(t *testing.T) {
	_ = core.LoadData()

	// "/give @p diam" のようにプレフィックスなしで打ったとき
	_, suggestions, _ := getHighlightAndSuggestions("/give @p diam")
	
	// サジェストに "minecraft:diamond" または "minecraft:diamond_sword" が含まれているか
	var foundDiamond bool
	for _, s := range suggestions {
		if s == "minecraft:diamond" || s == "minecraft:diamond_sword" {
			foundDiamond = true
			break
		}
	}
	if !foundDiamond {
		t.Errorf("expected to find 'minecraft:diamond' or 'minecraft:diamond_sword' in suggestions for 'diam', got: %v", suggestions)
	}
}

func TestSuggestionSorting(t *testing.T) {
	_ = core.LoadData()

	// "/g" に対するサジェストを抽出
	_, suggestions, _ := getHighlightAndSuggestions("/g")
	if len(suggestions) < 2 {
		t.Skip("not enough suggestions to verify sort order")
	}

	// アルファベット順（a-z）にソートされていることを確認
	for i := 0; i < len(suggestions)-1; i++ {
		if suggestions[i] > suggestions[i+1] {
			t.Errorf("suggestions are not sorted: %q is before %q", suggestions[i], suggestions[i+1])
		}
	}
}
