package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
)

type CustomSelectorVal struct {
	Key   string
	Value string
}

func addSelectorsQuestion() (string, error) {
	for {
		targetOptions := []string{
			"@p - Near Player",
			"@a - All Player",
			"@s - Myself",
			"@r - Random Player",
			"@n - Nearest Player (1.21+)",
			"PlayerName",
		}

		var choice string
		err := huh.NewSelect[string]().
			Title("Select a target selector type:").
			Options(huh.NewOptions(targetOptions...)...).
			Value(&choice).
			Run()
		if err != nil {
			return "", err
		}

		targetType := strings.Split(choice, " ")[0]

		if targetType == "PlayerName" {
			var name string
			err = huh.NewInput().
				Title("Enter the player name (Type \"back\" to go back).\nNote: Use '!' prefix (e.g. !notch) to skip Mojang verification:").
				Value(&name).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("player name cannot be empty")
					}
					return nil
				}).
				Run()
			if err != nil {
				return "", err
			}

			name = strings.TrimSpace(name)
			if strings.ToLower(name) == "back" {
				continue
			}

			if strings.HasPrefix(name, "!") {
				targetType = strings.TrimPrefix(name, "!")
			} else {
				fmt.Println(color.CyanString("Checking player existence..."))
				if mojangPlayerExists(name) {
					fmt.Println(color.GreenString("Player %s exists in Mojang database!", name))
					targetType = name
				} else {
					fmt.Println(color.YellowString("WARN: Player %s not found in Mojang database.", name))
					var useAnyway bool
					err = huh.NewConfirm().
						Title("Use this name anyway?").
						Value(&useAnyway).
						Run()
					if err != nil {
						return "", err
					}
					if useAnyway {
						targetType = name
					} else {
						continue
					}
				}
			}
		}

		greenBold := color.New(color.FgGreen, color.Bold)
		fmt.Println(color.BlueString("Target:"), greenBold.Sprint(targetType))

		if !strings.HasPrefix(targetType, "@") {
			fmt.Println(color.BlueString("Additional target selectors:"), color.GreenString("Skipped (Specific Player Name)"))
			return targetType, nil
		}

		var refine bool
		err = huh.NewConfirm().
			Title("Do you want to add more target selectors? (distance, score, tag, team, limit, etc.)").
			Value(&refine).
			Run()
		if err != nil {
			return "", err
		}

		if !refine {
			return targetType, nil
		}

		allSelectorTypes := []string{
			"distance",
			"score",
			"tag",
			"team",
			"limit",
			"level",
			"gamemode",
			"advancements",
			"predicate",
			"sort",
		}

		descriptions := map[string]string{
			"distance":     "Distance to Entity(=Player)",
			"score":        "The score value or range",
			"tag":          "The tag which the entity has",
			"team":         "The team which the entity joins",
			"limit":        "Amount limit",
			"level":        "Experience level",
			"gamemode":     "Player gamemode",
			"advancements": "The advancements which the player has",
			"predicate":    "Match predicates",
			"sort":         "Specify selection sorting order",
		}

		var configuredCustom []CustomSelectorVal

		for {
			var available []string
			for _, s := range allSelectorTypes {
				alreadyAdded := false
				for _, added := range configuredCustom {
					if added.Key == s {
						alreadyAdded = true
						break
					}
				}
				if !alreadyAdded {
					available = append(available, s)
				}
			}

			var menuChoices []huh.Option[string]

			// 1st item: Added selectors
			addedLabel := fmt.Sprintf("[Added Selectors] (%d configured)", len(configuredCustom))
			menuChoices = append(menuChoices, huh.NewOption(addedLabel, "__ADDED__"))

			// 2nd item onwards: Unconfigured custom selectors
			for _, av := range available {
				menuChoices = append(menuChoices, huh.NewOption(fmt.Sprintf("%s - %s", av, descriptions[av]), av))
			}

			// Last item: OK
			menuChoices = append(menuChoices, huh.NewOption("[OK - Finish]", "__OK__"))

			var choice string
			err = huh.NewSelect[string]().
				Title("Selector Configuration Menu:").
				Options(menuChoices...).
				Value(&choice).
				Run()
			if err != nil {
				return "", err
			}

			if choice == "__OK__" {
				break
			}

			if choice == "__ADDED__" {
				if len(configuredCustom) == 0 {
					fmt.Println(color.YellowString("No custom selectors configured yet."))
					continue
				}

				ti := textinput.New()
				ti.Placeholder = "New value"
				ti.CharLimit = 156
				ti.Width = 40

				m := addedSelectorsModel{
					options:   configuredCustom,
					textInput: ti,
				}

				p := tea.NewProgram(m, tea.WithAltScreen())
				res, err := p.Run()
				if err != nil {
					return "", err
				}

				finalModel := res.(addedSelectorsModel)
				configuredCustom = finalModel.options
				continue
			}

			// Add new custom selector
			selKey := choice
			var val string
			switch selKey {
			case "distance":
				err = huh.NewInput().Title("Distance (e.g. 1..5):").Value(&val).Run()
			case "score":
				err = huh.NewInput().Title("Score format (e.g. score=1..10):").Value(&val).Run()
			case "tag":
				err = huh.NewInput().Title("Tag (e.g. tag_name or !tag_name):").Value(&val).Run()
			case "team":
				err = huh.NewInput().Title("Team (e.g. team_name or !team_name):").Value(&val).Run()
			case "limit":
				err = huh.NewInput().Title("Limit (e.g. 1):").Value(&val).Run()
			case "level":
				err = huh.NewInput().Title("Exp level (e.g. 10 or 10..20):").Value(&val).Run()
			case "gamemode":
				err = huh.NewSelect[string]().
					Title("Player gamemode:").
					Options(huh.NewOptions("survival", "creative", "adventure", "spectator")...).
					Value(&val).
					Run()
			case "advancements":
				err = huh.NewInput().Title("Advancement (e.g. adv_id=true):").Value(&val).Run()
			case "predicate":
				err = huh.NewInput().Title("Predicate (e.g. predicate_id):").Value(&val).Run()
			case "sort":
				err = huh.NewSelect[string]().
					Title("Sort order:").
					Options(huh.NewOptions("nearest", "furthest", "random", "arbitrary")...).
					Value(&val).
					Run()
			}
			if err != nil {
				return "", err
			}

			val = strings.TrimSpace(val)
			if val != "" {
				configuredCustom = append(configuredCustom, CustomSelectorVal{
					Key:   selKey,
					Value: val,
				})
				fmt.Println(color.GreenString("Added %s: %s", selKey, val))
			}
		}

		if len(configuredCustom) > 0 {
			var parts []string
			for _, sc := range configuredCustom {
				parts = append(parts, fmt.Sprintf("%s=%s", sc.Key, sc.Value))
			}
			targetType = fmt.Sprintf("%s[%s]", targetType, strings.Join(parts, ","))
		}
		return targetType, nil
	}
}

