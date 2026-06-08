package core

import (
	"testing"
)

type ParserTestCase struct {
	Name            string
	Input           string
	WantTokensCount int  // 分割されるべき正しいトークン数
	WantErrorIndex  int  // 期待されるエラー発生位置（エラーなしは -1）
	IsExecutable    bool // コマンドとして成立しているか
}

func TestCommandParserTableDriven(t *testing.T) {
	// データの読み込み
	if err := LoadData(); err != nil {
		t.Fatalf("Failed to load data: %v", err)
	}

	tests := []ParserTestCase{
		{
			Name:            "1. Normal give stone (Normal, Optional check)",
			Input:           "/give @p stone",
			WantTokensCount: 3,
			WantErrorIndex:  -1,
			IsExecutable:    true,
		},
		{
			Name:            "2. Summon with nested NBT quotes and spaces (Parentheses and Quotes protection)",
			Input:           "/summon zombie ~ ~ ~ {NoAI:1b, CustomName:'\"Hello World\"'}",
			WantTokensCount: 6,
			WantErrorIndex:  -1,
			IsExecutable:    true,
		},
		{
			Name:            "3. Direct component coupling give netherite_sword[damage=0] 1 (Coupling separation)",
			Input:           "/give @p netherite_sword[damage=0] 1",
			WantTokensCount: 5,
			WantErrorIndex:  -1,
			IsExecutable:    true,
		},
		{
			Name:            "4. Multi sub-command execute连鎖 (Infinite loop redirect check)",
			Input:           "/execute as @a at @s run say hello",
			WantTokensCount: 8,
			WantErrorIndex:  -1,
			IsExecutable:    true,
		},
		{
			Name:            "5. Execute chain failing mid-way (Error Index detection)",
			Input:           "/execute as @a at @s invalid_cmd",
			WantTokensCount: 6,
			WantErrorIndex:  5,
			IsExecutable:    false,
		},
		{
			Name:            "6. Case-insensitivity support for commands and selectors",
			Input:           "/GIVE @P STONE",
			WantTokensCount: 3,
			WantErrorIndex:  -1,
			IsExecutable:    true,
		},
		{
			Name:            "7. Coordinate incompleteness check (error on next argument)",
			Input:           "/summon zombie ~ ~ {NoAI:1b}",
			WantTokensCount: 5,
			WantErrorIndex:  4,
			IsExecutable:    false,
		},
		{
			Name:            "8. Non-executable incomplete command",
			Input:           "/give",
			WantTokensCount: 1,
			WantErrorIndex:  -1,
			IsExecutable:    false,
		},
		{
			Name:            "9. Extremely nested braces inside say",
			Input:           `/say {"text":"hello [world] 'nested' {\"a\":1}}"}`,
			WantTokensCount: 2,
			WantErrorIndex:  -1,
			IsExecutable:    true,
		},
		{
			Name:            "10. BlockState coupling setblock stone[snowy=true] destroy",
			Input:           "/setblock ~ ~ ~ stone[snowy=true] destroy",
			WantTokensCount: 7,
			WantErrorIndex:  -1,
			IsExecutable:    true,
		},
		{
			Name:            "11. Incomplete coordinate at the end of input (not an error yet)",
			Input:           "/summon zombie ~ ~",
			WantTokensCount: 4,
			WantErrorIndex:  -1,
			IsExecutable:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			tokens := Tokenize(tc.Input)
			if len(tokens) != tc.WantTokensCount {
				var txts []string
				for _, tok := range tokens {
					txts = append(txts, tok.Text)
				}
				t.Errorf("Tokenize(%q) count = %d; want %d. Tokens: %v", tc.Input, len(tokens), tc.WantTokensCount, txts)
			}

			res := ParseCommand(tc.Input)
			if res.ErrorIdx != tc.WantErrorIndex {
				t.Errorf("ParseCommand(%q) ErrorIdx = %d; want %d. Words: %+v", tc.Input, res.ErrorIdx, tc.WantErrorIndex, res.Words)
			}

			if res.IsExecutable != tc.IsExecutable {
				t.Errorf("ParseCommand(%q) IsExecutable = %t; want %t. ActiveNode: %+v", tc.Input, res.IsExecutable, tc.IsExecutable, res.CurrentNode)
			}
		})
	}
}
