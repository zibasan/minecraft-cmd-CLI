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

// getHighlightAndSuggestions is a wrapper for backward compatibility with existing tests.
func getHighlightAndSuggestions(input string) ([]WordStyle, []string, string, int, string) {
	res := core.ParseCommand(input)
	var styles []WordStyle
	for _, w := range res.Words {
		styles = append(styles, WordStyle{
			Text:      w.Text,
			IsLiteral: w.IsLiteral,
			ArgIndex:  w.ArgIndex,
			IsError:   w.IsError,
		})
	}
	return styles, res.Suggestions, res.SyntaxPreview, res.ErrorIdx, res.CurrentParser
}

// getStyledRunes parses runes input and maps each char to styledChar using dynamic parsed words.
func getStyledRunes(input []rune) []styledChar {
	str := string(input)
	res := core.ParseCommand(str)

	yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF55"))
	cyanStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#55FFFF"))
	limeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#55FF55"))
	pinkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF55FF"))
	orangeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA00"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Underline(true)
	whiteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	var styledChars []styledChar
	runeIdx := 0

	for wIdx, w := range res.Words {
		isErrorWord := res.ErrorIdx != -1 && wIdx >= res.ErrorIdx
		var currentStyle lipgloss.Style

		if isErrorWord {
			currentStyle = errorStyle
		} else if w.IsLiteral {
			currentStyle = whiteStyle
		} else {
			switch w.ArgIndex % 5 {
			case 0:
				currentStyle = cyanStyle
			case 1:
				currentStyle = yellowStyle
			case 2:
				currentStyle = limeStyle
			case 3:
				currentStyle = pinkStyle
			case 4:
				currentStyle = orangeStyle
			default:
				currentStyle = whiteStyle
			}
		}

		wRunes := []rune(w.Text)
		for range wRunes {
			if runeIdx >= len(input) {
				break
			}
			styledChars = append(styledChars, styledChar{char: input[runeIdx], style: currentStyle})
			runeIdx++
		}

		for runeIdx < len(input) && input[runeIdx] == ' ' {
			spaceStyle := whiteStyle
			if res.ErrorIdx != -1 && wIdx >= res.ErrorIdx {
				spaceStyle = errorStyle
			}
			styledChars = append(styledChars, styledChar{char: ' ', style: spaceStyle})
			runeIdx++
		}
	}

	for runeIdx < len(input) {
		styledChars = append(styledChars, styledChar{char: input[runeIdx], style: whiteStyle})
		runeIdx++
	}

	return styledChars
}

// highlightCommand turns an input string into a LipGloss-styled formatted string.
func highlightCommand(input string) string {
	styled := getStyledRunes([]rune(input))
	var sb strings.Builder
	for _, sc := range styled {
		sb.WriteString(sc.style.Render(string(sc.char)))
	}
	return sb.String()
}

// getSyntaxGuide returns dynamic syntax guides based on the commands tree.
func getSyntaxGuide(input string) string {
	inputTrim := strings.TrimSpace(input)
	if inputTrim == "" {
		return "コマンドを入力してください (例: /give, /summon, /item, /gamemode, /tp, /effect):"
	}

	res := core.ParseCommand(input)

	var prefixParts []string
	for _, w := range res.Words {
		prefixParts = append(prefixParts, w.Text)
	}
	prefix := strings.Join(prefixParts, " ")
	if strings.HasPrefix(input, "/") && !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	guides := core.GetSyntaxGuides(res.CurrentNode, prefix)

	var sb strings.Builder
	if len(guides) > 0 {
		sb.WriteString("構文ガイド:\n")
		maxGuides := 5
		for i, g := range guides {
			if i >= maxGuides {
				sb.WriteString(fmt.Sprintf("  ... 他 %d 件の構文パターンがあります\n", len(guides)-maxGuides))
				break
			}
			sb.WriteString("  " + g + "\n")
		}
	} else {
		sb.WriteString("構文ガイド: なし\n")
	}

	return strings.TrimSuffix(sb.String(), "\n")
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
	highlighted := highlightCommand(string(m.input))
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	var sb strings.Builder
	sb.WriteString(" > ")

	if m.cursorIdx == len(m.input) {
		sb.WriteString(highlighted)
		sb.WriteString(cursorStyle.Render("█"))
	} else {
		styled := getStyledRunes(m.input)
		for i := 0; i < len(m.input); i++ {
			if i == m.cursorIdx {
				invertStyle := styled[i].style.Copy().Reverse(true)
				sb.WriteString(invertStyle.Render(string(m.input[i])))
			} else {
				sb.WriteString(styled[i].style.Render(string(styled[i].char)))
			}
		}
		sb.WriteString(cursorStyle.Render("█"))
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

	helpTitle := lipgloss.NewStyle().Foreground(lipgloss.Color("#55FF55")).Render("操作ヘルプ:")
	helpText := helpTitle + "\n" +
		"  [Tab]   : 選択中の候補を補完\n" +
		"  [Esc]   : 終了\n" +
		"  [Up/Down/Ctrl+N/Ctrl+P] : 候補の選択切り替え\n" +
		"  [Enter] : コマンド生成を実行"

	res := core.ParseCommand(string(m.input))
	guide := getSyntaxGuide(string(m.input))

	var cmdDesc string
	if len(res.Words) > 0 {
		cmdName := strings.TrimPrefix(res.Words[0].Text, "/")
		if desc, exists := core.CommandDescriptions[cmdName]; exists {
			title := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF55")).Render("コマンド説明:")
			descText := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Render(desc)
			cmdDesc = "\n" + title + "\n  " + descText
		}
	}

	var argGuidance string
	if res.CurrentNode != nil && res.CurrentNode.Type == "argument" {
		title := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF55")).Render("引数情報:")
		
		parserDisp := res.CurrentParser
		if regDisp, exists := core.RegistryTypeNames[res.Registry]; exists {
			parserDisp = regDisp
		} else if jp, exists := core.ParserTypeNames[res.CurrentParser]; exists {
			parserDisp = jp
		}
		
		propDesc := res.CurrentNode.GetPropertiesDescription()
		typeLine := lipgloss.NewStyle().Foreground(lipgloss.Color("#55FFFF")).Render("  データ型: ") + parserDisp + propDesc
		
		var descLine string
		if regDesc, exists := core.RegistryDescriptions[res.Registry]; exists {
			descLine = "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#55FF55")).Render("  説明: ") + lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Render(regDesc)
		} else if desc, exists := core.ArgumentDescriptions[res.CurrentNodeName]; exists {
			descLine = "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#55FF55")).Render("  説明: ") + lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Render(desc)
		}
		
		argGuidance = "\n" + title + "\n" + typeLine + descLine
	}

	content := guide + cmdDesc + argGuidance + "\n\n" + helpText

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

	container := lipgloss.JoinVertical(lipgloss.Left, elements...)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		MaxHeight(m.height).
		MaxWidth(m.width).
		Render(container)
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
