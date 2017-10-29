package main

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jakevoytko/crbot/api"
	"github.com/jakevoytko/crbot/app"
	"github.com/jakevoytko/crbot/feature"
	"github.com/jakevoytko/crbot/feature/moderation"
	"github.com/jakevoytko/crbot/feature/vote"
	"github.com/jakevoytko/crbot/model"
	"github.com/jakevoytko/crbot/util"
)

const (
	MainChannelID   = model.Snowflake(8675309)
	SecondChannelID = model.Snowflake(9000000)
	DirectMessageID = model.Snowflake(1)
)

func TestNewServer(t *testing.T) {
	runner := NewTestRunner(t)

	// Assert initial state.
	runner.AssertState()
}

func TestLearn_NoResponse(t *testing.T) {
	runner := NewTestRunner(t)

	// Commands that should never return a response.
	runner.SendMessageWithoutResponse(MainChannelID, "?")
	runner.SendMessageWithoutResponse(MainChannelID, "!")
	runner.SendMessageWithoutResponse(MainChannelID, ".")
	runner.SendMessageWithoutResponse(MainChannelID, "")
	runner.SendMessageWithoutResponse(MainChannelID, "!help")
	runner.SendMessageWithoutResponse(MainChannelID, "help")
	runner.SendMessageWithoutResponse(MainChannelID, ".help")

	// Test ?list. ?list tests will be interspersed through the learn examples
	// below, since learn and unlearn interact with it.
	runner.SendListMessage(MainChannelID)
}

func TestLearn_WrongFormat(t *testing.T) {
	runner := NewTestRunner(t)

	// Basic learn responses.
	// Wrong call format
	runner.SendMessage(MainChannelID, "?learn", MsgHelpLearn)
	runner.SendMessage(MainChannelID, "?learn test", MsgHelpLearn)
	runner.SendMessage(MainChannelID, "?learn ?call response", MsgHelpLearn)
	runner.SendMessage(MainChannelID, "?learn !call response", MsgHelpLearn)
	runner.SendMessage(MainChannelID, "?learn /call response", MsgHelpLearn)
	runner.SendMessage(MainChannelID, "?learn ", MsgHelpLearn)
	runner.SendMessage(MainChannelID, "?learn multi\nline\ncall response", MsgHelpLearn)
	// Wrong response format.
	runner.SendMessage(MainChannelID, "?learn call ?response", MsgHelpLearn)
	runner.SendMessage(MainChannelID, "?learn call !response", MsgHelpLearn)
}

