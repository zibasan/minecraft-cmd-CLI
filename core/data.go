package core

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
)

type BlockData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func LoadBlocks(filePath string) ([]BlockData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var blocks []BlockData
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&blocks)
	if err != nil {
		return nil, err
	}

	return blocks, nil
}

//go:embed data/*.json
var dataFS embed.FS

type EnchantmentData struct {
	Name     string `json:"name"`
	MaxLevel int    `json:"maxLevel"`
}

type BrigadierNode struct {
	Type       string                    `json:"type"`
	Parser     string                    `json:"parser,omitempty"`
	Executable bool                      `json:"executable,omitempty"`
	Properties map[string]interface{}    `json:"properties,omitempty"`
	Children   map[string]*BrigadierNode `json:"children,omitempty"`
}

func (n *BrigadierNode) GetRegistry() string {
	if n.Properties == nil {
		return ""
	}
	if reg, exists := n.Properties["registry"]; exists {
		if str, ok := reg.(string); ok {
			return str
		}
	}
	return ""
}

var (
	Blocks       []string
	Effects      []string
	Enchantments []EnchantmentData
	Items        []string
	Sounds       []string
	Entities     []string
	CommandTree  BrigadierNode
)

// LoadData reads and decodes the embedded JSON files into memory.
func LoadData() error {
	// 1. Blocks
	blockBytes, err := dataFS.ReadFile("data/blocks.json")
	if err != nil {
		return fmt.Errorf("failed to read blocks.json: %w", err)
	}
	if err := json.Unmarshal(blockBytes, &Blocks); err != nil {
		return fmt.Errorf("failed to parse blocks.json: %w", err)
	}

	// 2. Effects
	effectBytes, err := dataFS.ReadFile("data/effects.json")
	if err != nil {
		return fmt.Errorf("failed to read effects.json: %w", err)
	}
	if err := json.Unmarshal(effectBytes, &Effects); err != nil {
		return fmt.Errorf("failed to parse effects.json: %w", err)
	}

	// 3. Enchantments
	enchantmentBytes, err := dataFS.ReadFile("data/enchantments.json")
	if err != nil {
		return fmt.Errorf("failed to read enchantments.json: %w", err)
	}
	if err := json.Unmarshal(enchantmentBytes, &Enchantments); err != nil {
		return fmt.Errorf("failed to parse enchantments.json: %w", err)
	}

	// 4. Items
	itemBytes, err := dataFS.ReadFile("data/items.json")
	if err != nil {
		return fmt.Errorf("failed to read items.json: %w", err)
	}
	if err := json.Unmarshal(itemBytes, &Items); err != nil {
		return fmt.Errorf("failed to parse items.json: %w", err)
	}

	// 5. Sounds
	soundBytes, err := dataFS.ReadFile("data/sounds.json")
	if err != nil {
		return fmt.Errorf("failed to read sounds.json: %w", err)
	}
	if err := json.Unmarshal(soundBytes, &Sounds); err != nil {
		return fmt.Errorf("failed to parse sounds.json: %w", err)
	}

	// 6. Entities
	entityBytes, err := dataFS.ReadFile("data/entities.json")
	if err != nil {
		return fmt.Errorf("failed to read entities.json: %w", err)
	}
	if err := json.Unmarshal(entityBytes, &Entities); err != nil {
		return fmt.Errorf("failed to parse entities.json: %w", err)
	}

	// 7. Commands
	commandBytes, err := dataFS.ReadFile("data/commands.json")
	if err != nil {
		return fmt.Errorf("failed to read commands.json: %w", err)
	}
	if err := json.Unmarshal(commandBytes, &CommandTree); err != nil {
		return fmt.Errorf("failed to parse commands.json: %w", err)
	}

	return nil
}
