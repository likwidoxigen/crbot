package vote

import (
	"errors"
	"log"

	"github.com/jakevoytko/crbot/model"
)

type BallotParser struct {
	// The message that the parser looks for.
	Message string
	// Whether this message is in favor or against the measure.
	InFavor bool
}

// NewBallotParser works as advertised.
func NewBallotParser(message string, inFavor bool) *BallotParser {
	return &BallotParser{
		Message: message,
		InFavor: inFavor,
	}
}

// GetName returns the named type.
func (p *BallotParser) GetName() string {
	return p.Message
}

const (
	MsgHelpBallotInFavor = "Casts a ballot in favor of the current vote, if one is active"
	MsgHelpBallotAgainst = "Casts a ballot against the current vote, if one is active"
)

// GetHelpText returns the help text.
func (p *BallotParser) HelpText(command string) (string, error) {
	if p.InFavor {
		return MsgHelpBallotInFavor, nil
	}
	return MsgHelpBallotAgainst, nil
}

// Parse parses the given list command.
func (p *BallotParser) Parse(splitContent []string) (*model.Command, error) {
	if splitContent[0] != p.GetName() {
		log.Fatal("parseVoteBallot called with non-list command", errors.New("wat"))
	}
	return &model.Command{
		Type: model.Type_VoteBallot,
		Ballot: &model.BallotData{
			InFavor: p.InFavor,
		},
	}, nil
}
