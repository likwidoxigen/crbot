package learn

import (
	"errors"
	"fmt"

	"github.com/jakevoytko/crbot/api"
	"github.com/jakevoytko/crbot/log"
	"github.com/jakevoytko/crbot/model"
	stringmap "github.com/jakevoytko/go-stringmap"
)

// CustomLearnExecutor learns a user-generated command.
type CustomLearnExecutor struct {
	commandMap stringmap.StringMap
}

// NewCustomLearnExecutor works as advertised.
func NewCustomLearnExecutor(commandMap stringmap.StringMap) *CustomLearnExecutor {
	return &CustomLearnExecutor{commandMap: commandMap}
}

// GetType returns the type of this feature.
func (f *CustomLearnExecutor) GetType() int {
	return model.CommandTypeLearn
}

// PublicOnly returns whether the executor should be intercepted in a private channel.
func (f *CustomLearnExecutor) PublicOnly() bool {
	return false
}

// Execute replies over the given channel with a help message.
func (f *CustomLearnExecutor) Execute(s api.DiscordSession, channel model.Snowflake, command *model.Command) {
	if command.Learn == nil {
		log.Fatal("Incorrectly generated learn command", errors.New("wat"))
	}
	if !command.Learn.CallOpen {
		s.ChannelMessageSend(channel.Format(), fmt.Sprintf(MsgLearnFail, command.Learn.Call))
		return
	}

	// Teach the command.
	if has, err := f.commandMap.Has(command.Learn.Call); err != nil || has {
		if has {
			log.Fatal("Collision when adding a call for "+command.Learn.Call, errors.New("wat"))
		}
		log.Fatal("Error in LearnFeature#Execute, testing a command", err)
	}
	if err := f.commandMap.Set(command.Learn.Call, command.Learn.Response); err != nil {
		log.Fatal("Error storing a learn command. Dying since it might work with restart", err)
	}

	// Send ack.
	s.ChannelMessageSend(channel.Format(), fmt.Sprintf(MsgLearnSuccess, command.Learn.Call))
}
