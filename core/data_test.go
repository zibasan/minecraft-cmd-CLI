package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDataExternalOverride(t *testing.T) {
	// Create temporary sandbox directory
	tempDir, err := os.MkdirTemp("", "cmdforge_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save original working directory
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(origWd)
	}()

	// Create data directory in sandbox
	testDataDir := filepath.Join(tempDir, "data")
	err = os.MkdirAll(testDataDir, 0755)
	if err != nil {
		t.Fatalf("failed to create test data dir: %v", err)
	}

	// Write custom blocks.json and components.json
	testBlocks := []string{"custom:block_a", "custom:block_b"}
	blockFile := filepath.Join(testDataDir, "blocks.json")
	fileBytes, err := json.Marshal(testBlocks)
	if err != nil {
		t.Fatalf("failed to marshal test blocks: %v", err)
	}
	err = os.WriteFile(blockFile, fileBytes, 0644)
	if err != nil {
		t.Fatalf("failed to write test blocks file: %v", err)
	}

	testComponents := []string{"custom:comp_a"}
	compFile := filepath.Join(testDataDir, "components.json")
	compBytes, err := json.Marshal(testComponents)
	if err != nil {
		t.Fatalf("failed to marshal test components: %v", err)
	}
	err = os.WriteFile(compFile, compBytes, 0644)
	if err != nil {
		t.Fatalf("failed to write test components file: %v", err)
	}

	// Switch working directory to sandbox
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change working directory to sandbox: %v", err)
	}

	// Run LoadData to verify override
	err = LoadData()
	if err != nil {
		t.Fatalf("LoadData() returned error: %v", err)
	}

	if len(Blocks) != 2 || Blocks[0] != "custom:block_a" || Blocks[1] != "custom:block_b" {
		t.Errorf("expected Blocks to be overridden with %v, got %v", testBlocks, Blocks)
	}

	if len(Components) != 1 || Components[0] != "custom:comp_a" {
		t.Errorf("expected Components to be overridden with %v, got %v", testComponents, Components)
	}

	// Corrupt blocks.json to test fallback
	err = os.WriteFile(blockFile, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("failed to corrupt test blocks file: %v", err)
	}

	// Run LoadData again, should fallback without failing
	err = LoadData()
	if err != nil {
		t.Fatalf("LoadData() should succeed even with corrupted external JSON, but got: %v", err)
	}

	// Check that we got embedded data back (which shouldn't contain custom:block_a)
	foundCustom := false
	for _, b := range Blocks {
		if b == "custom:block_a" {
			foundCustom = true
			break
		}
	}
	if foundCustom {
		t.Error("expected Blocks to fallback to embedded data, but it still has overridden custom values")
	}
}
