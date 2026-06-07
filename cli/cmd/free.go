package cmd

import (
	"cmdforge/core"
	"fmt"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// WordStyle represents the styling attributes for a single input word.
type WordStyle struct {
	Text      string
	IsLiteral bool
	ArgIndex  int
	IsError   bool
}

// styledChar represents a single rune and its LipGloss style.
type styledChar struct {
	char  rune
	style lipgloss.Style
}

var parserTypeNames = map[string]string{
	"minecraft:entity":            "対象（プレイヤー名や @p, @a など）",
	"minecraft:game_profile":      "ゲームプロフィール（プレイヤー名）",
	"minecraft:score_holder":      "スコアホルダー（プレイヤー名または対象エンティティ）",
	"minecraft:item_stack":        "アイテムID (例: minecraft:diamond)",
	"minecraft:item_parser":       "アイテム (例: minecraft:diamond)",
	"minecraft:item_enchantment":  "エンチャントするアイテム",
	"minecraft:item_slot":         "インベントリスロット (例: container.0)",
	"minecraft:entity_summon":     "召喚するエンティティID (例: zombie)",
	"minecraft:mob_effect":        "ステータス効果ID (例: speed)",
	"minecraft:enchantment":       "エンチャント効果ID (例: sharpness)",
	"minecraft:block_pos":         "ブロック座標 (X Y Z)",
	"minecraft:vec3":              "3次元座標 (X Y Z)",
	"minecraft:vec2":              "2次元座標 (X Z)",
	"minecraft:column_pos":        "列座標 (X Z)",
	"minecraft:block_state":       "ブロックIDと状態 (例: air, stone)",
	"minecraft:block_input":       "ブロックIDと状態",
	"minecraft:resource_key":      "リソースキー ID",
	"minecraft:resource":          "リソース ID",
	"minecraft:resource_location": "リソースの場所 / パス",
	"minecraft:dimension":         "ディメンション名 (例: minecraft:overworld)",
	"minecraft:nbt_compound_tag":  "NBTデータ (例: {CustomName:'\"Zombie\"'})",
	"minecraft:message":           "メッセージテキスト",
	"minecraft:component":         "JSON テキストコンポーネント",
	"minecraft:item_predicate":    "アイテム判定条件",
	"minecraft:block_predicate":   "ブロック判定条件",
	"brigadier:string":            "文字列",
	"brigadier:integer":           "整数値",
	"brigadier:double":            "浮動小数点数値",
	"brigadier:bool":              "真偽値 (true / false)",
}

// getDynamicSuggestions returns dynamic autocomplete list based on brigadier parser types.
func getDynamicSuggestions(parser string, registry string) []string {
	switch parser {
	case "minecraft:entity", "minecraft:game_profile", "minecraft:score_holder":
		return []string{"@p", "@a", "@r", "@s", "@e"}
	case "minecraft:item_stack", "minecraft:item_parser", "minecraft:item_enchantment", "minecraft:item_slot":
		return core.Items
	case "minecraft:entity_summon":
		return core.Entities
	case "minecraft:mob_effect":
		return core.Effects
	case "minecraft:enchantment":
		var enchNames []string
		for _, e := range core.Enchantments {
			enchNames = append(enchNames, e.Name)
		}
		return enchNames
	case "minecraft:block_pos", "minecraft:vec3", "minecraft:vec2", "minecraft:column_pos":
		return []string{"~ ~ ~"}
	case "minecraft:block_state", "minecraft:block_input":
		return core.Blocks
	case "minecraft:resource_key", "minecraft:resource":
		if registry == "minecraft:attribute" {
			return []string{"generic.max_health", "generic.movement_speed", "generic.attack_damage"}
		}
		if registry == "minecraft:dimension" {
			return []string{"minecraft:overworld", "minecraft:the_nether", "minecraft:the_end"}
		}
	}
	return nil
}

// combineCoordinates joins consecutive coordinate tokens into a single word token.
func combineCoordinates(rawWords []string) []string {
	var result []string
	n := len(rawWords)
	for i := 0; i < n; {
		w := rawWords[i]
		if w == "" {
			result = append(result, w)
			i++
			continue
		}

		if core.IsValidPositionToken(w) {
			posParts := []string{w}
			idx := i + 1
			for idx < n && len(posParts) < 3 {
				nextW := rawWords[idx]
				if nextW != "" && core.IsValidPositionToken(nextW) {
					posParts = append(posParts, nextW)
					idx++
				} else {
					break
				}
			}
			result = append(result, strings.Join(posParts, " "))
			i = idx
		} else {
			result = append(result, w)
			i++
		}
	}
	return result
}

// isValidArgValue checks if a given word is valid for a brigadier parser.
func isValidArgValue(w string, parser string, registry string, isLast bool) bool {
	// 座標系のパース
	if parser == "minecraft:block_pos" || parser == "minecraft:vec3" || parser == "minecraft:vec2" || parser == "minecraft:column_pos" {
		parts := strings.Split(w, " ")
		for _, part := range parts {
			if part == "" {
				continue
			}
			if !core.IsValidPositionToken(part) {
				return false
			}
		}
		return true
	}

	// 候補リストがある場合
	suggestions := getDynamicSuggestions(parser, registry)
	if suggestions != nil {
		for _, s := range suggestions {
			cleanS := strings.TrimPrefix(s, "minecraft:")
			if isLast {
				// 入力中の最後の単語の場合は前方一致でOK
				if strings.HasPrefix(s, w) || strings.HasPrefix(cleanS, w) {
					return true
				}
			} else {
				// 確定した単語の場合は完全一致が必要
				if s == w || cleanS == w {
					return true
				}
			}
		}
		return false
	}

	// 数値の簡易バリデーション
	if parser == "brigadier:integer" || parser == "brigadier:double" {
		if isLast {
			return true
		}
		if parser == "brigadier:integer" {
			var val int
			_, err := fmt.Sscanf(w, "%d", &val)
			return err == nil
		} else {
			var val float64
			_, err := fmt.Sscanf(w, "%f", &val)
			return err == nil
		}
	}
	if parser == "brigadier:bool" {
		if isLast {
			return strings.HasPrefix("true", w) || strings.HasPrefix("false", w)
		}
		return w == "true" || w == "false"
	}

	return true
}

// getHighlightAndSuggestions parses input string against command tree to produce word styles, suggestions, and next syntax preview.
func getHighlightAndSuggestions(input string) ([]WordStyle, []string, string, int, string) {
	// 先頭の "/" を除去してパース開始する
	cleanInput := strings.TrimPrefix(input, "/")
	rawWords := combineCoordinates(strings.Split(cleanInput, " "))
	var styles []WordStyle
	currentNode := &core.CommandTree
	argIdx := 0
	errorIdx := -1
	var currentParser string

	// 各確定単語に対してツリーをたどる
	for i, w := range rawWords {
		isLast := i == len(rawWords)-1
		if isLast {
			break
		}
		if w == "" {
			continue
		}

		if errorIdx != -1 {
			dispText := w
			if i == 0 && strings.HasPrefix(input, "/") {
				dispText = "/" + w
			}
			styles = append(styles, WordStyle{
				Text:      dispText,
				IsLiteral: false,
				ArgIndex:  argIdx,
				IsError:   true,
			})
			continue
		}

		var matched *core.BrigadierNode
		if currentNode.Children != nil {
			if child, exists := currentNode.Children[w]; exists && child.Type == "literal" {
				matched = child
			} else {
				// リテラルとして完全一致しない場合は、引数 (type == "argument") のノードを探す
				for _, child := range currentNode.Children {
					if child.Type == "argument" {
						if isValidArgValue(w, child.Parser, child.GetRegistry(), false) {
							matched = child
							break
						}
					}
				}
			}
		}

		dispText := w
		if i == 0 && strings.HasPrefix(input, "/") {
			dispText = "/" + w
		}

		if matched != nil {
			isLiteral := matched.Type == "literal"
			styles = append(styles, WordStyle{
				Text:      dispText,
				IsLiteral: isLiteral,
				ArgIndex:  argIdx,
				IsError:   false,
			})
			if !isLiteral {
				argIdx++
			}
			currentNode = matched
		} else {
			errorIdx = i
			styles = append(styles, WordStyle{
				Text:      dispText,
				IsLiteral: false,
				ArgIndex:  argIdx,
				IsError:   true,
			})
		}
	}

	lastWord := rawWords[len(rawWords)-1]
	var suggestions []string
	var syntaxParts []string

	isFirstWord := len(rawWords) == 1
	hasSlash := strings.HasPrefix(input, "/")

	if errorIdx != -1 {
		dispText := lastWord
		if len(rawWords) == 1 && hasSlash {
			dispText = "/" + lastWord
		}
		styles = append(styles, WordStyle{
			Text:      dispText,
			IsLiteral: false,
			ArgIndex:  argIdx,
			IsError:   true,
		})
		return styles, nil, "", errorIdx, ""
	}

	var matchedLast *core.BrigadierNode
	if currentNode.Children != nil {
		if child, exists := currentNode.Children[lastWord]; exists && child.Type == "literal" {
			matchedLast = child
		} else {
			for _, child := range currentNode.Children {
				if child.Type == "argument" {
					if isValidArgValue(lastWord, child.Parser, child.GetRegistry(), true) {
						matchedLast = child
						break
					}
				}
			}
		}
	}

	if currentNode.Children != nil {
		var keys []string
		for k := range currentNode.Children {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			child := currentNode.Children[key]
			if child.Type == "literal" {
				if strings.HasPrefix(key, lastWord) {
					suggestVal := key
					if isFirstWord && hasSlash {
						suggestVal = "/" + key
					}
					suggestions = append(suggestions, suggestVal)
				}
				syntaxParts = append(syntaxParts, key)
			} else if child.Type == "argument" {
				list := getDynamicSuggestions(child.Parser, child.GetRegistry())
				var sortedList []string
				if list != nil {
					sortedList = make([]string, len(list))
					copy(sortedList, list)
					sort.Strings(sortedList)
				}
				for _, item := range sortedList {
					cleanItem := strings.TrimPrefix(item, "minecraft:")
					if strings.HasPrefix(item, lastWord) || strings.HasPrefix(cleanItem, lastWord) {
						suggestions = append(suggestions, item)
					}
				}
				syntaxParts = append(syntaxParts, "<"+key+">")
			}
		}
	}

	if matchedLast != nil && matchedLast.Type == "argument" {
		currentParser = matchedLast.Parser
	}

	lastIsError := false
	if lastWord != "" {
		hasValidMatch := false
		if matchedLast != nil {
			hasValidMatch = true
		} else if currentNode.Children != nil {
			for key, child := range currentNode.Children {
				if child.Type == "literal" && strings.HasPrefix(key, lastWord) {
					hasValidMatch = true
					break
				}
			}
		}
		if !hasValidMatch {
			lastIsError = true
			errorIdx = len(rawWords) - 1
		}
	}

	lastIsLiteral := false
	if matchedLast != nil && matchedLast.Type == "literal" {
		lastIsLiteral = true
	}

	if lastWord != "" {
		dispText := lastWord
		if len(rawWords) == 1 && hasSlash {
			dispText = "/" + lastWord
		}
		styles = append(styles, WordStyle{
			Text:      dispText,
			IsLiteral: lastIsLiteral,
			ArgIndex:  argIdx,
			IsError:   lastIsError,
		})
	}

	var nextSyntaxPreview string
	if len(syntaxParts) > 0 {
		nextSyntaxPreview = strings.Join(syntaxParts, " | ")
	}

	return styles, suggestions, nextSyntaxPreview, errorIdx, currentParser
}

// getStyledRunes parses runes input and maps each char to styledChar.
func getStyledRunes(input []rune) []styledChar {
	str := string(input)
	rawWords := combineCoordinates(strings.Split(str, " "))
	wordStyles, _, _, errorIdx, _ := getHighlightAndSuggestions(str)

	whiteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	cyanStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#55FFFF"))
	yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF55"))
	limeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#55FF55"))
	pinkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF55FF"))
	orangeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA00"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Underline(true)

	getArgStyle := func(argIdx int) lipgloss.Style {
		switch argIdx % 5 {
		case 0:
			return cyanStyle
		case 1:
			return yellowStyle
		case 2:
			return limeStyle
		case 3:
			return pinkStyle
		case 4:
			return orangeStyle
		default:
			return whiteStyle
		}
	}

	var styledChars []styledChar
	wordIdx := 0
	runeIdx := 0

	for _, word := range rawWords {
		var currentStyle lipgloss.Style
		isErrorWord := errorIdx != -1 && wordIdx >= errorIdx

		if isErrorWord {
			currentStyle = errorStyle
		} else if wordIdx < len(wordStyles) {
			ws := wordStyles[wordIdx]
			if ws.IsError {
				currentStyle = errorStyle
			} else if ws.IsLiteral {
				currentStyle = whiteStyle
			} else {
				currentStyle = getArgStyle(ws.ArgIndex)
			}
		} else {
			currentStyle = whiteStyle
		}

		for _, r := range word {
			styledChars = append(styledChars, styledChar{char: r, style: currentStyle})
			runeIdx++
		}

		// Keep spacing exactly matching input runes
		for runeIdx < len(input) && input[runeIdx] == ' ' {
			spaceStyle := whiteStyle
			if errorIdx != -1 && wordIdx >= errorIdx {
				spaceStyle = errorStyle
			}
			styledChars = append(styledChars, styledChar{char: ' ', style: spaceStyle})
			runeIdx++
		}
		wordIdx++
	}

	return styledChars
}

// getSyntaxGuide returns command tree syntax string based on input.
func getSyntaxGuide(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return "コマンドを入力してください (例: /give, /summon, /item):\n" +
			"  /give <targets> <item> [<count>]\n" +
			"  /summon <entity> [<pos>] [<nbt>]\n" +
			"  /item replace block <pos> <slot> with <item> [<count>]\n" +
			"  /item replace entity <targets> <slot> with <item> [<count>]\n" +
			"  /item modify block <pos> <slot> <modifier>\n" +
			"  /item modify entity <targets> <slot> <modifier>"
	}

	parts := strings.Split(input, " ")
	cmd := parts[0]

	if strings.HasPrefix("/give", cmd) {
		return "構文ガイド:\n  /give <targets> <item> [<count>]"
	}
	if strings.HasPrefix("/summon", cmd) {
		return "構文ガイド:\n  /summon <entity> [<pos>] [<nbt>]"
	}
	if strings.HasPrefix("/item", cmd) {
		return "構文ガイド:\n" +
			"  /item replace block <pos> <slot> with <item> [<count>]\n" +
			"  /item replace entity <targets> <slot> with <item> [<count>]\n" +
			"  /item modify block <pos> <slot> <modifier>\n" +
			"  /item modify entity <targets> <slot> <modifier>"
	}

	return "不明なコマンドです。以下のいずれかを入力してください:\n" +
		"  /give | /summon | /item"
}

