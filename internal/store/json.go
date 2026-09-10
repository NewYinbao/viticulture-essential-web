// Package store persists versioned JSON snapshots.
package store

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const Filename = "ee-state-v1.json"

func Load(dir string, snapshot any) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := os.ReadFile(filepath.Join(dir, Filename))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, snapshot)
}

func Save(dir string, snapshot any) error {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	target := filepath.Join(dir, Filename)
	if err = os.WriteFile(target+".tmp", data, 0600); err != nil {
		return err
	}
	return os.Rename(target+".tmp", target)
}
