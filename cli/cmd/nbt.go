package cmd

import (
	"cmdforge/core"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
)

// CustomNbtVal represents a configured non-boolean NBT tag.
type CustomNbtVal struct {
	Key   string
	Value string
	Type  string // "string", "int", "raw"
}

// GetApplicableOptions returns all NBT tag options applicable to the selected entity.
func GetApplicableOptions(selectedEntity string) []core.NbtTagOption {
	var options []core.NbtTagOption
	for _, opt := range core.NbtMasterList {
		isApplicable := len(opt.ApplicableTo) == 0
		if !isApplicable {
			for _, ent := range opt.ApplicableTo {
				if ent == selectedEntity {
					isApplicable = true
					break
				}
			}
		}
		if isApplicable {
			options = append(options, opt)
		}
	}
	return options
}

// GetFilteredNbtOptions filters the master list of NBT tags based on the selected entity
// and returns a list of huh.Option[string] for the multi-select UI.
func GetFilteredNbtOptions(selectedEntity string) []huh.Option[string] {
	var options []huh.Option[string]
	applicable := GetApplicableOptions(selectedEntity)
	for _, opt := range applicable {
		if opt.Type == "boolean" {
			options = append(options, huh.NewOption(opt.Label, opt.Key))
		}
	}
	return options
}

