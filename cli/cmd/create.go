package cmd

import (
	"cmdforge/core"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	copyFlag    bool
	silentFlag  bool
	noSlashFlag bool
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Generate Minecraft commands interactively",
	Run: func(cmd *cobra.Command, args []string) {
		greenBold := color.New(color.FgGreen, color.Bold)
		blueBold := color.New(color.FgBlue, color.Bold)

		if copyFlag {
			fmt.Println(color.BlueString("INFO"), greenBold.Sprint("The command will be copied to clipboard"))
		} else {
			fmt.Println(color.BlueString("INFO"), greenBold.Sprint("The command will not be copied to clipboard"))
		}

		if silentFlag {
			fmt.Println(color.YellowString("WARN"), color.YellowString("Notification will not be sent when the command is copied"))
		}

		if noSlashFlag {
			fmt.Println(color.BlueString("INFO"), color.YellowString("Slash (\"/\") will not be added to the command"))
		}

		supportedTypes := []string{
			"give",
			"teleport",
			"setblock",
			"fill",
			"say",
			"execute",
			"item",
			"effect",
			"enchant",
		}

		var commandType string
		err := huh.NewSelect[string]().
			Title("Select a command type:").
			Options(huh.NewOptions(supportedTypes...)...).
			Value(&commandType).
			Run()

		if err != nil {
			fmt.Println(color.YellowString("Cancelled."))
			return
		}

		fmt.Println(color.BlueString("Generate target:"), blueBold.Sprint(commandType))
		fmt.Println()

		var generatedCommand string

		switch commandType {
		case "give":
			selector, err := addSelectorsQuestion()
			if err != nil {
				return
			}
			fmt.Println(color.BlueString("Target selector:"), greenBold.Sprint(selector))

			item, err := selectItem()
			if err != nil {
				return
			}

			// Format items and amount
			parts := strings.SplitN(item, " ", 2)
			itemPart := parts[0]
			amountPart := "1"
			if len(parts) > 1 {
				amountPart = parts[1]
			}

			giveBuilder := core.GiveCommand{
				Selector: selector,
				Item:     itemPart,
				Amount:   amountPart,
			}
			// Parse components if present
			if idx := strings.Index(itemPart, "["); idx != -1 {
				giveBuilder.Item = itemPart[:idx]
				giveBuilder.Components = itemPart[idx+1 : len(itemPart)-1]
			}

			generatedCommand = giveBuilder.Build()

		case "teleport":
			tpOptions := []string{
				"1. tp <destination> - Teleport to a specific entity",
				"2. tp <targets> <destination> - Teleport specific targets to a specific entity",
				"3. tp <location> - Teleport to a specific location",
				"4. tp <targets> <location> - Teleport specific targets to a specific location",
				"5. tp <targets> <location> <rotation> - Teleport specific targets to a specific location with rotation(angle)",
				"6. tp <targets> <location> facing <facingLocation> - Teleport specific targets to a specific location with rotation(coordinate)",
				"7. tp <targets> <location> facing entity <facingEntity> [facingAnchor] - Teleport specific targets to a specific location with rotation(facing entity)",
			}

			var selectedTpOption string
			err = huh.NewSelect[string]().
				Title("Teleport Command type:").
				Options(huh.NewOptions(tpOptions...)...).
				Value(&selectedTpOption).
				Run()
			if err != nil {
				return
			}

			tpTypeStr := strings.Split(selectedTpOption, ".")[0]
			fmt.Println(color.BlueString("Selected teleport type:"), greenBold.Sprint(selectedTpOption))

			var targets, destination, location, rotation, facingLocation, facingEntity, facingAnchor string

			if contains([]string{"2", "4", "5", "6", "7"}, tpTypeStr) {
				fmt.Println(color.BlackString("<targets> - Specify the entity to teleport"))
				targets, err = addSelectorsQuestion()
				if err != nil {
					return
				}
				fmt.Println(color.BlueString("<targets>:"), greenBold.Sprint(targets))
			}

			if contains([]string{"1", "2"}, tpTypeStr) {
				fmt.Println(color.BlackString("<destination> - Specifies the entity to teleport to"))
				destination, err = addSelectorsQuestion()
				if err != nil {
					return
				}
				fmt.Println(color.BlueString("<destination>:"), greenBold.Sprint(destination))
			}

			if contains([]string{"3", "4", "5", "6", "7"}, tpTypeStr) {
				fmt.Println(color.BlackString("<location> - Specifies the location (e.g., ~ ~ ~, 0 64 0)"))
				location, err = promptCoordinates("Enter location coordinates:")
				if err != nil {
					return
				}
				fmt.Println(color.BlueString("<location>:"), greenBold.Sprint(location))
			}

			if tpTypeStr == "5" {
				fmt.Println(color.BlackString("<rotation> - Specify the direction (yaw and pitch)"))
				rotation, err = promptRotation()
				if err != nil {
					return
				}
				fmt.Println(color.BlueString("<rotation>:"), greenBold.Sprint(rotation))
			}

			if tpTypeStr == "6" {
				fmt.Println(color.BlackString("<facingLocation> - Specifies coordinates to face"))
				facingLocation, err = promptCoordinates("Enter facing coordinates:")
				if err != nil {
					return
				}
				fmt.Println(color.BlueString("<facingLocation>:"), greenBold.Sprint(facingLocation))
			}

			if tpTypeStr == "7" {
				fmt.Println(color.BlackString("<facingEntity> - Specifies entity to face"))
				facingEntity, err = addSelectorsQuestion()
				if err != nil {
					return
				}
				fmt.Println(color.BlueString("<facingEntity>:"), greenBold.Sprint(facingEntity))

				err = huh.NewSelect[string]().
					Title("Facing Anchor (optional):").
					Options(
						huh.NewOption("eyes", "eyes"),
						huh.NewOption("feet", "feet"),
						huh.NewOption("skip (do not specify anchor)", ""),
					).
					Value(&facingAnchor).
					Run()
				if err != nil {
					return
				}
			}

			builder := core.TeleportCommand{
				Type:           tpTypeStr,
				Targets:        targets,
				Destination:    destination,
				Location:       location,
				Rotation:       rotation,
				FacingLocation: facingLocation,
				FacingEntity:   facingEntity,
				FacingAnchor:   facingAnchor,
			}
			generatedCommand = builder.Build()

		case "setblock":
			pos, err := promptCoordinates("Enter location coordinates:")
			if err != nil {
				return
			}
			fmt.Println(color.BlueString("Position:"), greenBold.Sprint(pos))

			var block string
			err = huh.NewSelect[string]().
				Title("Block (type to filter e.g. diamond_block):").
				Options(buildSelectOptions(core.Blocks)...).
				Value(&block).
				Filtering(true).
				Run()
			if err != nil {
				return
			}

			sbOptions := []string{
				"destroy - Destroys block dropping items",
				"keep - Places only if original is air",
				"replace - Simply places block",
				"strict - Places without trigger update",
				"Skip - Use default replace",
			}

			var selectedSbOption string
			err = huh.NewSelect[string]().
				Title("Setblock Option:").
				Options(huh.NewOptions(sbOptions...)...).
				Value(&selectedSbOption).
				Run()
			if err != nil {
				return
			}

			opt := strings.Split(selectedSbOption, " - ")[0]
			if opt == "Skip" {
				opt = "replace"
			}

			sbBuilder := core.SetblockCommand{
				Position: pos,
				Block:    block,
				Option:   opt,
			}
			generatedCommand = sbBuilder.Build()

		case "fill":
			from, err := promptCoordinates("Enter starting block coordinates:")
			if err != nil {
				return
			}
			fmt.Println(color.BlueString("Position from:"), greenBold.Sprint(from))

			to, err := promptCoordinates("Enter ending block coordinates:")
			if err != nil {
				return
			}
			fmt.Println(color.BlueString("Position to:"), greenBold.Sprint(to))

			var block string
			err = huh.NewSelect[string]().
				Title("Block (type to filter):").
				Options(buildSelectOptions(core.Blocks)...).
				Value(&block).
				Filtering(true).
				Run()
			if err != nil {
				return
			}

			fillBuilder := core.FillCommand{
				From:  from,
				To:    to,
				Block: block,
			}
			generatedCommand = fillBuilder.Build()

		case "say":
			var message string
			err = huh.NewInput().
				Title("Message:").
				Value(&message).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("please enter a message")
					}
					return nil
				}).
				Run()
			if err != nil {
				return
			}

			sayBuilder := core.SayCommand{Message: message}
			generatedCommand = sayBuilder.Build()

		case "execute":
			var target, command string
			err = huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Target selector (e.g., @a):").
						Value(&target).
						Validate(func(s string) error {
							if strings.TrimSpace(s) == "" {
								return fmt.Errorf("target cannot be empty")
							}
							return nil
						}),
					huh.NewInput().
						Title("Command to execute:").
						Value(&command).
						Validate(func(s string) error {
							if strings.TrimSpace(s) == "" {
								return fmt.Errorf("command cannot be empty")
							}
							return nil
						}),
				),
			).Run()
			if err != nil {
				return
			}

			execBuilder := core.ExecuteCommand{
				Target:  target,
				Command: command,
			}
			generatedCommand = execBuilder.Build()

		case "item":
			fmt.Println(color.YellowString("WARN"), "The \"item\" command generator only supports \"replace\" subcommand.")

			var replaceTarget string
			err = huh.NewSelect[string]().
				Title("Replace target:").
				Options(huh.NewOptions("block", "entity")...).
				Value(&replaceTarget).
				Run()
			if err != nil {
				return
			}

			var targetPosOrSel string
			if replaceTarget == "block" {
				targetPosOrSel, err = promptCoordinates("Enter coordinates of the target block:")
				if err != nil {
					return
				}
			} else {
				targetPosOrSel, err = addSelectorsQuestion()
				if err != nil {
					return
				}
			}

			slot, err := getSlot()
			if err != nil {
				return
			}
			fmt.Println(color.BlueString("Selected Slot:"), greenBold.Sprint(slot))

			var replaceOption string
			err = huh.NewSelect[string]().
				Title("Item Command Option:").
				Options(
					huh.NewOption("with - Replaces item in slot with specified item", "with"),
					huh.NewOption("from - Copies item from specified slot to target slot", "from"),
				).
				Value(&replaceOption).
				Run()
			if err != nil {
				return
			}

			itemBuilder := core.ItemCommand{
				TargetType:     replaceTarget,
				TargetPosOrSel: targetPosOrSel,
				Slot:           slot,
			}

			if replaceOption == "with" {
				itemStr, err := selectItem()
				if err != nil {
					return
				}
				itemBuilder.Item = itemStr
				itemBuilder.IsFrom = false
			} else {
				var sourceTarget string
				err = huh.NewSelect[string]().
					Title("Source target:").
					Options(huh.NewOptions("block", "entity")...).
					Value(&sourceTarget).
					Run()
				if err != nil {
					return
				}

				var sourcePosOrSel string
				if sourceTarget == "block" {
					sourcePosOrSel, err = promptCoordinates("Enter coordinates of the source block:")
					if err != nil {
						return
					}
				} else {
					sourcePosOrSel, err = addSelectorsQuestion()
					if err != nil {
						return
					}
				}

				srcSlot, err := getSlot()
				if err != nil {
					return
				}

				itemBuilder.SourceType = sourceTarget
				itemBuilder.SourcePosOrSel = sourcePosOrSel
				itemBuilder.SourceSlot = srcSlot
				itemBuilder.IsFrom = true
			}

			generatedCommand = itemBuilder.Build()

		case "effect":
			var effType string
			err = huh.NewSelect[string]().
				Title("Effect type:").
				Options(
					huh.NewOption("give - give an effect to entity", "give"),
					huh.NewOption("clear - clear effect(s)", "clear"),
				).
				Value(&effType).
				Run()
			if err != nil {
				return
			}

			target, err := addSelectorsQuestion()
			if err != nil {
				return
			}

			effBuilder := core.EffectCommand{
				Type:   effType,
				Target: target,
			}

			if effType == "give" {
				var effName string
				err = huh.NewSelect[string]().
					Title("Select effect:").
					Options(buildSelectOptions(core.Effects)...).
					Value(&effName).
					Filtering(true).
					Run()
				if err != nil {
					return
				}
				effBuilder.Effect = effName

				var durOption string
				err = huh.NewSelect[string]().
					Title("Specify duration option:").
					Options(
						huh.NewOption("infinite - Set the duration to infinite", "infinite"),
						huh.NewOption("seconds - Set the duration in seconds", "seconds"),
					).
					Value(&durOption).
					Run()
				if err != nil {
					return
				}

				if durOption == "seconds" {
					var secStr string
					err = huh.NewInput().
						Title("Duration in seconds:").
						Value(&secStr).
						Validate(func(s string) error {
							if val, err := strconv.Atoi(s); err != nil || val <= 0 {
								return fmt.Errorf("duration must be a positive integer")
							}
							return nil
						}).
						Run()
					if err != nil {
						return
					}
					effBuilder.Duration = secStr
				} else {
					effBuilder.Duration = "infinite"
				}

				var ampStr string
				err = huh.NewInput().
					Title("Amplifier level (0 for level 1, 0-255):").
					Value(&ampStr).
					Validate(func(s string) error {
						val, err := strconv.Atoi(s)
						if err != nil || val < 0 || val > 255 {
							return fmt.Errorf("amplifier must be between 0 and 255")
						}
						return nil
					}).
					Run()
				if err != nil {
					return
				}
				effBuilder.Amplifier = ampStr

				var hideParticles bool
				err = huh.NewConfirm().
					Title("Do you want to hide effect particles?").
					Value(&hideParticles).
					Run()
				if err != nil {
					return
				}
				effBuilder.HideParticles = hideParticles

			} else {
				var clearAll bool
				err = huh.NewConfirm().
					Title("Clear all effects?").
					Value(&clearAll).
					Run()
				if err != nil {
					return
				}

				if clearAll {
					effBuilder.ClearAll = true
				} else {
					var effName string
					err = huh.NewSelect[string]().
						Title("Select effect to clear:").
						Options(buildSelectOptions(core.Effects)...).
						Value(&effName).
						Filtering(true).
						Run()
					if err != nil {
						return
					}
					effBuilder.Effect = effName
				}
			}

			generatedCommand = effBuilder.Build()

		case "enchant":
			target, err := addSelectorsQuestion()
			if err != nil {
				return
			}

			var enchantment string
			var choices []huh.Option[string]
			for _, e := range core.Enchantments {
				choices = append(choices, huh.NewOption(e.Name, e.Name))
			}

			err = huh.NewSelect[string]().
				Title("Enchantment name:").
				Options(choices...).
				Value(&enchantment).
				Filtering(true).
				Run()
			if err != nil {
				return
			}

			// Find max level
			maxLevel := 1
			for _, e := range core.Enchantments {
				if e.Name == enchantment {
					maxLevel = e.MaxLevel
					break
				}
			}

			var enchantLevel string
			if maxLevel == 1 {
				enchantLevel = "1"
			} else {
				err = huh.NewInput().
					Title(fmt.Sprintf("Enchantment level (1~%d):", maxLevel)).
					Value(&enchantLevel).
					Validate(func(s string) error {
						val, err := strconv.Atoi(s)
						if err != nil || val < 1 || val > maxLevel {
							return fmt.Errorf("level must be between 1 and %d", maxLevel)
						}
						return nil
					}).
					Run()
				if err != nil {
					return
				}
			}

			enchBuilder := core.EnchantCommand{
				Target:      target,
				Enchantment: enchantment,
				Level:       enchantLevel,
			}
			generatedCommand = enchBuilder.Build()
		}

		finalCommand := core.FormatCommand(generatedCommand, !noSlashFlag)
		fmt.Println()
		fmt.Printf("%s %s %s\n\n", color.GreenString("✓"), color.New(color.FgGreen, color.Bold).Sprint("Generated! Command:"), color.New(color.FgBlue).Sprint(finalCommand))

		if copyFlag {
			var dummy string
			err = huh.NewInput().
				Title("Press Enter to copy to clipboard...").
				Value(&dummy).
				Run()
			if err != nil {
				return
			}

			if err := clipboard.WriteAll(finalCommand); err != nil {
				fmt.Println(color.RedString("✗ Failed to copy command to clipboard"))
			} else {
				fmt.Println(color.GreenString("✓ Command copied to clipboard!"))
			}
		}
	},
}