func TestIntegration(t *testing.T) {
	runner := NewTestRunner(t)
	// Test ?list. ?list tests will be interspersed through the learn examples
	// below, since learn and unlearn interact with it.
	runner.SendListMessage(MainChannelID)

	// Valid learns.
	runner.SendLearnMessage(MainChannelID, "?learn call response", NewLearn("call", "response"))
	runner.SendLearnMessage(MainChannelID, "?learn call2 multi word response", NewLearn("call2", "multi word response"))
	runner.SendLearnMessage(MainChannelID, "?learn call3 multi\nline\nresponse\n", NewLearn("call3", "multi\nline\nresponse\n"))
	runner.SendLearnMessage(MainChannelID, "?learn call4 \\/leave", NewLearn("call4", "\\/leave"))
	runner.SendLearnMessage(MainChannelID, "?learn bearshrug ʅʕ•ᴥ•ʔʃ", NewLearn("bearshrug", "ʅʕ•ᴥ•ʔʃ"))
	runner.SendLearnMessage(MainChannelID, "?learn emoji ⛄⛄⛄⛄", NewLearn("emoji", "⛄⛄⛄⛄")) // Emoji is "snowman without snow", in case this isn't showing up in your editor.
	runner.SendLearnMessage(MainChannelID, "?learn args1 hello $1", NewLearn("args1", "hello $1"))
	runner.SendLearnMessage(MainChannelID, "?learn args2 $1", NewLearn("args2", "$1"))
	runner.SendLearnMessage(MainChannelID, "?learn args3 $1 $1", NewLearn("args3", "$1 $1"))
	runner.SendLearnMessage(MainChannelID, "?learn args4 $1 $1 $1 $1 $1", NewLearn("args4", "$1 $1 $1 $1 $1"))
	// Cannot overwrite a learn.
	runner.SendMessage(MainChannelID, "?learn call response", fmt.Sprintf(MsgLearnFail, "call"))
	// List should now include learns.
	runner.SendListMessage(MainChannelID)
	// Extra whitespace test.
	runner.SendLearnMessage(MainChannelID, "?learn  spaceBeforeCall response", NewLearn("spaceBeforeCall", "response"))
	runner.SendLearnMessage(MainChannelID, "?learn spaceBeforeResponse  response", NewLearn("spaceBeforeResponse", "response"))
	runner.SendLearnMessage(MainChannelID, "?learn spaceInResponse response  two  spaces", NewLearn("spaceInResponse", "response  two  spaces"))

	// Test learned commands.
	runner.SendMessage(MainChannelID, "?call", "response")
	runner.SendMessage(MainChannelID, "?call2", "multi word response")
	runner.SendMessage(MainChannelID, "?call3", "multi\nline\nresponse\n")
	runner.SendMessage(MainChannelID, "?call4", "\\/leave")
	runner.SendMessage(MainChannelID, "?bearshrug", "ʅʕ•ᴥ•ʔʃ")
	runner.SendMessage(MainChannelID, "?emoji", "⛄⛄⛄⛄")
	runner.SendMessage(MainChannelID, "?args1 world", "hello world")
	runner.SendMessage(MainChannelID, "?args2 world", "world")
	runner.SendMessage(MainChannelID, "?args3 world", "world world")
	runner.SendMessage(MainChannelID, "?args3     leadingspaces", "    leadingspaces     leadingspaces")
	runner.SendMessage(MainChannelID, "?args4 world", "world world world world $1")
	runner.SendMessage(MainChannelID, "?args4     leadingspaces", "    leadingspaces     leadingspaces     leadingspaces     leadingspaces $1")

	runner.SendMessage(MainChannelID, "?args1", MsgCustomNeedsArgs)
	runner.SendMessage(MainChannelID, "?spaceBeforeCall", "response")
	runner.SendMessage(MainChannelID, "?spaceBeforeResponse", "response")
	runner.SendMessage(MainChannelID, "?spaceInResponse", "response  two  spaces")
	// Fallback commands aren't triggered unless they lead a message.
	runner.SendMessageWithoutResponse(MainChannelID, " ?call")
	runner.SendMessageWithoutResponse(MainChannelID, "i just met you, and this is lazy, but here's my number, ?call me maybe")
	runner.SendMessageWithoutResponse(MainChannelID, "\n?call")
	// List should still have the messages.
	runner.SendListMessage(MainChannelID)

	// Test unlearn.
	// Wrong format.
	runner.SendMessage(MainChannelID, "?unlearn", MsgHelpUnlearn)
	runner.SendMessage(MainChannelID, "?unlearn ", MsgHelpUnlearn)
	// Can't unlearn in a private channel
	runner.SendMessage(DirectMessageID, "?unlearn call", MsgUnlearnMustBePublic)
	// Can't unlearn builtin commands.
	runner.SendMessage(MainChannelID, "?unlearn help", fmt.Sprintf(MsgUnlearnFail, "help"))
	runner.SendMessage(MainChannelID, "?unlearn learn", fmt.Sprintf(MsgUnlearnFail, "learn"))
	runner.SendMessage(MainChannelID, "?unlearn list", fmt.Sprintf(MsgUnlearnFail, "list"))
	runner.SendMessage(MainChannelID, "?unlearn unlearn", fmt.Sprintf(MsgUnlearnFail, "unlearn"))
	runner.SendMessage(MainChannelID, "?unlearn ?help", MsgHelpUnlearn)
	runner.SendMessage(MainChannelID, "?unlearn ?learn", MsgHelpUnlearn)
	runner.SendMessage(MainChannelID, "?unlearn ?list", MsgHelpUnlearn)
	runner.SendMessage(MainChannelID, "?unlearn ?unlearn", MsgHelpUnlearn)
	// Unrecognized command.
	runner.SendMessage(MainChannelID, "?unlearn  bears", fmt.Sprintf(MsgUnlearnFail, "bears"))
	runner.SendMessage(MainChannelID, "?unlearn somethingIdon'tknow", fmt.Sprintf(MsgUnlearnFail, "somethingIdon'tknow"))
	// Valid unlearn.
	runner.SendUnlearnMessage(MainChannelID, "?unlearn call", "call")
	runner.SendMessageWithoutResponse(MainChannelID, "?call")
	// List should work after the unlearn.
	runner.SendListMessage(MainChannelID)
	// Can then relearn.
	runner.SendLearnMessage(MainChannelID, "?learn call another response", NewLearn("call", "another response"))
	runner.SendMessage(MainChannelID, "?call", "another response")
	// List should work after the relearn.
	runner.SendListMessage(MainChannelID)
	// Unlearn with 2 spaces.
	runner.SendUnlearnMessage(MainChannelID, "?unlearn  call", "call")
	runner.SendMessageWithoutResponse(MainChannelID, "?call")

	// Unrecognized help commands.
	runner.SendMessage(MainChannelID, "?help", MsgDefaultHelp)
	runner.SendMessage(MainChannelID, "?help abunchofgibberish", MsgDefaultHelp)
	runner.SendMessage(MainChannelID, "?help ??help", MsgDefaultHelp)
	// All recognized help commands.
	runner.SendMessage(MainChannelID, "?help help", MsgHelpHelp)
	runner.SendMessage(MainChannelID, "?help learn", MsgHelpLearn)
	runner.SendMessage(MainChannelID, "?help list", MsgHelpList)
	runner.SendMessage(MainChannelID, "?help unlearn", MsgHelpUnlearn)
	runner.SendMessage(MainChannelID, "?help ?help", MsgHelpHelp)
	runner.SendMessage(MainChannelID, "?help ?learn", MsgHelpLearn)
	runner.SendMessage(MainChannelID, "?help ?list", MsgHelpList)
	runner.SendMessage(MainChannelID, "?help ?unlearn", MsgHelpUnlearn)
	runner.SendMessage(MainChannelID, "?help  help", MsgHelpHelp)
	// Help with custom commands.
	runner.SendLearnMessage(MainChannelID, "?learn help-noarg response", NewLearn("help-noarg", "response"))
	runner.SendLearnMessage(MainChannelID, "?learn help-arg response $1", NewLearn("help-arg", "response $1"))
	runner.SendMessage(MainChannelID, "?help help-noarg", "?help-noarg")
	runner.SendMessage(MainChannelID, "?help help-arg", "?help-arg <args>")
	runner.SendUnlearnMessage(MainChannelID, "?unlearn help-noarg", "help-noarg")
	runner.SendUnlearnMessage(MainChannelID, "?unlearn help-arg", "help-arg")
	runner.SendMessage(MainChannelID, "?help help-noarg", MsgDefaultHelp)
	runner.SendMessage(MainChannelID, "?help help-arg", MsgDefaultHelp)

	// Moderation
	rickListedUser := &discordgo.User{
		ID:            "2",
		Email:         "email@example.com",
		Username:      "username",
		Avatar:        "avatar",
		Discriminator: "discriminator",
		Token:         "token",
		Verified:      true,
		MFAEnabled:    false,
		Bot:           false,
	}
	runner.SendMessageAs(rickListedUser, MainChannelID, "?help help-arg", MsgDefaultHelp)
	runner.SendMessageAs(rickListedUser, DirectMessageID, "?help help-arg", moderation.MsgRickList)
	runner.SendLearnMessageAs(rickListedUser, DirectMessageID, "?learn rick list", NewLearn("rick", "list"))
}