// bubbletea model for added custom target selector parameters
type addedSelectorsModel struct {
	options   []CustomSelectorVal
	cursor    int
	textInput textinput.Model
	editing   bool
}

func (m addedSelectorsModel) Init() tea.Cmd {
	return nil
}

func (m addedSelectorsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.editing {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				val := strings.TrimSpace(m.textInput.Value())
				if val != "" {
					m.options[m.cursor].Value = val
				}
				m.editing = false
				m.textInput.Blur()
				return m, nil

			case "esc":
				m.editing = false
				m.textInput.Blur()
				return m, nil
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc", "enter":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}

		case "e":
			if len(m.options) > 0 {
				m.editing = true
				m.textInput.SetValue(m.options[m.cursor].Value)
				m.textInput.Focus()
				return m, textinput.Blink
			}

		case "d":
			if len(m.options) > 0 {
				m.options = append(m.options[:m.cursor], m.options[m.cursor+1:]...)
				if m.cursor >= len(m.options) && m.cursor > 0 {
					m.cursor = len(m.options) - 1
				}
				if len(m.options) == 0 {
					return m, tea.Quit
				}
			}
		}
	}
	return m, nil
}

func (m addedSelectorsModel) View() string {
	var s strings.Builder
	fmt.Fprintf(&s, "\n  %s\n\n", color.CyanString("Configure Added Selectors:"))

	if len(m.options) == 0 {
		s.WriteString("    No custom selectors configured yet.\n")
	} else {
		for i, sc := range m.options {
			if m.cursor == i {
				s.WriteString(fmt.Sprintf("  %s %s=%s\n",
					color.CyanString(">"),
					color.CyanString(sc.Key),
					color.YellowString(sc.Value)))
			} else {
				s.WriteString(fmt.Sprintf("    %s=%s\n", sc.Key, sc.Value))
			}
		}
	}

	if m.editing {
		fmt.Fprintf(&s, "\n  %s\n", color.GreenString("Editing value for %s:", m.options[m.cursor].Key))
		fmt.Fprintf(&s, "  %s\n", m.textInput.View())
		fmt.Fprintf(&s, "\n  %s\n", color.HiBlackString("(press 'enter' to save, 'esc' to cancel)"))
	} else {
		fmt.Fprintf(&s, "\n  %s\n", color.HiBlackString("(press 'e' to edit, 'd' to delete, 'q' or 'enter' to go back)"))
	}

	return s.String()
}