func init() {
	createCmd.Flags().BoolVarP(&copyFlag, "copy", "c", true, "Whether to copy command to clipboard")
	createCmd.Flags().BoolVarP(&silentFlag, "silent", "s", false, "Whether to notify when the command is copied")
	createCmd.Flags().BoolVar(&noSlashFlag, "no-slash", false, "Remove the leading slash (\"/\")")
	rootCmd.AddCommand(createCmd)
}

func contains(arr []string, target string) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}
	return false
}

func promptCoordinates(title string) (string, error) {
	var x, y, z string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("X coordinate").Value(&x).Placeholder("~"),
			huh.NewInput().Title("Y coordinate").Value(&y).Placeholder("~"),
			huh.NewInput().Title("Z coordinate").Value(&z).Placeholder("~"),
		),
	).Run()
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(x) == "" {
		x = "~"
	}
	if strings.TrimSpace(y) == "" {
		y = "~"
	}
	if strings.TrimSpace(z) == "" {
		z = "~"
	}

	return fmt.Sprintf("%s %s %s", x, y, z), nil
}

func promptRotation() (string, error) {
	var yaw, pitch string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("yaw (horizontal rotation)").Value(&yaw).Placeholder("90"),
			huh.NewInput().Title("pitch (vertical rotation)").Value(&pitch).Placeholder("0"),
		),
	).Run()
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(yaw) == "" {
		yaw = "90"
	}
	if strings.TrimSpace(pitch) == "" {
		pitch = "0"
	}

	return fmt.Sprintf("%s %s", yaw, pitch), nil
}

