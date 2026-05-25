package core

import (
	"fmt"
	"regexp"
	"strings"
)

// IsValidPositionToken checks if a coordinate token is valid (e.g. ~ ~5 -10 ^ ^3).
func IsValidPositionToken(token string) bool {
	re := regexp.MustCompile(`^(?:[~^]-?\d+|[~^]|-?\d+)$`)
	return re.MatchString(token)
}

// IsValidPosition checks if a position string contains exactly 3 valid coordinate tokens.
func IsValidPosition(pos string) bool {
	tokens := strings.Fields(pos)
	if len(tokens) != 3 {
		return false
	}
	for _, t := range tokens {
		if !IsValidPositionToken(t) {
			return false
		}
	}
	return true
}

// FormatCommand adds or removes the leading slash based on preferences.
func FormatCommand(cmd string, includeSlash bool) string {
	if includeSlash {
		if !strings.HasPrefix(cmd, "/") {
			return "/" + cmd
		}
		return cmd
	}
	return strings.TrimPrefix(cmd, "/")
}

// 1. Give Command
type GiveCommand struct {
	Selector   string
	Item       string
	Components string // e.g. "item_name='\"Sword\"',damage=10"
	Amount     string
}

func (g *GiveCommand) Build() string {
	itemStr := g.Item
	if g.Components != "" {
		itemStr = fmt.Sprintf("%s[%s]", g.Item, g.Components)
	}
	amountStr := "1"
	if g.Amount != "" {
		amountStr = g.Amount
	}
	return fmt.Sprintf("give %s %s %s", g.Selector, itemStr, amountStr)
}

// 2. Teleport Command
type TeleportCommand struct {
	Type           string // "1" to "7"
	Targets        string
	Destination    string
	Location       string
	Rotation       string
	FacingLocation string
	FacingEntity   string
	FacingAnchor   string
}

func (t *TeleportCommand) Build() string {
	switch t.Type {
	case "1":
		return fmt.Sprintf("teleport %s", t.Destination)
	case "2":
		return fmt.Sprintf("teleport %s %s", t.Targets, t.Destination)
	case "3":
		return fmt.Sprintf("teleport %s", t.Location)
	case "4":
		return fmt.Sprintf("teleport %s %s", t.Targets, t.Location)
	case "5":
		return fmt.Sprintf("teleport %s %s %s", t.Targets, t.Location, t.Rotation)
	case "6":
		return fmt.Sprintf("teleport %s %s facing %s", t.Targets, t.Location, t.FacingLocation)
	case "7":
		anchorStr := ""
		if t.FacingAnchor != "" {
			anchorStr = " " + t.FacingAnchor
		}
		return fmt.Sprintf("teleport %s %s facing entity %s%s", t.Targets, t.Location, t.FacingEntity, anchorStr)
	default:
		return ""
	}
}

// 3. Setblock Command
type SetblockCommand struct {
	Position string
	Block    string
	Option   string // destroy, keep, replace, strict, etc.
}

func (s *SetblockCommand) Build() string {
	blockStr := s.Block
	if !strings.HasPrefix(blockStr, "minecraft:") {
		blockStr = "minecraft:" + blockStr
	}
	optStr := ""
	if s.Option != "" {
		optStr = " " + s.Option
	}
	return fmt.Sprintf("setblock %s %s%s", s.Position, blockStr, optStr)
}

// 4. Fill Command
type FillCommand struct {
	From  string
	To    string
	Block string
}

func (f *FillCommand) Build() string {
	blockStr := f.Block
	if !strings.HasPrefix(blockStr, "minecraft:") {
		blockStr = "minecraft:" + blockStr
	}
	return fmt.Sprintf("fill %s %s %s", f.From, f.To, blockStr)
}

// 5. Say Command
type SayCommand struct {
	Message string
}

func (s *SayCommand) Build() string {
	return fmt.Sprintf("say %s", s.Message)
}

// 6. Execute Command
type ExecuteCommand struct {
	Target  string
	Command string
}

func (e *ExecuteCommand) Build() string {
	return fmt.Sprintf("execute as %s at @s run %s", e.Target, e.Command)
}

// 7. Item Command
type ItemCommand struct {
	TargetType       string // "block" or "entity"
	TargetPosOrSel   string // coordinates for block, selector for entity
	Slot             string
	SourceType       string // "block" or "entity"
	SourcePosOrSel   string // coordinates for block, selector for entity
	SourceSlot       string
	Item             string // only used if replacing with item
	IsFrom           bool   // true: replace from ..., false: replace with ...
}

func (i *ItemCommand) Build() string {
	target := fmt.Sprintf("%s %s", i.TargetType, i.TargetPosOrSel)

	if i.IsFrom {
		source := fmt.Sprintf("%s %s", i.SourceType, i.SourcePosOrSel)
		return fmt.Sprintf("item replace %s %s from %s %s", target, i.Slot, source, i.SourceSlot)
	}

	return fmt.Sprintf("item replace %s %s with %s", target, i.Slot, i.Item)
}

// 8. Effect Command
type EffectCommand struct {
	Type          string // "give" or "clear"
	Target        string
	Effect        string // used in give or clear specific
	Duration      string // used in give
	Amplifier     string // used in give
	HideParticles bool   // used in give
	ClearAll      bool   // used in clear
}

func (e *EffectCommand) Build() string {
	if e.Type == "clear" {
		if e.ClearAll {
			return fmt.Sprintf("effect clear %s", e.Target)
		}
		effectStr := e.Effect
		if !strings.HasPrefix(effectStr, "minecraft:") {
			effectStr = "minecraft:" + effectStr
		}
		return fmt.Sprintf("effect clear %s %s", e.Target, effectStr)
	}

	effectStr := e.Effect
	if !strings.HasPrefix(effectStr, "minecraft:") {
		effectStr = "minecraft:" + effectStr
	}

	durStr := "infinite"
	if e.Duration != "" {
		durStr = e.Duration
	}

	ampStr := "0"
	if e.Amplifier != "" {
		ampStr = e.Amplifier
	}

	particleStr := "false"
	if e.HideParticles {
		particleStr = "true"
	}

	return fmt.Sprintf("effect give %s %s %s %s %s", e.Target, effectStr, durStr, ampStr, particleStr)
}

// 9. Enchant Command
type EnchantCommand struct {
	Target      string
	Enchantment string
	Level       string
}

func (e *EnchantCommand) Build() string {
	enchStr := e.Enchantment
	if !strings.HasPrefix(enchStr, "minecraft:") {
		enchStr = "minecraft:" + enchStr
	}
	return fmt.Sprintf("enchant %s %s %s", e.Target, enchStr, e.Level)
}