// PromptEntityNbt displays the interactive multi-select UI for boolean NBT tags
// and optional custom NBT strings, then merges them.
func PromptEntityNbt(selectedEntity string) (string, error) {
	// 0. Confirm if user wants to specify NBT tags
	var specifyNbt bool
	err := huh.NewConfirm().
		Title("Do you want to specify NBT tags?").
		Value(&specifyNbt).
		Run()
	if err != nil {
		return "", err
	}
	if !specifyNbt {
		return "", nil
	}

	var configuredCustom []CustomNbtVal
	var selectedBooleans []string

	applicableOpts := GetApplicableOptions(selectedEntity)

	for {
		var menuChoices []huh.Option[string]

		// 1st item: Added NBT tags
		addedLabel := fmt.Sprintf("[Added NBT Tags] (%d configured)", len(configuredCustom))
		menuChoices = append(menuChoices, huh.NewOption(addedLabel, "__ADDED__"))

		// 2nd item: Boolean NBT tags (persistent, never disappears)
		booleanLabel := "[Boolean NBT Tags] (NoAI, Silent, etc.)"
		if len(selectedBooleans) > 0 {
			booleanLabel = fmt.Sprintf("[Boolean NBT Tags] (%d active: %s)", len(selectedBooleans), strings.Join(selectedBooleans, ", "))
		}
		menuChoices = append(menuChoices, huh.NewOption(booleanLabel, "__BOOLEAN__"))

		// 3rd item onwards: Unconfigured custom NBT tags
		for _, opt := range applicableOpts {
			if opt.Type == "boolean" {
				continue
			}
			alreadyConfigured := false
			if opt.Type != "raw" { // Allow adding multiple Custom NBT tags
				for _, cfg := range configuredCustom {
					if cfg.Key == opt.Key {
						alreadyConfigured = true
						break
					}
				}
			}
			if !alreadyConfigured {
				menuChoices = append(menuChoices, huh.NewOption(opt.Label, opt.Key))
			}
		}

		// Last item: OK
		menuChoices = append(menuChoices, huh.NewOption("[OK - Finish]", "__OK__"))

		var choice string
		err := huh.NewSelect[string]().
			Title("Summon NBT Configuration Menu:").
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
				fmt.Println(color.YellowString("No custom NBT tags configured yet."))
				continue
			}

			for {
				if len(configuredCustom) == 0 {
					break
				}

				ti := textinput.New()
				ti.Placeholder = "New value"
				ti.CharLimit = 156
				ti.Width = 40

				m := addedTagsModel{
					tags:      configuredCustom,
					textInput: ti,
				}

				p := tea.NewProgram(m, tea.WithAltScreen())
				res, err := p.Run()
				if err != nil {
					return "", err
				}

				finalModel := res.(addedTagsModel)
				configuredCustom = finalModel.tags

				if finalModel.action == "edit_huh" {
					idx := finalModel.selected
					selectedTag := configuredCustom[idx]

					if selectedTag.Type == "raw" {
						var key, rawVal string
						err = huh.NewForm(
							huh.NewGroup(
								huh.NewInput().Title("Raw NBT Key (e.g. Attributes)").Value(&key).Placeholder(selectedTag.Key),
								huh.NewInput().Title("Raw NBT Value (e.g. [{Name:\"generic.max_health\",Base:20f}])").Value(&rawVal).Placeholder(selectedTag.Value),
							),
						).Run()
						if err == nil {
							key = strings.TrimSpace(key)
							rawVal = strings.TrimSpace(rawVal)
							if key != "" && rawVal != "" {
								configuredCustom[idx].Key = key
								configuredCustom[idx].Value = rawVal
								fmt.Println(color.GreenString("Updated %s to %s", key, rawVal))
							}
						}
					}
					continue
				}

				break
			}
			continue

		} else if choice == "__BOOLEAN__" {
			var boolOpts []huh.Option[string]
			for _, opt := range applicableOpts {
				if opt.Type == "boolean" {
					boolOpts = append(boolOpts, huh.NewOption(opt.Label, opt.Key))
				}
			}

			if len(boolOpts) == 0 {
				fmt.Println(color.YellowString("No Boolean NBT options available for this entity."))
				continue
			}

			err = huh.NewMultiSelect[string]().
				Title("Select Boolean NBT tags (Space to check/uncheck):").
				Options(boolOpts...).
				Value(&selectedBooleans).
				Height(12).
				Run()
			if err != nil {
				return "", err
			}

		} else {
			var targetOpt core.NbtTagOption
			for _, opt := range applicableOpts {
				if opt.Key == choice {
					targetOpt = opt
					break
				}
			}

			var val string
			switch targetOpt.Type {
			case "select":
				var options []huh.Option[string]
				for _, c := range targetOpt.Choices {
					options = append(options, huh.NewOption(c.Label, c.Value))
				}
				err = huh.NewSelect[string]().
					Title(fmt.Sprintf("Select value for %s:", targetOpt.Key)).
					Options(options...).
					Value(&val).
					Run()
				if err == nil {
					val = strings.TrimSpace(val)
				}
			case "string":
				err = huh.NewInput().
					Title(fmt.Sprintf("Enter string value for %s:", targetOpt.Key)).
					Value(&val).
					Run()
				if err == nil && strings.TrimSpace(val) != "" {
					val = strings.TrimSpace(val)
					if !strings.HasPrefix(val, "'") && !strings.HasPrefix(val, "\"") {
						val = fmt.Sprintf(`'"%s"'`, val)
					}
				}
			case "int":
				err = huh.NewInput().
					Title(fmt.Sprintf("Enter integer value for %s:", targetOpt.Key)).
					Value(&val).
					Validate(func(s string) error {
						_, err := strconv.Atoi(strings.TrimSpace(s))
						if err != nil {
							return fmt.Errorf("must be a valid integer")
						}
						return nil
					}).
					Run()
				if err == nil {
					val = strings.TrimSpace(val)
				}
			case "raw":
				var key, rawVal string
				err = huh.NewForm(
					huh.NewGroup(
						huh.NewInput().Title("Raw NBT Key (e.g. Attributes)").Value(&key),
						huh.NewInput().Title("Raw NBT Value (e.g. [{Name:\"generic.max_health\",Base:20f}])").Value(&rawVal),
					),
				).Run()
				if err == nil && strings.TrimSpace(key) != "" && strings.TrimSpace(rawVal) != "" {
					targetOpt.Key = strings.TrimSpace(key)
					val = strings.TrimSpace(rawVal)
				}
			}

			if err != nil {
				return "", err
			}

			if val != "" {
				configuredCustom = append(configuredCustom, CustomNbtVal{
					Key:   targetOpt.Key,
					Value: val,
					Type:  targetOpt.Type,
				})
				fmt.Println(color.GreenString("Added %s: %s", targetOpt.Key, val))
			}
		}
	}

	assembledBool := core.AssembleNbtString(selectedBooleans)

	var customParts []string
	for _, cfg := range configuredCustom {
		customParts = append(customParts, fmt.Sprintf("%s:%s", cfg.Key, cfg.Value))
	}
	assembledCustom := ""
	if len(customParts) > 0 {
		assembledCustom = "{" + strings.Join(customParts, ",") + "}"
	}

	finalNbt := MergeNbtStrings(assembledBool, assembledCustom)
	return finalNbt, nil
}

