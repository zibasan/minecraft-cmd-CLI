package cmd

import (
	"fmt"
	"sort"
	"strings"

	"cmdforge/core"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
)

type CustomComponentVal struct {
	Key        string
	Serialized string
}

var ComponentDescriptions = map[string]string{
	"item_name":                  "Item Name (Override the original name)",
	"custom_name":                "Item Name (looks like it was edited with an anvil)",
	"lore":                       "Item Lore",
	"damage":                     "How much to reduce the durability",
	"enchantment_glint_override": "Whether show glint of enchantment (no enchantments)",
	"enchantments":               "Item Enchantments",
	"food":                       "Setting edible items",
	"break_sound":                "The sound played when the item is broken",
	"max_damage":                 "The maximum durability value of that item",
	"max_stack_size":             "The maximum stack size value of that item",
	"can_break":                  "Specify breakable blocks in adventure mode",
	"can_place_on":               "Specify blocks on which this can be placed",
	"rarity":                     "Item Rarity",
}

func addItemComponentsQuestion() (string, error) {
	var configuredCustom []CustomComponentVal

	for {
		available := []string{}
		for k := range ComponentDescriptions {
			already := false
			for _, added := range configuredCustom {
				if added.Key == k {
					already = true
					break
				}
			}
			if !already {
				available = append(available, k)
			}
		}

		choices := []string{}
		for _, av := range available {
			choices = append(choices, fmt.Sprintf("%s - %s", av, ComponentDescriptions[av]))
		}
		sort.Strings(choices)

		var menuChoices []huh.Option[string]

		// 1st: Added components
		addedLabel := fmt.Sprintf("[Added Components] (%d configured)", len(configuredCustom))
		menuChoices = append(menuChoices, huh.NewOption(addedLabel, "__ADDED__"))

		// 2nd onwards: Unconfigured custom components
		for _, ch := range choices {
			key := strings.Split(ch, " ")[0]
			menuChoices = append(menuChoices, huh.NewOption(ch, key))
		}

		// Last: OK
		menuChoices = append(menuChoices, huh.NewOption("[OK - Finish]", "__OK__"))

		var choice string
		err := huh.NewSelect[string]().
			Title("Additional components (Select OK to finish):").
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
				fmt.Println(color.YellowString("No custom components configured yet."))
				continue
			}

			for {
				if len(configuredCustom) == 0 {
					break
				}

				ti := textinput.New()
				ti.Placeholder = "Edit component string"
				ti.CharLimit = 256
				ti.Width = 60

				m := addedComponentsModel{
					components: configuredCustom,
					textInput:  ti,
				}

				p := tea.NewProgram(m, tea.WithAltScreen())
				res, err := p.Run()
				if err != nil {
					return "", err
				}

				finalModel := res.(addedComponentsModel)
				configuredCustom = finalModel.components

				if finalModel.action == "edit_huh" {
					idx := finalModel.selected
					compKey := configuredCustom[idx].Key
					var compVal string
					var err error

					switch compKey {
					case "enchantment_glint_override":
						var glint bool
						err = huh.NewConfirm().Title("enchantment_glint_override:").Value(&glint).Run()
						if err == nil {
							compVal = fmt.Sprintf("enchantment_glint_override=%t", glint)
						}
					case "rarity":
						err = huh.NewSelect[string]().
							Title("Rarity:").
							Options(huh.NewOptions("common", "uncommon", "rare", "epic")...).
							Value(&compVal).
							Run()
						if err == nil && compVal != "" {
							compVal = "rarity=" + compVal
						}
					case "break_sound":
						var sound string
						err = huh.NewSelect[string]().
							Title("Select sound:").
							Options(buildSelectOptions(core.Sounds)...).
							Value(&sound).
							Filtering(true).
							Height(12).
							Run()
						if err == nil {
							compVal = fmt.Sprintf(`break_sound="%s"`, sound)
						}
					case "enchantments":
						enchantmentsList := []string{}
						for {
							var enchantName string
							var enchOptions []huh.Option[string]
							for _, e := range core.Enchantments {
								enchOptions = append(enchOptions, huh.NewOption(e.Name, e.Name))
							}

							err = huh.NewSelect[string]().
								Title("Select enchantment (Select OK to finish):").
								Options(append([]huh.Option[string]{huh.NewOption("OK", "OK")}, enchOptions...)...).
								Value(&enchantName).
								Filtering(true).
								Height(12).
								Run()
							if err != nil || enchantName == "OK" {
								break
							}

							var levelStr string
							err = huh.NewInput().Title("Level (1-255):").Value(&levelStr).Run()
							if err != nil {
								break
							}
							enchantmentsList = append(enchantmentsList, fmt.Sprintf(`"%s":%s`, enchantName, levelStr))

							var addMore bool
							err = huh.NewConfirm().Title("Add another enchantment?").Value(&addMore).Run()
							if err != nil || !addMore {
								break
							}
						}
						if len(enchantmentsList) > 0 {
							compVal = fmt.Sprintf("enchantments={levels:{%s}}", strings.Join(enchantmentsList, ","))
						}
					case "food":
						var nutrition, saturation string
						var canAlwaysEat bool
						err = huh.NewForm(
							huh.NewGroup(
								huh.NewInput().Title("Nutrition (int):").Value(&nutrition),
								huh.NewInput().Title("Saturation (float):").Value(&saturation),
								huh.NewConfirm().Title("Can always eat?").Value(&canAlwaysEat),
							),
						).Run()
						if err == nil {
							compVal = fmt.Sprintf("food={nutrition:%s,saturation:%s,can_always_eat:%t}", nutrition, saturation, canAlwaysEat)
						}
					case "can_break":
						blocksList := []string{}
						for {
							var block string
							err = huh.NewSelect[string]().
								Title("Select breakable block (Select OK to finish):").
								Options(append([]huh.Option[string]{huh.NewOption("OK", "OK")}, buildSelectOptions(core.Blocks)...)...).
								Value(&block).
								Filtering(true).
								Height(12).
								Run()
							if err != nil || block == "OK" {
								break
							}
							blocksList = append(blocksList, fmt.Sprintf(`"%s"`, block))

							var addMore bool
							err = huh.NewConfirm().Title("Add another block?").Value(&addMore).Run()
							if err != nil || !addMore {
								break
							}
						}
						if len(blocksList) > 0 {
							compVal = fmt.Sprintf("can_break={blocks:[%s]}", strings.Join(blocksList, ","))
						}
					case "can_place_on":
						blocksList := []string{}
						for {
							var block string
							err = huh.NewSelect[string]().
								Title("Select placeable block (Select OK to finish):").
								Options(append([]huh.Option[string]{huh.NewOption("OK", "OK")}, buildSelectOptions(core.Blocks)...)...).
								Value(&block).
								Filtering(true).
								Height(12).
								Run()
							if err != nil || block == "OK" {
								break
							}
							blocksList = append(blocksList, fmt.Sprintf(`"%s"`, block))

							var addMore bool
							err = huh.NewConfirm().Title("Add another block?").Value(&addMore).Run()
							if err != nil || !addMore {
								break
							}
						}
						if len(blocksList) > 0 {
							compVal = fmt.Sprintf("can_place_on={blocks:[%s]}", strings.Join(blocksList, ","))
						}
					}

					if err == nil && compVal != "" {
						configuredCustom[idx].Serialized = compVal
						fmt.Println(color.GreenString("Updated component: %s", compVal))
					}
					continue
				}

				break
			}
			continue
		}

		compKey := choice
		var compVal string
		switch compKey {
		case "item_name":
			err = huh.NewInput().Title("item_name:").Value(&compVal).Run()
			if err == nil && compVal != "" {
				compVal = fmt.Sprintf(`item_name='"%s"'`, compVal)
			}
		case "custom_name":
			err = huh.NewInput().Title("custom_name:").Value(&compVal).Run()
			if err == nil && compVal != "" {
				compVal = fmt.Sprintf(`custom_name='"%s"'`, compVal)
			}
		case "lore":
			var rawLore string
			err = huh.NewInput().Title("lore (insert '<br>' to start a new line):").Value(&rawLore).Run()
			if err == nil && rawLore != "" {
				parts := strings.Split(rawLore, "<br>")
				var loreLines []string
				for _, p := range parts {
					loreLines = append(loreLines, fmt.Sprintf(`'"%s"'`, strings.TrimSpace(p)))
				}
				compVal = fmt.Sprintf("lore=[%s]", strings.Join(loreLines, ","))
			}
		case "damage":
			err = huh.NewInput().Title("damage:").Value(&compVal).Run()
			if err == nil && compVal != "" {
				compVal = "damage=" + compVal
			}
		case "enchantment_glint_override":
			var glint bool
			err = huh.NewConfirm().Title("enchantment_glint_override:").Value(&glint).Run()
			if err == nil {
				compVal = fmt.Sprintf("enchantment_glint_override=%t", glint)
			}
		case "enchantments":
			enchantmentsList := []string{}
			for {
				var enchantName string
				var enchOptions []huh.Option[string]
				for _, e := range core.Enchantments {
					enchOptions = append(enchOptions, huh.NewOption(e.Name, e.Name))
				}

				err = huh.NewSelect[string]().
					Title("Select enchantment (Select OK to finish):").
					Options(append([]huh.Option[string]{huh.NewOption("OK", "OK")}, enchOptions...)...).
					Value(&enchantName).
					Filtering(true).
					Height(12).
					Run()
				if err != nil || enchantName == "OK" {
					break
				}

				var levelStr string
				err = huh.NewInput().Title("Level (1-255):").Value(&levelStr).Run()
				if err != nil {
					break
				}
				enchantmentsList = append(enchantmentsList, fmt.Sprintf(`"%s":%s`, enchantName, levelStr))

				var addMore bool
				err = huh.NewConfirm().Title("Add another enchantment?").Value(&addMore).Run()
				if err != nil || !addMore {
					break
				}
			}
			if len(enchantmentsList) > 0 {
				compVal = fmt.Sprintf("enchantments={levels:{%s}}", strings.Join(enchantmentsList, ","))
			}
		case "food":
			var nutrition, saturation string
			var canAlwaysEat bool
			err = huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Title("Nutrition (int):").Value(&nutrition),
					huh.NewInput().Title("Saturation (float):").Value(&saturation),
					huh.NewConfirm().Title("Can always eat?").Value(&canAlwaysEat),
				),
			).Run()
			if err == nil {
				compVal = fmt.Sprintf("food={nutrition:%s,saturation:%s,can_always_eat:%t}", nutrition, saturation, canAlwaysEat)
			}
		case "break_sound":
			var sound string
			err = huh.NewSelect[string]().
				Title("Select sound:").
				Options(buildSelectOptions(core.Sounds)...).
				Value(&sound).
				Filtering(true).
				Height(12).
				Run()
			if err == nil {
				compVal = fmt.Sprintf(`break_sound="%s"`, sound)
			}
		case "max_damage":
			err = huh.NewInput().Title("max_damage:").Value(&compVal).Run()
			if err == nil && compVal != "" {
				compVal = "max_damage=" + compVal
			}
		case "max_stack_size":
			err = huh.NewInput().Title("max_stack_size (1-99):").Value(&compVal).Run()
			if err == nil && compVal != "" {
				compVal = "max_stack_size=" + compVal
			}
		case "can_break":
			blocksList := []string{}
			for {
				var block string
				err = huh.NewSelect[string]().
					Title("Select breakable block (Select OK to finish):").
					Options(append([]huh.Option[string]{huh.NewOption("OK", "OK")}, buildSelectOptions(core.Blocks)...)...).
					Value(&block).
					Filtering(true).
					Height(12).
					Run()
				if err != nil || block == "OK" {
					break
				}
				blocksList = append(blocksList, fmt.Sprintf(`"%s"`, block))

				var addMore bool
				err = huh.NewConfirm().Title("Add another block?").Value(&addMore).Run()
				if err != nil || !addMore {
					break
				}
			}
			if len(blocksList) > 0 {
				compVal = fmt.Sprintf("can_break={blocks:[%s]}", strings.Join(blocksList, ","))
			}
		case "can_place_on":
			blocksList := []string{}
			for {
				var block string
				err = huh.NewSelect[string]().
					Title("Select placeable block (Select OK to finish):").
					Options(append([]huh.Option[string]{huh.NewOption("OK", "OK")}, buildSelectOptions(core.Blocks)...)...).
					Value(&block).
					Filtering(true).
					Height(12).
					Run()
				if err != nil || block == "OK" {
					break
				}
				blocksList = append(blocksList, fmt.Sprintf(`"%s"`, block))

				var addMore bool
				err = huh.NewConfirm().Title("Add another block?").Value(&addMore).Run()
				if err != nil || !addMore {
					break
				}
			}
			if len(blocksList) > 0 {
				compVal = fmt.Sprintf("can_place_on={blocks:[%s]}", strings.Join(blocksList, ","))
			}
		case "rarity":
			err = huh.NewSelect[string]().
				Title("Rarity:").
				Options(huh.NewOptions("common", "uncommon", "rare", "epic")...).
				Value(&compVal).
				Run()
			if err == nil && compVal != "" {
				compVal = "rarity=" + compVal
			}
		}

		if err != nil {
			return "", err
		}

		if compVal != "" {
			configuredCustom = append(configuredCustom, CustomComponentVal{
				Key:        compKey,
				Serialized: compVal,
			})
			fmt.Println(color.GreenString("Added component: %s", compVal))
		}
	}

	var finalParts []string
	for _, cc := range configuredCustom {
		finalParts = append(finalParts, cc.Serialized)
	}
	return strings.Join(finalParts, ","), nil
}