func TestVote(t *testing.T) {
	runner := NewTestRunner(t)
	runner.SendVoteStatusMessage(MainChannelID)

	// Calls vote with no args, and then actually starts a vote.
	author := newUser("author", 0 /* id */, false /* bot */)
	runner.AddUser(author)
	runner.SendMessageAs(author, MainChannelID, "?vote", vote.MsgHelpVote)
	runner.SendVoteMessageAs(author, MainChannelID)
	runner.SendVoteStatusMessage(MainChannelID)

	// Assert that a second vote can't be started.
	runner.SendMessageAs(author, MainChannelID, "?vote another vote", vote.MsgActiveVote)

	// Time the vote out.
	runner.ExpireVote(MainChannelID)
	runner.SendVoteStatusMessage(MainChannelID)

	// A second vote can be started once it is expired.
	runner.SendVoteMessageAs(author, MainChannelID)
	runner.SendVoteStatusMessage(MainChannelID)
}

func TestVote_Pass(t *testing.T) {
	runner := NewTestRunner(t)

	// Initialize users.
	users := []*discordgo.User{
		newUser("user0", 0 /* id */, false /* bot */),
		newUser("user1", 1 /* id */, false /* bot */),
		newUser("user2", 2 /* id */, false /* bot */),
		newUser("user3", 3 /* id */, false /* bot */),
		newUser("user4", 4 /* id */, false /* bot */),
	}
	for _, user := range users {
		runner.AddUser(user)
	}

	// Start the vote.
	runner.SendVoteMessageAs(users[0], MainChannelID)
	runner.SendVoteStatusMessage(MainChannelID)

	// Cast votes
	for _, user := range users {
		runner.CastBallotAs(user, MainChannelID, true /* inFavor */)
		runner.SendVoteStatusMessage(MainChannelID)
	}

	runner.ExpireVote(MainChannelID)
	runner.SendVoteStatusMessage(MainChannelID)
}