// MergeNbtStrings merges the assembled boolean NBT tags and custom NBT string into one compound.
func MergeNbtStrings(assembled, custom string) string {
	assembled = strings.TrimSpace(assembled)
	custom = strings.TrimSpace(custom)

	if assembled == "" {
		if custom == "" {
			return ""
		}
		if !strings.HasPrefix(custom, "{") {
			return "{" + custom + "}"
		}
		return custom
	}

	if custom == "" {
		return assembled
	}

	// Remove outer braces of assembled
	assembledContent := strings.TrimPrefix(assembled, "{")
	assembledContent = strings.TrimSuffix(assembledContent, "}")

	// Remove outer braces of custom if present
	customContent := custom
	if strings.HasPrefix(customContent, "{") && strings.HasSuffix(customContent, "}") {
		customContent = strings.TrimPrefix(customContent, "{")
		customContent = strings.TrimSuffix(customContent, "}")
	}

	return "{" + assembledContent + "," + customContent + "}"
}

func isNbtHuhType(optType string) bool {
	return optType == "raw"
}

// bubbletea model for added custom NBT tags list to support keys "e" and "d" with inline textinput
type addedTagsModel struct {
	tags      []CustomNbtVal
	cursor    int
	textInput textinput.Model
	editing   bool
	action    string // "edit_huh", "back", etc.
	selected  int    // index of tag to edit via huh
}

func (m addedTagsModel) Init() tea.Cmd {
	return nil
}

func (m addedTagsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.editing {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				val := strings.TrimSpace(m.textInput.Value())
				if val != "" {
					selectedTag := m.tags[m.cursor]
					if selectedTag.Type == "string" {
						if !strings.HasPrefix(val, "'") && !strings.HasPrefix(val, "\"") {
							val = fmt.Sprintf(`'"%s"'`, val)
						}
					}
					m.tags[m.cursor].Value = val
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
			m.action = "back"
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.tags)-1 {
				m.cursor++
			}

		case "e":
			if len(m.tags) > 0 {
				tag := m.tags[m.cursor]
				if isNbtHuhType(tag.Type) {
					m.action = "edit_huh"
					m.selected = m.cursor
					return m, tea.Quit
				} else {
					m.editing = true
					m.textInput.SetValue(m.tags[m.cursor].Value)
					m.textInput.Focus()
					return m, textinput.Blink
				}
			}

		case "d":
			if len(m.tags) > 0 {
				m.tags = append(m.tags[:m.cursor], m.tags[m.cursor+1:]...)
				if m.cursor >= len(m.tags) && m.cursor > 0 {
					m.cursor = len(m.tags) - 1
				}
				if len(m.tags) == 0 {
					return m, tea.Quit
				}
			}
		}
	}
	return m, nil
}

func (m addedTagsModel) View() string {
	var s strings.Builder
	fmt.Fprintf(&s, "\n  %s\n\n", color.CyanString("Configure Added NBT Tags:"))

	if len(m.tags) == 0 {
		s.WriteString("    No NBT tags configured yet.\n")
	} else {
		for i, tag := range m.tags {
			if m.cursor == i {
				s.WriteString(fmt.Sprintf("  %s %s: %s\n",
					color.CyanString(">"),
					color.CyanString(tag.Key),
					color.YellowString(tag.Value)))
			} else {
				s.WriteString(fmt.Sprintf("    %s: %s\n", tag.Key, tag.Value))
			}
		}
	}

	if m.editing {
		fmt.Fprintf(&s, "\n  %s\n", color.GreenString("Editing value for %s:", m.tags[m.cursor].Key))
		fmt.Fprintf(&s, "  %s\n", m.textInput.View())
		fmt.Fprintf(&s, "\n  %s\n", color.HiBlackString("(press 'enter' to save, 'esc' to cancel)"))
	} else {
		fmt.Fprintf(&s, "\n  %s\n", color.HiBlackString("(press 'e' to edit, 'd' to delete, 'q' or 'enter' to go back)"))
	}

	return s.String()
}
