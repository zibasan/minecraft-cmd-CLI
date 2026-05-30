package cmd

import (
	"cmdforge/core"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
)

// GetFilteredNbtOptions filters the master list of NBT tags based on the selected entity
// and returns a list of huh.Option[string] for the multi-select UI.
func GetFilteredNbtOptions(selectedEntity string) []huh.Option[string] {
	var options []huh.Option[string]
	for _, opt := range core.NbtMasterList {
		// If ApplicableTo is empty, it is a common option.
		// Otherwise, check if selectedEntity is in the ApplicableTo list.
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
			options = append(options, huh.NewOption(opt.Label, opt.Key))
		}
	}
	return options
}

// PromptEntityNbt displays the interactive multi-select UI for boolean NBT tags
// and optional custom NBT strings, then merges them.
func PromptEntityNbt(selectedEntity string) (string, error) {
	opts := GetFilteredNbtOptions(selectedEntity)

	var selectedKeys []string
	if len(opts) > 0 {
		err := huh.NewMultiSelect[string]().
			Title("Select entity NBT properties (Space to select, Enter to confirm):").
			Options(opts...).
			Value(&selectedKeys).
			Height(12).
			Run()
		if err != nil {
			return "", err
		}
	}

	assembled := core.AssembleNbtString(selectedKeys)
	if assembled != "" {
		fmt.Println(color.BlueString("Selected NBT tags:"), color.New(color.FgGreen).Sprint(assembled))
	}

	// Ask to add a custom raw NBT tag
	var addCustom bool
	err := huh.NewConfirm().
		Title("Do you want to add a custom NBT tag (e.g. Attributes)?").
		Value(&addCustom).
		Run()
	if err != nil {
		return "", err
	}

	var customNbt string
	if addCustom {
		err = huh.NewInput().
			Title("Enter custom NBT (e.g. Attributes:[{Name:\"generic.max_health\",Base:20f}]):").
			Value(&customNbt).
			Run()
		if err != nil {
			return "", err
		}
	}

	return MergeNbtStrings(assembled, customNbt), nil
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
