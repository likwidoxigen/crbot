package learn

import (
	"errors"
	"fmt"

	"github.com/jakevoytko/crbot/api"
	"github.com/jakevoytko/crbot/log"
	"github.com/jakevoytko/crbot/model"
	stringmap "github.com/jakevoytko/go-stringmap"
)

// UnlearnExecutor attempts to unlearn a custom command and returns the result to the user.
type UnlearnExecutor struct {
	commandMap stringmap.StringMap
}

// NewUnlearnExecutor works as advertised.
func NewUnlearnExecutor(commandMap stringmap.StringMap) *UnlearnExecutor {
	return &UnlearnExecutor{commandMap: commandMap}
}

// GetType returns the type of this feature.
func (e *UnlearnExecutor) GetType() int {
	return model.CommandTypeUnlearn
}

// PublicOnly returns whether the executor should be intercepted in a private channel.
func (e *UnlearnExecutor) PublicOnly() bool {
	return true
}

// Execute replies over the given channel indicating successful unlearning, or
// failure to unlearn.
func (e *UnlearnExecutor) Execute(s api.DiscordSession, channel model.Snowflake, command *model.Command) {
	if command.Unlearn == nil {
		log.Fatal("Incorrectly generated unlearn command", errors.New("wat"))
	}

	if !command.Unlearn.CallOpen {
		s.ChannelMessageSend(channel.Format(), fmt.Sprintf(MsgUnlearnFail, command.Unlearn.Call))
		return
	}

	// Remove the command.
	if has, err := e.commandMap.Has(command.Unlearn.Call); !has || err != nil {
		if has {
			log.Fatal("Tried to unlearn command that doesn't exist: "+command.Unlearn.Call, errors.New("wat"))
		}
		log.Fatal("Error in UnlearnFeature#execute, testing a command", err)
	}
	if err := e.commandMap.Delete(command.Unlearn.Call); err != nil {
		log.Fatal("Unsuccessful unlearning a key; Dying since it might work with a restart", err)
	}

	// Send ack.
	s.ChannelMessageSend(channel.Format(), fmt.Sprintf(MsgUnlearnSuccess, command.Unlearn.Call))
}