// bubbletea model
type freeModel struct {
	input           []rune
	cursorIdx       int
	suggestions     []string
	selectedSuggest int
	width, height   int
	quitting        bool
	currentParser   string
}

func (m freeModel) Init() tea.Cmd {
	return nil
}

func (m *freeModel) updateSuggestions() {
	_, suggestions, _, _, currentParser := getHighlightAndSuggestions(string(m.input))
	m.suggestions = suggestions
	m.currentParser = currentParser
	if m.selectedSuggest >= len(m.suggestions) {
		m.selectedSuggest = 0
	}
}

func (m *freeModel) autocomplete() {
	if len(m.suggestions) == 0 {
		return
	}
	selected := m.suggestions[m.selectedSuggest]
	str := string(m.input)

	// Replace the last word before the cursor
	leftPart := str[:m.cursorIdx]
	rightPart := str[m.cursorIdx:]

	lastSpaceIdx := strings.LastIndex(leftPart, " ")
	var newLeft string
	if lastSpaceIdx == -1 {
		newLeft = selected + " "
	} else {
		newLeft = leftPart[:lastSpaceIdx+1] + selected + " "
	}

	m.input = []rune(newLeft + rightPart)
	m.cursorIdx = len([]rune(newLeft))
	m.selectedSuggest = 0
	m.updateSuggestions()
}

