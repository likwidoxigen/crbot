package moderation

import (
	"fmt"
	"strings"

	"github.com/jakevoytko/crbot/api"
	"github.com/jakevoytko/crbot/config"
	"github.com/jakevoytko/crbot/log"
	"github.com/jakevoytko/crbot/model"
)

// RickListInfoExecutor prints a rick roll.
type RickListInfoExecutor struct {
	config *config.Config
}

// NewRickListInfoExecutor works as advertised.
func NewRickListInfoExecutor(config *config.Config) *RickListInfoExecutor {
	return &RickListInfoExecutor{
		config: config,
	}
}

// GetType returns the type.
func (e *RickListInfoExecutor) GetType() int {
	return model.Type_RickListInfo
}

// PublicOnly returns whether the executor should be intercepted in a private channel.
func (e *RickListInfoExecutor) PublicOnly() bool {
	return false
}

const (
	MsgRickListEmpty = "Nobody is on the ricklist."
	MsgRickListUsers = "On the Rick list: "
)

// Execute replies over the given channel with a rick roll.
func (e *RickListInfoExecutor) Execute(s api.DiscordSession, channel model.Snowflake, command *model.Command) {
	if len(e.config.RickList) == 0 {
		if _, err := s.ChannelMessageSend(channel.Format(), MsgRickListEmpty); err != nil {
			log.Info("Failed to send ricklist message", err)
		}
		return
	}

	users := make([]string, 0, len(e.config.RickList))
	for _, ricklisted := range e.config.RickList {
		ricklistedFormat := ricklisted.Format()
		user, err := s.User(ricklistedFormat)
		if err != nil {
			log.Info(fmt.Sprintf("Unable to get info for user %s", ricklisted), err)
			users = append(users, ricklistedFormat)
			continue
		}
		users = append(users, "@"+user.Username)
	}

	finalString := MsgRickListUsers + "[" + strings.Join(users, ", ") + "]"

	if _, err := s.ChannelMessageSend(channel.Format(), finalString); err != nil {
		log.Info("Failed to send ricklist message", err)
	}
}
