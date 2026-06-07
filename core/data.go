package core

import (
	"bytes"
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
	Components   []string
	CommandTree  BrigadierNode
)

// loadJSONData checks if the JSON file exists locally in the current directory's data/ folder.
// If it exists, it attempts to load and parse it to override embedded data.
// If it fails or does not exist, it falls back to the embedded asset.
func loadJSONData(fileName string, v interface{}, embedPath string) error {
	filePath := "data/" + fileName

	if _, err := os.Stat(filePath); err == nil {
		fileBytes, err := os.ReadFile(filePath)
		if err == nil {
			// Strip UTF-8 BOM if present
			fileBytes = bytes.TrimPrefix(fileBytes, []byte("\xef\xbb\xbf"))
			if err := json.Unmarshal(fileBytes, v); err == nil {
				return nil
			} else {
				fmt.Fprintf(os.Stderr, "Warning: failed to parse external file %s: %v. Falling back to embedded data.\n", filePath, err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Warning: failed to read external file %s: %v. Falling back to embedded data.\n", filePath, err)
		}
	}

	embedBytes, err := dataFS.ReadFile(embedPath)
	if err != nil {
		return fmt.Errorf("failed to read embedded data %s: %w", embedPath, err)
	}
	if err := json.Unmarshal(embedBytes, v); err != nil {
		return fmt.Errorf("failed to parse embedded data %s: %w", embedPath, err)
	}

	return nil
}

// LoadData reads and decodes the JSON files into memory, prioritizing local files.
func LoadData() error {
	// 1. Blocks
	if err := loadJSONData("blocks.json", &Blocks, "data/blocks.json"); err != nil {
		return err
	}

	// 2. Effects
	if err := loadJSONData("effects.json", &Effects, "data/effects.json"); err != nil {
		return err
	}

	// 3. Enchantments
	if err := loadJSONData("enchantments.json", &Enchantments, "data/enchantments.json"); err != nil {
		return err
	}

	// 4. Items
	if err := loadJSONData("items.json", &Items, "data/items.json"); err != nil {
		return err
	}

	// 5. Sounds
	if err := loadJSONData("sounds.json", &Sounds, "data/sounds.json"); err != nil {
		return err
	}

	// 6. Entities
	if err := loadJSONData("entities.json", &Entities, "data/entities.json"); err != nil {
		return err
	}

	// 7. Components
	if err := loadJSONData("components.json", &Components, "data/components.json"); err != nil {
		return err
	}

	// 8. Commands
	if err := loadJSONData("commands.json", &CommandTree, "data/commands.json"); err != nil {
		return err
	}

	return nil
}
