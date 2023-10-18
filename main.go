package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type plan struct {
	ResourceChanges []json.RawMessage `json:"resource_changes"`
}

type resourceChange struct {
	Address string `json:"address"`
	Change  change `json:"change"`
}

type change struct {
	Before json.RawMessage `json:"before"`
	After  json.RawMessage `json:"after"`
}

func main() {
	dir := flag.String("dir", "diff", "Directory to write the output")
	flag.Parse()
	if err := process(*dir); err != nil {
		panic(err)
	}
}

func process(dir string) error {
	var p plan
	m := json.NewDecoder(os.Stdin)
	if err := m.Decode(&p); err != nil {
		return fmt.Errorf("failed to read input: %v", err)
	}
	before := filepath.Join(dir, "before")
	after := filepath.Join(dir, "after")
	if err := os.RemoveAll(before); err != nil {
		return fmt.Errorf("failed to clean %q: %w", before, err)
	}
	if err := os.RemoveAll(after); err != nil {
		return fmt.Errorf("failed to clean %q: %w", after, err)
	}
	if err := os.MkdirAll(before, 0777); err != nil {
		return fmt.Errorf("failed to create %q: %w", before, err)
	}
	if err := os.MkdirAll(after, 0777); err != nil {
		return fmt.Errorf("failed to create %q: %w", after, err)
	}

	for i, raw := range p.ResourceChanges {
		var rc resourceChange
		if err := json.Unmarshal(raw, &rc); err != nil {
			return fmt.Errorf("failed to unmarshal resource change %d: %w", i, err)
		}
		if err := write(filepath.Join(before, fmt.Sprintf("%s.json", rc.Address)), rc.Change.Before); err != nil {
			return err
		}
		if err := write(filepath.Join(after, fmt.Sprintf("%s.json", rc.Address)), rc.Change.After); err != nil {
			return err
		}
	}

	return nil
}

func write(path string, resource json.RawMessage) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create %q: %w", path, err)
	}
	defer f.Close()

	e := json.NewEncoder(f)
	e.SetIndent("", " ")

	if err := e.Encode(resource); err != nil {
		return fmt.Errorf("failed to write to %q: %w", path, err)
	}
	return nil
}
