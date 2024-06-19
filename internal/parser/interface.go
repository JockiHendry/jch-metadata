package parser

import (
	"fmt"
	"jch-metadata/internal/output"
	"os"
)

type Action string

func ConvertAction(actionArgument string) (Action, error) {
	if actionArgument == string(ShowAction) {
		return ShowAction, nil
	} else if actionArgument == string(ClearAction) {
		return ClearAction, nil
	} else if actionArgument == string(ExtractAction) {
		return ExtractAction, nil
	}
	return "", fmt.Errorf("invalid action: %s", actionArgument)
}

const (
	ShowAction    Action = "show"
	ClearAction   Action = "clear"
	ExtractAction Action = "extract"
)

type Parser struct {
	Name      string
	Container bool
	Support   func(file *os.File, startOffset int64) (bool, error)
	Handle    func(file *os.File, action Action, startOffset int64, parsers []Parser) error
}

func StartParsing(parsers []Parser, file *os.File, action Action, startOffset int64) (bool, error) {
	parsed := false
	for _, p := range parsers {
		if startOffset > 0 && p.Container {
			continue
		}
		supported, err := p.Support(file, startOffset)
		if err != nil {
			return false, err
		}
		if !supported {
			continue
		}
		output.Printf(startOffset > 0, "File type is %s\n\n", p.Name)
		err = p.Handle(file, action, startOffset, parsers)
		if err != nil {
			output.Printf(startOffset > 0, "Error handling file: %s", err)
			return false, err
		}
		parsed = true
		break
	}
	return parsed, nil
}
