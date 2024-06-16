package parser

import (
	"fmt"
	"os"
)

type Action string

func ConvertAction(actionArgument string) (Action, error) {
	if actionArgument == string(ShowAction) {
		return ShowAction, nil
	} else if actionArgument == string(ClearAction) {
		return ClearAction, nil
	}
	return "", fmt.Errorf("invalid action: %s", actionArgument)
}

const (
	ShowAction  Action = "show"
	ClearAction Action = "clear"
)

type Parser struct {
	Name    string
	Support func(file *os.File) (bool, error)
	Handle  func(file *os.File, action Action) error
}