func TestVote_Fail(t *testing.T) {
	runner := NewTestRunner(t)

	// Initialize users.
	users := []*discordgo.User{
		newUser("user0", 0 /* id */, false /* bot */),
		newUser("user1", 1 /* id */, false /* bot */),
		newUser("user2", 2 /* id */, false /* bot */),
		newUser("user3", 3 /* id */, false /* bot */),
		newUser("user4", 4 /* id */, false /* bot */),
	}
	for _, user := range users {
		runner.AddUser(user)
	}

	// Start the vote.
	runner.SendVoteMessageAs(users[0], MainChannelID)
	runner.SendVoteStatusMessage(MainChannelID)

	// Cast votes
	for _, user := range users {
		runner.CastBallotAs(user, MainChannelID, false /* inFavor */)
		runner.SendVoteStatusMessage(MainChannelID)
	}

	runner.ExpireVote(MainChannelID)
	runner.SendVoteStatusMessage(MainChannelID)
}

func TestVote_Tie(t *testing.T) {
	runner := NewTestRunner(t)

	// Initialize users.
	users := []*discordgo.User{
		newUser("user0", 0 /* id */, false /* bot */),
		newUser("user1", 1 /* id */, false /* bot */),
		newUser("user2", 2 /* id */, false /* bot */),
		newUser("user3", 3 /* id */, false /* bot */),
		newUser("user4", 4 /* id */, false /* bot */),
		newUser("user5", 5 /* id */, false /* bot */),
		newUser("user6", 6 /* id */, false /* bot */),
		newUser("user7", 7 /* id */, false /* bot */),
		newUser("user8", 8 /* id */, false /* bot */),
		newUser("user9", 9 /* id */, false /* bot */),
	}
	for _, user := range users {
		runner.AddUser(user)
	}

	// Start the vote.
	runner.SendVoteMessageAs(users[0], MainChannelID)
	runner.SendVoteStatusMessage(MainChannelID)

	// Cast votes
	for _, user := range users {
		runner.CastBallotAs(user, MainChannelID, false /* inFavor */)
		runner.SendVoteStatusMessage(MainChannelID)
	}

	runner.ExpireVote(MainChannelID)
	runner.SendVoteStatusMessage(MainChannelID)
}

func TestVote_TwoVotes(t *testing.T) {
	runner := NewTestRunner(t)

	// Initialize users.
	users := []*discordgo.User{
		newUser("user0", 0 /* id */, false /* bot */),
		newUser("user1", 1 /* id */, false /* bot */),
		newUser("user2", 2 /* id */, false /* bot */),
		newUser("user3", 3 /* id */, false /* bot */),
		newUser("user4", 4 /* id */, false /* bot */),
	}
	for _, user := range users {
		runner.AddUser(user)
	}

	// Start the vote.
	runner.SendVoteMessageAs(users[0], MainChannelID)
	runner.SendVoteStatusMessage(MainChannelID)

	// Cast votes
	for _, user := range users {
		runner.CastBallotAs(user, MainChannelID, true /* inFavor */)
		runner.SendVoteStatusMessage(MainChannelID)
	}

	runner.ExpireVote(MainChannelID)
	runner.SendVoteStatusMessage(MainChannelID)

	// Start the vote again.
	runner.SendVoteMessageAs(users[0], MainChannelID)
	runner.SendVoteStatusMessage(MainChannelID)

	// Cast votes again.
	for _, user := range users {
		runner.CastBallotAs(user, MainChannelID, false /* inFavor */)
		runner.SendVoteStatusMessage(MainChannelID)
	}

	runner.ExpireVote(MainChannelID)
	runner.SendVoteStatusMessage(MainChannelID)
}

func TestVote_TwoChannels(t *testing.T) {
	runner := NewTestRunner(t)

	// Initialize users.
	users := []*discordgo.User{
		newUser("user0", 0 /* id */, false /* bot */),
		newUser("user1", 1 /* id */, false /* bot */),
		newUser("user2", 2 /* id */, false /* bot */),
		newUser("user3", 3 /* id */, false /* bot */),
		newUser("user4", 4 /* id */, false /* bot */),
	}
	for _, user := range users {
		runner.AddUser(user)
	}

	// Start the votes.
	runner.SendVoteMessageAs(users[0], MainChannelID)
	runner.SendVoteMessageAs(users[0], SecondChannelID)
	runner.SendVoteStatusMessage(MainChannelID)
	runner.SendVoteStatusMessage(SecondChannelID)

	// Cast votes
	for _, user := range users {
		runner.CastBallotAs(user, MainChannelID, true /* inFavor */)
		runner.CastBallotAs(user, SecondChannelID, false /* inFavor */)
		runner.SendVoteStatusMessage(MainChannelID)
		runner.SendVoteStatusMessage(SecondChannelID)
	}

	// Expires both votes.
	runner.ExpireVote(MainChannelID)
	runner.ExpireVote(SecondChannelID)
	runner.SendVoteStatusMessage(MainChannelID)
	runner.SendVoteStatusMessage(SecondChannelID)
}

