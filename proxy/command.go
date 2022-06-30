package proxy

import (
	"errors"
	"fmt"

	"gitlab.com/Njinx/instx/updater"
)

func callCommand(key string) (string, error) {
	var commands = map[string]func() (string, error){
		"update": cmdUpdate,
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

type ErrInvalidCommand struct {
	cmdName string
}

func (err *ErrInvalidCommand) Error() string {
	return fmt.Sprintf("Invalid command: \"%s\"", err.cmdName)
}