func (m freeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEnter:
			cmdStr := string(m.input)
			if strings.TrimSpace(cmdStr) != "" {
				_ = clipboard.WriteAll(cmdStr)
				m.quitting = true
				return m, tea.Quit
			}

		case tea.KeyUp, tea.KeyCtrlP:
			if len(m.suggestions) > 0 {
				m.selectedSuggest--
				if m.selectedSuggest < 0 {
					m.selectedSuggest = len(m.suggestions) - 1
				}
			}

		case tea.KeyDown, tea.KeyCtrlN:
			if len(m.suggestions) > 0 {
				m.selectedSuggest++
				if m.selectedSuggest >= len(m.suggestions) {
					m.selectedSuggest = 0
				}
			}

		case tea.KeyTab:
			m.autocomplete()

		case tea.KeyLeft:
			if m.cursorIdx > 0 {
				m.cursorIdx--
			}

		case tea.KeyRight:
			if m.cursorIdx < len(m.input) {
				m.cursorIdx++
			}

		case tea.KeyBackspace:
			if m.cursorIdx > 0 {
				m.input = append(m.input[:m.cursorIdx-1], m.input[m.cursorIdx:]...)
				m.cursorIdx--
				m.selectedSuggest = 0
				m.updateSuggestions()
			}

		case tea.KeySpace:
			m.input = append(m.input[:m.cursorIdx], append([]rune{' '}, m.input[m.cursorIdx:]...)...)
			m.cursorIdx++
			m.selectedSuggest = 0
			m.updateSuggestions()

		case tea.KeyRunes:
			m.input = append(m.input[:m.cursorIdx], append(msg.Runes, m.input[m.cursorIdx:]...)...)
			m.cursorIdx += len(msg.Runes)
			m.selectedSuggest = 0
			m.updateSuggestions()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m freeModel) renderInputLine() string {
	styled := getStyledRunes(m.input)
	cursorStyle := lipgloss.NewStyle().Background(lipgloss.Color("#FFFFFF")).Foreground(lipgloss.Color("#000000"))

	var sb strings.Builder
	sb.WriteString(" > ")

	for i := 0; i < len(m.input); i++ {
		if i == m.cursorIdx {
			sb.WriteString(cursorStyle.Render(string(m.input[i])))
		} else {
			sb.WriteString(styled[i].style.Render(string(styled[i].char)))
		}
	}

	if m.cursorIdx == len(m.input) {
		sb.WriteString(cursorStyle.Render(" "))
	}

	return sb.String()
}