// TestRunner is a helper that executes messages incrementally, and asserts that
// the global state is always what is expected.
type TestRunner struct {
	// Test object
	T *testing.T

	// State
	GistsCount           int
	DiscordMessagesCount int
	Learns               map[string]*Learn
	ActiveVoteMap        map[model.Snowflake]*Vote // channel->vote. May be nil

	// Fakes
	CustomMap      *util.InMemoryStringMap
	VoteMap        *util.InMemoryStringMap
	Gist           *util.InMemoryGist
	DiscordSession *util.InMemoryDiscordSession
	UTCClock       *util.FakeUTCClock

	// Real objects
	FeatureRegistry *feature.Registry

	// Controllers under test
	Handler func(api.DiscordSession, *discordgo.MessageCreate)
}

func NewTestRunner(t *testing.T) *TestRunner {
	// Initialize fakes.
	customMap := util.NewInMemoryStringMap()
	voteMap := util.NewInMemoryStringMap()
	gist := util.NewInMemoryGist()
	discordSession := util.NewInMemoryDiscordSession()
	discordSession.SetChannel(&discordgo.Channel{
		ID:   MainChannelID.Format(),
		Type: discordgo.ChannelTypeGuildText,
	})
	discordSession.SetChannel(&discordgo.Channel{
		ID:   SecondChannelID.Format(),
		Type: discordgo.ChannelTypeGuildText,
	})
	discordSession.SetChannel(&discordgo.Channel{
		ID:   DirectMessageID.Format(),
		Type: discordgo.ChannelTypeDM,
	})

	rickList := []model.Snowflake{model.Snowflake(2)}

	utcClock := util.NewFakeUTCClock()

	registry := InitializeRegistry(customMap, voteMap, gist, &app.Config{RickList: rickList}, utcClock)
	return &TestRunner{
		T:                    t,
		Learns:               map[string]*Learn{},
		ActiveVoteMap:        map[model.Snowflake]*Vote{},
		GistsCount:           0,
		DiscordMessagesCount: 0,
		CustomMap:            customMap,
		VoteMap:              voteMap,
		Gist:                 gist,
		DiscordSession:       discordSession,
		UTCClock:             utcClock,
		FeatureRegistry:      registry,
		Handler:              getHandleMessage(customMap, registry),
	}
}

func (r *TestRunner) AssertState() {
	r.T.Helper()

	// Assert counts.
	assertNumCommands(r.T, r.CustomMap, len(r.Learns))
	assertNumGists(r.T, r.Gist, r.GistsCount)
	assertNumDiscordMessages(r.T, r.DiscordSession, r.DiscordMessagesCount)
	assertVote(r.T, r.UTCClock, r.VoteMap, r.ActiveVoteMap)

	// Assert command map state.
	for _, learn := range r.Learns {
		assertCommand(r.T, r.CustomMap, learn.Call, learn.Response)
	}
}

func (r *TestRunner) SendMessage(channel model.Snowflake, message, expectedResponse string) {
	r.T.Helper()

	sendMessage(r.DiscordSession, r.Handler, channel, message)
	r.DiscordMessagesCount++
	assertNewMessages(r.T, r.DiscordSession,
		[]*util.Message{util.NewMessage(channel.Format(), expectedResponse)})
	r.AssertState()
}

func (r *TestRunner) SendMessageAs(author *discordgo.User, channel model.Snowflake, message, expectedResponse string) {
	r.T.Helper()

	sendMessageAs(author, r.DiscordSession, r.Handler, channel, message)
	r.DiscordMessagesCount++
	assertNewMessages(r.T, r.DiscordSession,
		[]*util.Message{util.NewMessage(channel.Format(), expectedResponse)})
	r.AssertState()
}

func (r *TestRunner) SendLearnMessage(channel model.Snowflake, message string, learn *Learn) {
	r.T.Helper()

	sendMessage(r.DiscordSession, r.Handler, channel, message)
	r.DiscordMessagesCount++
	r.Learns[learn.Call] = learn
	assertNewMessages(r.T, r.DiscordSession,
		[]*util.Message{util.NewMessage(channel.Format(), fmt.Sprintf(MsgLearnSuccess, learn.Call))})
	r.AssertState()
	r.SendListMessage(channel)
}

