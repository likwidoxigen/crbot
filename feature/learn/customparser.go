package learn

import (
	"errors"
	"strings"

	"github.com/jakevoytko/crbot/log"
	"github.com/jakevoytko/crbot/model"
)

// CustomParser parses all fallthrough commands.
type CustomParser struct {
	commandMap model.StringMap
}

// NewCustomParser works as advertised.
func NewCustomParser(commandMap model.StringMap) *CustomParser {
	return &CustomParser{
		commandMap: commandMap,
	}
}

// GetName returns nothing, since it doesn't have a user-invokable name.
func (p *CustomParser) GetName() string {
	return ""
}

// HelpText returns help text for the given custom command.
func (p *CustomParser) HelpText(command string) (string, error) {
	ok, err := p.commandMap.Has(command)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", nil
	}

	val, err := p.commandMap.Get(command)
	if err != nil {
		return "", err
	}
	response := "?" + command
	if strings.Contains(val, "$1") {
		response = response + " <args>"
	}
	return response, nil
}

// Parse parses the given custom command.
func (f *CustomParser) Parse(splitContent []string) (*model.Command, error) {
	// TODO(jake): Drop this and external hash check, handle missing commands solely in execute.
	has, err := f.commandMap.Has(splitContent[0][1:])
	if err != nil {
		return nil, err
	}
	if !has {
		log.Fatal("parseCustom called with non-custom command", errors.New("wat"))
	}
	return &model.Command{
		Type: model.Type_Custom,
		Custom: &model.CustomData{
			Call: splitContent[0][1:],
			Args: strings.Join(splitContent[1:], " "),
		},
	}, nil
}