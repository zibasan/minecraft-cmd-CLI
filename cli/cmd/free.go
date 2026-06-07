package cmd

import (
	"cmdforge/core"
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// CommandNode represents a node in the Minecraft command syntax tree.
type CommandNode struct {
	Name        string
	IsLiteral   bool
	Children    []CommandNode
	Description string
	DynamicList func() []string
}

// WordStyle represents the styling attributes for a single input word.
type WordStyle struct {
	Text      string
	IsLiteral bool
	ArgIndex  int
}

// styledChar represents a single rune and its LipGloss style.
type styledChar struct {
	char  rune
	style lipgloss.Style
}

// 簡易コマンドツリー定義
var commandTree = CommandNode{
	Name:      "root",
	IsLiteral: true,
	Children: []CommandNode{
		// 1. /give <targets> <item> [<count>]
		{
			Name:      "/give",
			IsLiteral: true,
			Children: []CommandNode{
				{
					Name:      "<targets>",
					IsLiteral: false,
					DynamicList: func() []string {
						return []string{"@p", "@a", "@r", "@s", "@e"}
					},
					Children: []CommandNode{
						{
							Name:      "<item>",
							IsLiteral: false,
							DynamicList: func() []string {
								return core.Items
							},
							Children: []CommandNode{
								{
									Name:      "[<count>]",
									IsLiteral: false,
									DynamicList: func() []string {
										return []string{"1", "64"}
									},
								},
							},
						},
					},
				},
			},
		},
		// 2. /summon <entity> [<pos>] [<nbt>]
		{
			Name:      "/summon",
			IsLiteral: true,
			Children: []CommandNode{
				{
					Name:      "<entity>",
					IsLiteral: false,
					DynamicList: func() []string {
						return core.Entities
					},
					Children: []CommandNode{
						{
							Name:      "[<pos>]",
							IsLiteral: false,
							DynamicList: func() []string {
								return []string{"~ ~ ~"}
							},
							Children: []CommandNode{
								{
									Name:      "[<nbt>]",
									IsLiteral: false,
									DynamicList: func() []string {
										return []string{"{}", "{NoAI:1b}", "{Silent:1b}"}
									},
								},
							},
						},
					},
				},
			},
		},
		// 3. /item <replace|modify>
		{
			Name:      "/item",
			IsLiteral: true,
			Children: []CommandNode{
				// /item replace
				{
					Name:      "replace",
					IsLiteral: true,
					Children: []CommandNode{
						// block <pos> <slot> with <item> [<count>]
						{
							Name:      "block",
							IsLiteral: true,
							Children: []CommandNode{
								{
									Name:      "<pos>",
									IsLiteral: false,
									DynamicList: func() []string {
										return []string{"~ ~ ~"}
									},
									Children: []CommandNode{
										{
											Name:      "<slot>",
											IsLiteral: false,
											DynamicList: func() []string {
												return []string{"container.0", "weapon.mainhand", "armor.chest"}
											},
											Children: []CommandNode{
												{
													Name:      "with",
													IsLiteral: true,
													Children: []CommandNode{
														{
															Name:      "<item>",
															IsLiteral: false,
															DynamicList: func() []string {
																return core.Items
															},
															Children: []CommandNode{
																{
																	Name:      "[<count>]",
																	IsLiteral: false,
																	DynamicList: func() []string {
																		return []string{"1", "64"}
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						// entity <targets> <slot> with <item> [<count>]
						{
							Name:      "entity",
							IsLiteral: true,
							Children: []CommandNode{
								{
									Name:      "<targets>",
									IsLiteral: false,
									DynamicList: func() []string {
										return []string{"@p", "@a", "@r", "@s", "@e"}
									},
									Children: []CommandNode{
										{
											Name:      "<slot>",
											IsLiteral: false,
											DynamicList: func() []string {
												return []string{"weapon.mainhand", "armor.chest", "armor.head"}
											},
											Children: []CommandNode{
												{
													Name:      "with",
													IsLiteral: true,
													Children: []CommandNode{
														{
															Name:      "<item>",
															IsLiteral: false,
															DynamicList: func() []string {
																return core.Items
															},
															Children: []CommandNode{
																{
																	Name:      "[<count>]",
																	IsLiteral: false,
																	DynamicList: func() []string {
																		return []string{"1", "64"}
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				// /item modify
				{
					Name:      "modify",
					IsLiteral: true,
					Children: []CommandNode{
						// block <pos> <slot> <modifier>
						{
							Name:      "block",
							IsLiteral: true,
							Children: []CommandNode{
								{
									Name:      "<pos>",
									IsLiteral: false,
									DynamicList: func() []string {
										return []string{"~ ~ ~"}
									},
									Children: []CommandNode{
										{
											Name:      "<slot>",
											IsLiteral: false,
											DynamicList: func() []string {
												return []string{"container.0", "weapon.mainhand"}
											},
											Children: []CommandNode{
												{
													Name:      "<modifier>",
													IsLiteral: false,
													DynamicList: func() []string {
														return []string{"example_modifier"}
													},
												},
											},
										},
									},
								},
							},
						},
						// entity <targets> <slot> <modifier>
						{
							Name:      "entity",
							IsLiteral: true,
							Children: []CommandNode{
								{
									Name:      "<targets>",
									IsLiteral: false,
									DynamicList: func() []string {
										return []string{"@p", "@a", "@r", "@s", "@e"}
									},
									Children: []CommandNode{
										{
											Name:      "<slot>",
											IsLiteral: false,
											DynamicList: func() []string {
												return []string{"weapon.mainhand", "armor.chest"}
											},
											Children: []CommandNode{
												{
													Name:      "<modifier>",
													IsLiteral: false,
													DynamicList: func() []string {
														return []string{"example_modifier"}
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	},
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

// getHighlightAndSuggestions parses input string against command tree to produce word styles, suggestions, and next syntax preview.
func getHighlightAndSuggestions(input string) ([]WordStyle, []string, string) {
	rawWords := combineCoordinates(strings.Split(input, " "))
	var styles []WordStyle
	currentNode := &commandTree
	argIdx := 0

	// Trailing empty word indicates user wants to see next suggestions
	for i, w := range rawWords {
		isLast := i == len(rawWords)-1
		if isLast {
			break
		}
		if w == "" {
			continue
		}

		var matched *CommandNode
		for _, child := range currentNode.Children {
			if child.IsLiteral {
				if child.Name == w {
					matched = &child
					break
				}
			} else {
				// dynamic placeholders match anything that doesn't match literal sibling
				matched = &child
				break
			}
		}

		if matched != nil {
			styles = append(styles, WordStyle{
				Text:      w,
				IsLiteral: matched.IsLiteral,
				ArgIndex:  argIdx,
			})
			if !matched.IsLiteral {
				argIdx++
			}
			currentNode = matched
		} else {
			styles = append(styles, WordStyle{
				Text:      w,
				IsLiteral: false,
				ArgIndex:  argIdx,
			})
			argIdx++
		}
	}

	lastWord := rawWords[len(rawWords)-1]
	var suggestions []string
	var nextSyntaxPreview string
	var syntaxParts []string

	for _, child := range currentNode.Children {
		if child.IsLiteral {
			if strings.HasPrefix(child.Name, lastWord) {
				suggestions = append(suggestions, child.Name)
			}
		} else {
			if child.DynamicList != nil {
				list := child.DynamicList()
				for _, item := range list {
					cleanItem := strings.TrimPrefix(item, "minecraft:")
					if strings.HasPrefix(item, lastWord) || strings.HasPrefix(cleanItem, lastWord) {
						suggestions = append(suggestions, item)
					}
				}
			} else {
				if strings.HasPrefix(child.Name, lastWord) {
					suggestions = append(suggestions, child.Name)
				}
			}
		}
		syntaxParts = append(syntaxParts, child.Name)
	}

	if len(syntaxParts) > 0 {
		nextSyntaxPreview = strings.Join(syntaxParts, " | ")
	}

	// Last word styling
	var lastIsLiteral bool
	for _, child := range currentNode.Children {
		if child.IsLiteral && child.Name == lastWord {
			lastIsLiteral = true
			break
		}
	}
	if lastWord != "" {
		styles = append(styles, WordStyle{
			Text:      lastWord,
			IsLiteral: lastIsLiteral,
			ArgIndex:  argIdx,
		})
	}

	return styles, suggestions, nextSyntaxPreview
}

// getStyledRunes parses runes input and maps each char to styledChar.
func getStyledRunes(input []rune) []styledChar {
	str := string(input)
	rawWords := combineCoordinates(strings.Split(str, " "))
	wordStyles, _, _ := getHighlightAndSuggestions(str)

	whiteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	cyanStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#55FFFF"))
	yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF55"))
	limeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#55FF55"))
	pinkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF55FF"))
	orangeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA00"))

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
		if wordIdx < len(wordStyles) {
			ws := wordStyles[wordIdx]
			if ws.IsLiteral {
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
			styledChars = append(styledChars, styledChar{char: ' ', style: whiteStyle})
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
}

func (m freeModel) Init() tea.Cmd {
	return nil
}

func (m *freeModel) updateSuggestions() {
	_, suggestions, _ := getHighlightAndSuggestions(string(m.input))
	m.suggestions = suggestions
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
	content := guide + "\n\n" + helpText

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