func (r *TestRunner) SendVoteMessageAs(author *discordgo.User, channel model.Snowflake) {
	r.T.Helper()

	sendMessageAs(author, r.DiscordSession, r.Handler, channel, "?vote a vote has been called")
	r.DiscordMessagesCount++
	r.ActiveVoteMap[channel] = newVote(channel, author, "a vote has been called", r.UTCClock.Now().Add(vote.VoteDuration))
	assertNewMessages(r.T, r.DiscordSession,
		[]*util.Message{util.NewMessage(channel.Format(), fmt.Sprintf(vote.MsgBroadcastNewVote, author.Mention(), "a vote has been called"))})
	r.AssertState()
}

func (r *TestRunner) CastBallotAs(author *discordgo.User, channel model.Snowflake, inFavor bool) {
	r.T.Helper()

	voteString := "?no"
	expectedMessage := fmt.Sprintf(vote.MsgVotedAgainst, author.Mention())
	toAppend := &(r.ActiveVoteMap[channel].VotesAgainst)
	if inFavor {
		voteString = "?yes"
		expectedMessage = fmt.Sprintf(vote.MsgVotedInFavor, author.Mention())
		toAppend = &(r.ActiveVoteMap[channel].VotesFor)
	}

	sendMessageAs(author, r.DiscordSession, r.Handler, channel, voteString)

	// Update internal state.
	r.DiscordMessagesCount++
	id, _ := model.ParseSnowflake(author.ID)
	*toAppend = append(*toAppend, id)

	// Reconstruct the status string and assert internal state.
	activeVote := r.ActiveVoteMap[channel]
	reconstructedVote := vote.NewVote(
		0, /* voteID */
		channel,
		id,
		activeVote.Message,
		time.Time{},
		activeVote.TimestampEnd,
		activeVote.VotesFor,
		activeVote.VotesAgainst,
		vote.VoteOutcomeNotDone)

	assertNewMessages(r.T, r.DiscordSession, []*util.Message{
		util.NewMessage(channel.Format(), expectedMessage+"\n"+vote.StatusLine(r.UTCClock, reconstructedVote)),
	})
	r.AssertState()
}

// Advances the clock enough that the vote expires.
func (r *TestRunner) ExpireVote(channel model.Snowflake) {
	r.UTCClock.Advance(vote.VoteDuration)
	r.ActiveVoteMap[channel] = nil
}

func (r *TestRunner) SendLearnMessageAs(author *discordgo.User, channel model.Snowflake, message string, learn *Learn) {
	r.T.Helper()

	sendMessageAs(author, r.DiscordSession, r.Handler, channel, message)
	r.DiscordMessagesCount++
	r.Learns[learn.Call] = learn
	assertNewMessages(r.T, r.DiscordSession,
		[]*util.Message{util.NewMessage(channel.Format(), fmt.Sprintf(MsgLearnSuccess, learn.Call))})
	r.AssertState()
	r.SendListMessage(channel)
}

func (r *TestRunner) SendUnlearnMessage(channel model.Snowflake, message string, call string) {
	r.T.Helper()

	sendMessage(r.DiscordSession, r.Handler, channel, message)
	r.DiscordMessagesCount++
	delete(r.Learns, call)
	assertNewMessages(r.T, r.DiscordSession,
		[]*util.Message{util.NewMessage(channel.Format(), fmt.Sprintf(MsgUnlearnSuccess, call))})
	r.AssertState()
	r.SendListMessage(channel)
}

func (r *TestRunner) SendMessageWithoutResponse(channel model.Snowflake, message string) {
	r.T.Helper()

	sendMessage(r.DiscordSession, r.Handler, channel, message)
	assertNewMessages(r.T, r.DiscordSession, []*util.Message{})
	r.AssertState()
}

