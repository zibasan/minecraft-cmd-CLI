package cmd

import (
	"cmdforge/core"
	"fmt"
	"strconv"
	"strings"

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

			var addedChoices []huh.Option[int]
			for idx, cfg := range configuredCustom {
				label := fmt.Sprintf("%s: %s", cfg.Key, cfg.Value)
				addedChoices = append(addedChoices, huh.NewOption(label, idx))
			}
			addedChoices = append(addedChoices, huh.NewOption("<< Back >>", -1))

			var selectedIdx int
			err = huh.NewSelect[int]().
				Title("Configure Added NBT Tags (Select a tag to edit/delete):").
				Options(addedChoices...).
				Value(&selectedIdx).
				Run()
			if err != nil {
				return "", err
			}

			if selectedIdx == -1 {
				continue
			}

			selectedTag := configuredCustom[selectedIdx]

			var action string
			err = huh.NewSelect[string]().
				Title(fmt.Sprintf("Action for '%s: %s':", selectedTag.Key, selectedTag.Value)).
				Options(
					huh.NewOption("Edit (e)", "edit"),
					huh.NewOption("Delete (d)", "delete"),
					huh.NewOption("<< Cancel >>", "cancel"),
				).
				Value(&action).
				Run()
			if err != nil {
				return "", err
			}

			if action == "edit" {
				var newVal string
				inputTitle := fmt.Sprintf("Enter new value for %s:", selectedTag.Key)
				err = huh.NewInput().
					Title(inputTitle).
					Value(&newVal).
					Run()
				if err != nil {
					return "", err
				}
				newVal = strings.TrimSpace(newVal)
				if newVal != "" {
					if selectedTag.Type == "string" {
						if !strings.HasPrefix(newVal, "'") && !strings.HasPrefix(newVal, "\"") {
							newVal = fmt.Sprintf(`'"%s"'`, newVal)
						}
					}
					configuredCustom[selectedIdx].Value = newVal
					fmt.Println(color.GreenString("Updated %s to %s", selectedTag.Key, newVal))
				}
			} else if action == "delete" {
				configuredCustom = append(configuredCustom[:selectedIdx], configuredCustom[selectedIdx+1:]...)
				fmt.Println(color.GreenString("Deleted tag %s. It is now back in the available list.", selectedTag.Key))
			}

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
			if targetOpt.Type == "string" {
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
			} else if targetOpt.Type == "int" {
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
			} else if targetOpt.Type == "raw" {
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