func mojangPlayerExists(name string) bool {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.mojang.com/users/profiles/minecraft/" + name)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
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
				// verify player with Mojang API
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

		fmt.Println(color.BlueString("Target:"), color.GreenString(targetType))

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

		var addedSelectors []string

		for {
			var available []string
			for _, s := range allSelectorTypes {
				alreadyAdded := false
				for _, added := range addedSelectors {
					if strings.HasPrefix(added, s+"=") {
						alreadyAdded = true
						break
					}
				}
				if !alreadyAdded {
					available = append(available, s)
				}
			}

			selectorChoices := []string{}
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
			for _, av := range available {
				selectorChoices = append(selectorChoices, fmt.Sprintf("%s - %s", av, descriptions[av]))
			}
			selectorChoices = append(selectorChoices, "OK")

			var selType string
			err = huh.NewSelect[string]().
				Title("Additional selectors (Select OK to finish):").
				Options(huh.NewOptions(selectorChoices...)...).
				Value(&selType).
				Run()
			if err != nil {
				return "", err
			}

			selKey := strings.Split(selType, " ")[0]
			if selKey == "OK" {
				break
			}

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
				addedSelectors = append(addedSelectors, fmt.Sprintf("%s=%s", selKey, val))
			}
		}

		if len(addedSelectors) > 0 {
			targetType = fmt.Sprintf("%s[%s]", targetType, strings.Join(addedSelectors, ","))
		}
		return targetType, nil
	}
}