func (r *TestRunner) SendListMessage(channel model.Snowflake) {
	r.T.Helper()

	sendMessage(r.DiscordSession, r.Handler, channel, "?list")
	r.DiscordMessagesCount++
	r.GistsCount++
	assertNewMessages(r.T, r.DiscordSession, []*util.Message{util.NewMessage(channel.Format(), "The list of commands is here: https://www.example.com/success")})

	// Assert gist state. Cannot be in AssertState because this would fail at the
	// next learn or unlearn.
	// TODO(jake): Remove duplication between this and listfeature. Maybe just assert number of lines?
	if r.GistsCount > 0 {
		var buffer bytes.Buffer
		buffer.WriteString("List of builtins:")
		buffer.WriteString("\n")
		buffer.WriteString(" - ?f1: ")
		buffer.WriteString(vote.MsgHelpBallotInFavor)
		buffer.WriteString("\n")
		buffer.WriteString(" - ?f2: ")
		buffer.WriteString(vote.MsgHelpBallotAgainst)
		buffer.WriteString("\n")
		buffer.WriteString(" - ?help: ")
		buffer.WriteString(MsgHelpHelp)
		buffer.WriteString("\n")
		buffer.WriteString(" - ?learn: ")
		buffer.WriteString(MsgHelpLearn)
		buffer.WriteString("\n")
		buffer.WriteString(" - ?list: ")
		buffer.WriteString(MsgHelpList)
		buffer.WriteString("\n")
		buffer.WriteString(" - ?no: ")
		buffer.WriteString(vote.MsgHelpBallotAgainst)
		buffer.WriteString("\n")
		buffer.WriteString(" - ?ricklist: ")
		buffer.WriteString(moderation.MsgHelpRickListInfo)
		buffer.WriteString("\n")
		buffer.WriteString(" - ?unlearn: ")
		buffer.WriteString(MsgHelpUnlearn)
		buffer.WriteString("\n")
		buffer.WriteString(" - ?vote: ")
		buffer.WriteString(vote.MsgHelpVote)
		buffer.WriteString("\n")
		buffer.WriteString(" - ?votestatus: ")
		buffer.WriteString(vote.MsgHelpStatus)
		buffer.WriteString("\n")
		buffer.WriteString(" - ?yes: ")
		buffer.WriteString(vote.MsgHelpBallotInFavor)
		buffer.WriteString("\n\n")

		buffer.WriteString("List of learned commands:\n")

		all, _ := r.CustomMap.GetAll()
		custom := make([]string, 0, len(all))
		for name := range all {
			custom = append(custom, name)
		}
		sort.Strings(custom)
		for _, name := range custom {
			buffer.WriteString(" - ?")
			buffer.WriteString(name)
			if strings.Contains(all[name], "$1") {
				buffer.WriteString(" <args>")
			}
			buffer.WriteString("\n")
		}

		generated := buffer.String()
		actual := r.Gist.Messages[len(r.Gist.Messages)-1]
		if generated != actual {
			r.T.Fatalf(fmt.Sprintf("Gist failure, got `%v` expected `%v`", actual, generated))
		}
	}

	r.AssertState()
}

func (r *TestRunner) SendVoteStatusMessage(channel model.Snowflake) {
	r.T.Helper()

	sendMessage(r.DiscordSession, r.Handler, channel, "?votestatus")
	r.DiscordMessagesCount++

	activeVote, _ := r.ActiveVoteMap[channel]

	if activeVote == nil {
		assertNewMessages(r.T, r.DiscordSession, []*util.Message{util.NewMessage(channel.Format(), vote.MsgNoActiveVote)})
	} else {
		// Calculate the expected status messages.
		forMessage := vote.MsgOneVoteFor
		if len(activeVote.VotesFor) != 1 {
			forMessage = fmt.Sprintf(vote.MsgVotesFor, len(activeVote.VotesFor))
		}
		againstMessage := vote.MsgOneVoteAgainst
		if len(activeVote.VotesAgainst) != 1 {
			againstMessage = fmt.Sprintf(vote.MsgVotesAgainst, len(r.ActiveVoteMap[channel].VotesAgainst))
		}
		statusMessage := vote.MsgStatusVotesNeeded
		if len(activeVote.VotesAgainst)+len(activeVote.VotesFor) >= 5 {
			if len(activeVote.VotesFor) > len(activeVote.VotesAgainst) {
				statusMessage = vote.MsgStatusVotePassing
			} else {
				statusMessage = vote.MsgStatusVoteFailing
			}
		}

		// The time remaining is independently tested, so just assert its presence.
		timeMessage := vote.TimeString(r.UTCClock, activeVote.TimestampEnd)

		// Build the expected string and assert that it's in the message buffer.
		var buffer bytes.Buffer
		buffer.WriteString(fmt.Sprintf(vote.MsgVoteOwner, activeVote.Author.Username))
		buffer.WriteString(activeVote.Message)
		buffer.WriteString("\n")
		buffer.WriteString(vote.MsgSpacer)
		buffer.WriteString("\n")
		buffer.WriteString(statusMessage)
		buffer.WriteString(". ")
		buffer.WriteString(forMessage)
		buffer.WriteString(", ")
		buffer.WriteString(againstMessage)
		buffer.WriteString(". ")
		buffer.WriteString(timeMessage)
		assertNewMessages(r.T, r.DiscordSession, []*util.Message{util.NewMessage(channel.Format(), buffer.String())})
	}

	r.AssertState()
}

func (r *TestRunner) AddUser(user *discordgo.User) {
	r.DiscordSession.Users[user.ID] = user
}

func assertNumCommands(t *testing.T, customMap *util.InMemoryStringMap, count int) {
	t.Helper()

	if all, _ := customMap.GetAll(); len(all) != count {
		t.Errorf(fmt.Sprintf("Should have %v commands", count))
	}
}