func (m freeModel) renderSuggestArea() string {
	if len(m.suggestions) == 0 {
		return ""
	}

	highlightStyle := lipgloss.NewStyle().Background(lipgloss.Color("#55FFFF")).Foreground(lipgloss.Color("#000000"))
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))

	maxVisible := 10
	start := 0
	if m.selectedSuggest >= maxVisible {
		start = m.selectedSuggest - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(m.suggestions) {
		end = len(m.suggestions)
		start = end - maxVisible
		if start < 0 {
			start = 0
		}
	}

	var lines []string
	for i := start; i < end; i++ {
		s := m.suggestions[i]
		if i == m.selectedSuggest {
			lines = append(lines, highlightStyle.Render("  "+s+"  "))
		} else {
			lines = append(lines, normalStyle.Render("  "+s))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m freeModel) renderTopArea(height int) string {
	if height <= 0 {
		return ""
	}

	helpText := "操作ヘルプ:\n" +
		"  [Tab]   : 選択中の候補を補完\n" +
		"  [Esc]   : 終了\n" +
		"  [Up/Down/Ctrl+N/Ctrl+P] : 候補の選択切り替え\n" +
		"  [Enter] : コマンド生成を実行"

	guide := getSyntaxGuide(string(m.input))

	var typeGuidance string
	if m.currentParser != "" {
		displayName := m.currentParser
		if jpName, exists := parserTypeNames[m.currentParser]; exists {
			displayName = jpName
		}
		yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF55"))
		typeGuidance = "\n" + yellowStyle.Render("型: "+displayName)
	}

	content := guide + typeGuidance + "\n\n" + helpText

	return lipgloss.NewStyle().
		Height(height).
		AlignVertical(lipgloss.Bottom).
		Render(content)
}

func (m freeModel) View() string {
	if m.quitting {
		return "終了します。\n"
	}

	suggestStr := m.renderSuggestArea()
	suggestLines := 0
	if suggestStr != "" {
		suggestLines = lipgloss.Height(suggestStr)
	}

	topHeight := m.height - suggestLines - 1
	if topHeight < 0 {
		topHeight = 0
	}

	topStr := m.renderTopArea(topHeight)
	inputStr := m.renderInputLine()

	var elements []string
	if topStr != "" {
		elements = append(elements, topStr)
	}
	if suggestStr != "" {
		elements = append(elements, suggestStr)
	}
	elements = append(elements, inputStr)

	return lipgloss.JoinVertical(lipgloss.Left, elements...)
}

// freeCmd represents the free command.
var freeCmd = &cobra.Command{
	Use:   "free",
	Short: "Minecraft chat playground mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := core.LoadData(); err != nil {
			return err
		}

		initialModel := freeModel{
			input:     []rune("/"),
			cursorIdx: 1,
			width:     80,
			height:    24,
		}
		initialModel.updateSuggestions()

		p := tea.NewProgram(initialModel, tea.WithAltScreen())
		finalModel, err := p.Run()
		if err != nil {
			return err
		}

		m := finalModel.(freeModel)
		cmdStr := string(m.input)
		if strings.TrimSpace(cmdStr) != "" {
			fmt.Printf("\nGenerated Command (copied to clipboard):\n  %s\n", cmdStr)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(freeCmd)
}