func selectItem() (string, error) {
	var itemName string
	err := huh.NewSelect[string]().
		Title("Item (type to filter e.g., diamond_sword):").
		Options(buildSelectOptions(core.Items)...).
		Value(&itemName).
		Filtering(true).
		Run()
	if err != nil {
		return "", err
	}

	var addComp bool
	err = huh.NewConfirm().
		Title("Add item component(s)?").
		Value(&addComp).
		Run()
	if err != nil {
		return "", err
	}

	componentsStr := ""
	if addComp {
		componentsStr, err = addItemComponentsQuestion()
		if err != nil {
			return "", err
		}
	}

	var amount string
	err = huh.NewInput().
		Title("Item amount (Press enter for default 1):").
		Value(&amount).
		Validate(func(s string) error {
			s = strings.TrimSpace(s)
			if s == "" {
				return nil
			}
			if _, err := strconv.Atoi(s); err != nil {
				return fmt.Errorf("amount must be a positive integer")
			}
			return nil
		}).
		Run()
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(amount) == "" {
		amount = "1"
	}

	if componentsStr != "" {
		return fmt.Sprintf("%s[%s] %s", itemName, componentsStr, amount), nil
	}
	return fmt.Sprintf("%s %s", itemName, amount), nil
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
	var addedComponents []string

	for {
		available := []string{}
		for k := range ComponentDescriptions {
			already := false
			for _, added := range addedComponents {
				if strings.HasPrefix(added, k+"=") {
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
		choices = append(choices, "OK")

		var selComp string
		err := huh.NewSelect[string]().
			Title("Additional components (Select OK to finish):").
			Options(huh.NewOptions(choices...)...).
			Value(&selComp).
			Run()
		if err != nil {
			return "", err
		}

		compKey := strings.Split(selComp, " ")[0]
		if compKey == "OK" {
			break
		}

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
			addedComponents = append(addedComponents, compVal)
		}
	}

	return strings.Join(addedComponents, ","), nil
}

func getSlot() (string, error) {
	slotOptions := []struct {
		slot, desc string
		hasNum     bool
		min, max   int
	}{
		{"contents", "An entity with one slot (item_frame etc.)", false, 0, 0},
		{"container", "A block with multiple slots (chest, hopper etc.)", true, 0, 53},
		{"hotbar", "Player hotbar slots", true, 0, 8},
		{"inventory", "Player inventory slots", true, 0, 26},
		{"enderchest", "Ender chest slots", true, 0, 26},
		{"villager", "Villager trading slots", true, 0, 7},
		{"player.crafting", "Player crafting grid slots", true, 0, 3},
		{"horse", "Horse inventory slots", true, 0, 14},
		{"weapon", "Player weapon slots", false, 0, 0},
		{"armor", "Player armor slots", false, 0, 0},
		{"saddle", "Saddle slot", false, 0, 0},
		{"horse.chest", "Horse chest slots", false, 0, 0},
		{"player.cursor", "Item held by cursor", false, 0, 0},
	}

	choices := []huh.Option[string]{}
	for _, opt := range slotOptions {
		choices = append(choices, huh.NewOption(fmt.Sprintf("%s - %s", opt.slot, opt.desc), opt.slot))
	}

	var slotKey string
	err := huh.NewSelect[string]().
		Title("Select slot type:").
		Options(choices...).
		Value(&slotKey).
		Run()
	if err != nil {
		return "", err
	}

	// Find the configuration
	var selectedOpt struct {
		slot, desc string
		hasNum     bool
		min, max   int
	}
	for _, opt := range slotOptions {
		if opt.slot == slotKey {
			selectedOpt = opt
			break
		}
	}

	slotPart := ""

	if selectedOpt.hasNum {
		var numStr string
		err = huh.NewInput().
			Title(fmt.Sprintf("Enter slot number (%d-%d):", selectedOpt.min, selectedOpt.max)).
			Value(&numStr).
			Validate(func(s string) error {
				val, err := strconv.Atoi(s)
				if err != nil || val < selectedOpt.min || val > selectedOpt.max {
					return fmt.Errorf("must be between %d and %d", selectedOpt.min, selectedOpt.max)
				}
				return nil
			}).
			Run()
		if err != nil {
			return "", err
		}
		slotPart = "." + numStr
	} else if slotKey == "weapon" {
		var weaponHand string
		err = huh.NewSelect[string]().
			Title("Select weapon slot:").
			Options(huh.NewOptions("mainhand", "offhand")...).
			Value(&weaponHand).
			Run()
		if err != nil {
			return "", err
		}
		slotPart = "." + weaponHand
	} else if slotKey == "armor" {
		var armorSlot string
		err = huh.NewSelect[string]().
			Title("Select armor slot:").
			Options(huh.NewOptions("head", "chest", "legs", "feet")...).
			Value(&armorSlot).
			Run()
		if err != nil {
			return "", err
		}
		slotPart = "." + armorSlot
	}

	return slotKey + slotPart, nil
}

func buildSelectOptions(list []string) []huh.Option[string] {
	var choices []huh.Option[string]
	for _, val := range list {
		lbl := val
		if !strings.HasPrefix(lbl, "minecraft:") {
			lbl = "minecraft:" + lbl
		}
		choices = append(choices, huh.NewOption(lbl, val))
	}
	return choices
}