func assertNumGists(t *testing.T, gist *util.InMemoryGist, count int) {
	t.Helper()

	if len(gist.Messages) != count {
		t.Errorf(fmt.Sprintf("Should have %v gists", count))
	}
}

func assertNumDiscordMessages(t *testing.T, discordSession *util.InMemoryDiscordSession, count int) {
	t.Helper()

	if len(discordSession.Messages) != count {
		t.Errorf(fmt.Sprintf("Should have %v discord messages", count))
	}
}

func assertVote(t *testing.T, utcClock model.UTCClock, voteMap *util.InMemoryStringMap, activeVoteMap map[model.Snowflake]*Vote) {
	t.Helper()

	modelHelper := vote.NewModelHelper(voteMap, utcClock)
	for channel, vote := range activeVoteMap {
		ok, _ := modelHelper.IsVoteActive(channel)
		if vote != nil && !ok {
			t.Errorf("Expected a vote to be active, but was not")
		}
		if vote == nil && ok {
			t.Errorf("Expected a vote to not be active, but one was")
		}
	}
}

func sendMessage(discordSession api.DiscordSession, handler func(api.DiscordSession, *discordgo.MessageCreate), channel model.Snowflake, message string) {
	author := &discordgo.User{
		ID:            "1",
		Email:         "email@example.com",
		Username:      "username",
		Avatar:        "avatar",
		Discriminator: "discriminator",
		Token:         "token",
		Verified:      true,
		MFAEnabled:    false,
		Bot:           false,
	}

	sendMessageAs(author, discordSession, handler, channel, message)
}

func sendMessageAs(author *discordgo.User, discordSession api.DiscordSession, handler func(api.DiscordSession, *discordgo.MessageCreate), channel model.Snowflake, message string) {
	messageCreate := &discordgo.MessageCreate{
		&discordgo.Message{
			ID:              "messageID",
			ChannelID:       channel.Format(),
			Content:         message,
			Timestamp:       "timestamp",
			EditedTimestamp: "edited timestamp",
			MentionRoles:    []string{},
			Tts:             false,
			MentionEveryone: false,
			Author:          author,
			Attachments:     []*discordgo.MessageAttachment{},
			Embeds:          []*discordgo.MessageEmbed{},
			Mentions:        []*discordgo.User{},
			Reactions:       []*discordgo.MessageReactions{},
		},
	}
	handler(discordSession, messageCreate)
}

func assertNewMessages(t *testing.T, discordSession *util.InMemoryDiscordSession, newMessages []*util.Message) {
	t.Helper()

	if len(discordSession.Messages) < len(newMessages) {
		t.Errorf(fmt.Sprintf(
			"Needed at least %v messages, had %v", len(newMessages), len(discordSession.Messages)))
		return
	}

	for i := 0; i < len(newMessages); i++ {
		actualMessage := discordSession.Messages[len(discordSession.Messages)-len(newMessages)+i]
		if !reflect.DeepEqual(newMessages[i], actualMessage) {
			t.Errorf("Expected message \n '%v' \n on channel '%v', got message \n '%v' \n on channel '%v'",
				newMessages[i].Message,
				newMessages[i].Channel,
				actualMessage.Message,
				actualMessage.Channel)
		}
	}
}

func assertCommand(t *testing.T, commandMap *util.InMemoryStringMap, call, response string) {
	t.Helper()

	if _, err := commandMap.Get(call); err != nil {
		t.Errorf("Response should be present for call " + call)
		return
	}
	if value, _ := commandMap.Get(call); value != response {
		t.Errorf(fmt.Sprintf("Wrong response for %v, expected %v got %v", call, response, value))
	}
}

type Learn struct {
	Call     string
	Response string
}

func NewLearn(call, response string) *Learn {
	return &Learn{
		Call:     call,
		Response: response,
	}
}

func newUser(name string, id model.Snowflake, bot bool) *discordgo.User {
	idStr := id.Format()
	return &discordgo.User{
		ID:            idStr,
		Email:         "email@example.com",
		Username:      name,
		Avatar:        "avatar",
		Discriminator: idStr,
		Token:         "token",
		Verified:      true,
		MFAEnabled:    false,
		Bot:           bot,
	}
}

// Enough information to reconstruct the status message.
type Vote struct {
	Channel      model.Snowflake
	Author       *discordgo.User
	Message      string
	VotesFor     []model.Snowflake
	VotesAgainst []model.Snowflake
	TimestampEnd time.Time
}

func newVote(channel model.Snowflake, author *discordgo.User, message string, timestampEnd time.Time) *Vote {
	return &Vote{
		Channel:      channel,
		Author:       author,
		Message:      message,
		VotesFor:     []model.Snowflake{},
		VotesAgainst: []model.Snowflake{},
		TimestampEnd: timestampEnd,
	}
}
