package cmd

import (
	"cmdforge/core"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "Manage block ID definitions",
}

// 1. block add
var blockAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a custom block ID to core/data/blocks.json and src/data/blocks.ts",
	Run: func(cmd *cobra.Command, args []string) {
		for {
			var id string
			err := huh.NewInput().
				Title("Block ID to add (e.g. minecraft:my_block):").
				Value(&id).
				Run()
			if err != nil {
				fmt.Println(color.YellowString("Cancelled."))
				return
			}

			id = strings.TrimSpace(id)
			if id == "" {
				fmt.Println(color.YellowString("Cancelled. No ID entered."))
				return
			}

			normalized := id
			if strings.HasPrefix(id, "minecraft:") {
				normalized = strings.TrimPrefix(id, "minecraft:")
			}

			// Check duplicates
			exists := false
			for _, b := range core.Blocks {
				if b == normalized || b == id {
					exists = true
					break
				}
			}

			if exists {
				fmt.Println(color.YellowString("ID %s already exists.", id))
				var addAnother bool
				huh.NewConfirm().
					Title("Add another?").
					Value(&addAnother).
					Run()
				if !addAnother {
					return
				}
				continue
			}

			// Validate ID structure
			matched, _ := regexp.MatchString("^[a-z0-9_:-]+$", id)
			if !matched {
				fmt.Println(color.RedString("Invalid ID format. Allowed: lowercase letters, numbers, underscore, colon, hyphen."))
				continue
			}

			// Add to slice and save
			core.Blocks = append(core.Blocks, normalized)
			if err := saveBlocks(); err != nil {
				fmt.Println(color.RedString("Error saving blocks: %v", err))
				return
			}
			fmt.Println(color.GreenString("Added %s successfully.", id))

			var addAnother bool
			huh.NewConfirm().
				Title("Add another?").
				Value(&addAnother).
				Run()
			if !addAnother {
				return
			}
		}
	},
}

// 2. block list
var blockListCmd = &cobra.Command{
	Use:   "list",
	Short: "List known block IDs",
	Run: func(cmd *cobra.Command, args []string) {
		if len(core.Blocks) == 0 {
			fmt.Println(color.RedString("No blocks data found."))
			return
		}

		fmt.Println(color.GreenString("Known block IDs (%d):", len(core.Blocks)))
		for _, b := range core.Blocks {
			fmt.Printf(" - minecraft:%s\n", b)
		}
	},
}

// 3. block remove
var blockRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a block ID from core/data/blocks.json and src/data/blocks.ts",
	Run: func(cmd *cobra.Command, args []string) {
		if len(core.Blocks) == 0 {
			fmt.Println(color.YellowString("No block IDs to remove."))
			return
		}

		// build choices for autocomplete select
		var choices []huh.Option[string]
		for _, b := range core.Blocks {
			choices = append(choices, huh.NewOption("minecraft:"+b, b))
		}

		var pick string
		err := huh.NewSelect[string]().
			Title("Select block to remove:").
			Options(choices...).
			Value(&pick).
			Filtering(true).
			Run()

		if err != nil {
			fmt.Println(color.YellowString("Cancelled."))
			return
		}

		// filter out the removed block
		newList := make([]string, 0, len(core.Blocks)-1)
		for _, b := range core.Blocks {
			if b != pick {
				newList = append(newList, b)
			}
		}
		core.Blocks = newList

		if err := saveBlocks(); err != nil {
			fmt.Println(color.RedString("Failed to save changes: %v", err))
			return
		}

		fmt.Println(color.GreenString("Removed minecraft:%s successfully.", pick))
	},
}

