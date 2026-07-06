package extension

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type updateState struct {
	CheckedAt  time.Time
	LatestHead string
}

func readState(path string) (updateState, error) {
	file, err := os.Open(path)
	if err != nil {
		return updateState{}, err
	}
	defer file.Close()
	var state updateState
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		switch key {
		case "checked_for_update_at":
			parsed, err := time.Parse(time.RFC3339, value)
			if err == nil {
				state.CheckedAt = parsed
			}
		case "latest_head":
			state.LatestHead = value
		}
	}
	return state, scanner.Err()
}

func writeState(path string, state updateState) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body := fmt.Sprintf("checked_for_update_at: %s\nlatest_head: %s\n", state.CheckedAt.Format(time.RFC3339), state.LatestHead)
	return os.WriteFile(path, []byte(body), 0o600)
}
