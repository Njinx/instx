package proxy

import (
	"encoding/json"
	"errors"
	"fmt"

	"gitlab.com/Njinx/instx/updater"
)

func callCommand(key string) (string, error) {
	var commands = map[string]func() (string, error){
		"update": cmdUpdate,
		"stats":  cmdStats,
	}

	fn, ok := commands[key]
	if !ok {
		return "", &ErrInvalidCommand{"update"}
	} else {
		return fn()
	}
}

func cmdUpdate() (string, error) {
	if err := updater.ForceUpdate(); errors.Is(err, &updater.ErrUpdateInProgress{}) {
		return "Update already in progress", nil
	} else {
		return "Updating list of instances. This may take a while.", nil
	}
}

func cmdStats() (string, error) {
	updatedCanidatesMutex.Lock()

	canidates := updater.NewCanidatesMarshalable(updatedCanidates)
	json, err := json.Marshal(canidates)
	if err != nil {
		return "", err
	}

	updatedCanidatesMutex.Unlock()

	return string(json), nil
}

type ErrInvalidCommand struct {
	Name string
}

func (err *ErrInvalidCommand) Error() string {
	return fmt.Sprintf("Invalid command: \"%s\"", err.Name)
}