// 4. block search
var CategoryKeywords = map[string][]string{
	"Wood Blocks": {
		"oak_", "spruce_", "birch_", "jungle_", "acacia_", "dark_oak_",
		"mangrove_", "cherry_", "bamboo_", "planks", "log", "wood", "stripped_",
	},
	"Wood Furniture": {
		"door", "trapdoor", "fence", "fence_gate", "hanging_sign", "sign",
		"button", "pressure_plate", "slab", "stairs",
	},
	"Plants & Foliage": {
		"sapling", "leaves", "azalea", "flower", "fern", "grass",
		"tall_grass", "moss", "vine", "bamboo", "mangrove_propagule",
	},
	"Flowers": {
		"allium", "dandelion", "tulip", "rose", "peony", "blue_orchid",
		"oxeye_daisy", "poppy",
	},
	"Fungi & Mushrooms": {"mushroom", "mushroom_block", "fungus"},
	"Ores & Minerals": {
		"ore", "ancient_debris", "deepslate", "raw_", "diamond_",
		"gold_", "iron_", "lapis_", "redstone_", "emerald_",
	},
	"Stone & Basalt": {
		"stone", "andesite", "diorite", "granite", "tuff", "basalt",
		"deepslate", "cobbled_deepslate",
	},
	"Building Blocks": {
		"stone", "brick", "bricks", "concrete", "concrete_powder",
		"terracotta", "tile", "polished", "smooth", "mosaic", "packed_", "cut_",
	},
	"Slabs/Stairs/Walls": {"slab", "stairs", "wall"},
	"Decorative": {
		"chiseled", "pillar", "carved", "banner", "sculk", "block",
		"bookshelf", "lantern", "torch", "lamp", "beacon",
	},
	"Glass & Stained Glass": {"glass", "stained_glass", "glass_pane"},
	"Wool & Carpets":        {"wool", "carpet", "bed"},
	"Redstone & Mechanisms": {
		"redstone", "comparator", "repeater", "lever", "observer",
		"piston", "sticky_piston", "hopper", "dropper", "dispenser",
		"note_block", "rail", "detector_rail", "powered_rail", "activator_rail",
	},
	"Water/Lava/Natural": {
		"water", "lava", "dirt", "grass_block", "sand", "gravel",
		"clay", "moss", "soul_sand", "soul_soil",
	},
	"Vegetation & Crops": {
		"wheat", "carrots", "potatoes", "beetroots", "sweet_berry",
		"sugar_cane", "cactus", "kelp",
	},
	"Coral & Ocean": {
		"coral", "coral_block", "coral_fan", "coral_wall_fan", "sea_pickle",
		"seagrass",
	},
	"Nether & End": {
		"netherrack", "nether", "end_stone", "end_portal", "end_rod",
		"respawn_anchor",
	},
	"Misc": {
		"chest", "barrel", "anvil", "ladder", "furnace", "enchanting_table",
		"crafting_table", "cauldron",
	},
}

var blockSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search blocks by category and name",
	Run: func(cmd *cobra.Command, args []string) {
		if len(core.Blocks) == 0 {
			fmt.Println(color.RedString("No blocks data found."))
			return
		}

		// build categories list
		categories := []string{"All"}
		var keys []string
		for k := range CategoryKeywords {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		categories = append(categories, keys...)

		for {
			var chosenCat string
			err := huh.NewSelect[string]().
				Title("Choose category to search:").
				Options(huh.NewOptions(categories...)...).
				Value(&chosenCat).
				Run()

			if err != nil {
				fmt.Println(color.YellowString("Search cancelled."))
				return
			}

			// filter blocks
			var filtered []string
			if chosenCat == "All" {
				filtered = core.Blocks
			} else {
				// 💡 修正ポイント: 泥臭いループ処理を core パッケージの純粋関数に丸投げする！
				filtered = core.FilterBlocksByCategory(core.Blocks, CategoryKeywords[chosenCat])
			}

			if len(filtered) == 0 {
				fmt.Println(color.YellowString("No blocks match that category."))
				continue
			}

			// show block search
			var choices []huh.Option[string]
			choices = append(choices, huh.NewOption("<< Back to categories >>", "__BACK__"))
			for _, b := range filtered {
				choices = append(choices, huh.NewOption("minecraft:"+b, b))
			}

			var pick string
			err = huh.NewSelect[string]().
				Title(fmt.Sprintf("Search blocks (%d) — type to filter:", len(filtered))).
				Options(choices...).
				Value(&pick).
				Filtering(true).
				Run()

			if err != nil {
				fmt.Println(color.YellowString("Search cancelled."))
				return
			}

			if pick == "__BACK__" {
				continue
			}

			fullID := "minecraft:" + pick
			fmt.Println(color.GreenString("Selected: %s", fullID))

			var copyToClipboard bool
			err = huh.NewConfirm().
				Title("Copy selected block ID to clipboard?").
				Value(&copyToClipboard).
				Run()

			if err != nil {
				return
			}

			if copyToClipboard {
				if err := clipboard.WriteAll(fullID); err != nil {
					fmt.Println(color.RedString("✗ Failed to copy to clipboard"))
				} else {
					fmt.Println(color.GreenString("✓ Copied to clipboard: %s", fullID))
				}
			} else {
				fmt.Println(color.BlueString("Skipped copying to clipboard."))
			}
			return
		}
	},
}

func init() {
	blockCmd.AddCommand(blockAddCmd)
	blockCmd.AddCommand(blockListCmd)
	blockCmd.AddCommand(blockRemoveCmd)
	blockCmd.AddCommand(blockSearchCmd)
	rootCmd.AddCommand(blockCmd)
}

func saveBlocks() error {
	sort.Strings(core.Blocks)

	// JSON
	jsonBytes, err := json.MarshalIndent(core.Blocks, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile("core/data/blocks.json", jsonBytes, 0644)
	if err != nil {
		return err
	}

	// TS
	tsContent := fmt.Sprintf("// Auto-generated by CLI block command\nexport const BLOCKS = %s as string[];\nexport default BLOCKS;\n", string(jsonBytes))
	err = os.WriteFile("src/data/blocks.ts", []byte(tsContent), 0644)
	if err != nil {
		return err
	}

	return nil
}