func isComponentHuhType(key string) bool {
	switch key {
	case "enchantment_glint_override", "rarity", "break_sound", "enchantments", "food", "can_break", "can_place_on":
		return true
	}
	return false
}

// bubbletea model for added custom item components
type addedComponentsModel struct {
	components []CustomComponentVal
	cursor     int
	textInput  textinput.Model
	editing    bool
	action     string // "edit_huh", "back", etc.
	selected   int    // index of component to edit via huh
}

func (m addedComponentsModel) Init() tea.Cmd {
	return nil
}

func (m addedComponentsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.editing {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				val := strings.TrimSpace(m.textInput.Value())
				if val != "" {
					m.components[m.cursor].Serialized = val
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
			if m.cursor < len(m.components)-1 {
				m.cursor++
			}

		case "e":
			if len(m.components) > 0 {
				key := m.components[m.cursor].Key
				if isComponentHuhType(key) {
					m.action = "edit_huh"
					m.selected = m.cursor
					return m, tea.Quit
				} else {
					m.editing = true
					m.textInput.SetValue(m.components[m.cursor].Serialized)
					m.textInput.Focus()
					return m, textinput.Blink
				}
			}

		case "d":
			if len(m.components) > 0 {
				m.components = append(m.components[:m.cursor], m.components[m.cursor+1:]...)
				if m.cursor >= len(m.components) && m.cursor > 0 {
					m.cursor = len(m.components) - 1
				}
				if len(m.components) == 0 {
					return m, tea.Quit
				}
			}
		}
	}
	return m, nil
}

func (m addedComponentsModel) View() string {
	var s strings.Builder
	fmt.Fprintf(&s, "\n  %s\n\n", color.CyanString("Configure Added Components:"))

	if len(m.components) == 0 {
		s.WriteString("    No custom components configured yet.\n")
	} else {
		for i, cc := range m.components {
			if m.cursor == i {
				fmt.Fprintf(&s, "  %s %s\n",
					color.CyanString(">"),
					color.YellowString(cc.Serialized))
			} else {
				fmt.Fprintf(&s, "    %s\n", cc.Serialized)
			}
		}
	}

	if m.editing {
		fmt.Fprintf(&s, "\n  %s\n", color.GreenString("Editing component string for %s:", m.components[m.cursor].Key))
		fmt.Fprintf(&s, "  %s\n", m.textInput.View())
		fmt.Fprintf(&s, "\n  %s\n", color.HiBlackString("(press 'enter' to save, 'esc' to cancel)"))
	} else {
		fmt.Fprintf(&s, "\n  %s\n", color.HiBlackString("(press 'e' to edit, 'd' to delete, 'q' or 'enter' to go back)"))
	}

	return s.String()
}
