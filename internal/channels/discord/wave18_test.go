package discord

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type wave18Session struct {
	state *discordgo.State

	sendErr        error
	complexSendErr error
	typingErr      error

	sentChannels  []string
	sentMessages  []string
	complexSends  []*discordgo.MessageSend
	typingCalls   []string
	closeCalls    int
	commandNames  []string
	registeredFns []interface{}
	respondCalls  int
}

func newWave18Session() *wave18Session {
	state := &discordgo.State{}
	state.User = &discordgo.User{
		ID:       "bot-18",
		Username: "Wave18Bot",
	}

	return &wave18Session{state: state}
}

func (s *wave18Session) Open() error {
	return nil
}

func (s *wave18Session) Close() error {
	s.closeCalls++
	return nil
}

func (s *wave18Session) AddHandler(handler interface{}) func() {
	s.registeredFns = append(s.registeredFns, handler)
	return func() {}
}

func (s *wave18Session) ChannelMessageSend(
	channelID string,
	content string,
	options ...discordgo.RequestOption,
) (*discordgo.Message, error) {
	if s.sendErr != nil {
		return nil, s.sendErr
	}
	s.sentChannels = append(s.sentChannels, channelID)
	s.sentMessages = append(s.sentMessages, content)
	return &discordgo.Message{ID: "wave18-msg", Content: content}, nil
}

func (s *wave18Session) ChannelMessageSendComplex(
	channelID string,
	data *discordgo.MessageSend,
	options ...discordgo.RequestOption,
) (*discordgo.Message, error) {
	if s.complexSendErr != nil {
		return nil, s.complexSendErr
	}
	s.sentChannels = append(s.sentChannels, channelID)
	s.complexSends = append(s.complexSends, data)
	return &discordgo.Message{ID: "wave18-complex-msg", Content: data.Content}, nil
}

func (s *wave18Session) ChannelMessageEditComplex(
	edit *discordgo.MessageEdit,
	options ...discordgo.RequestOption,
) (*discordgo.Message, error) {
	return &discordgo.Message{}, nil
}

func (s *wave18Session) ChannelTyping(
	channelID string,
	options ...discordgo.RequestOption,
) error {
	s.typingCalls = append(s.typingCalls, channelID)
	return s.typingErr
}

func (s *wave18Session) InteractionRespond(
	interaction *discordgo.Interaction,
	resp *discordgo.InteractionResponse,
	options ...discordgo.RequestOption,
) error {
	s.respondCalls++
	return nil
}

func (s *wave18Session) ApplicationCommandCreate(
	appID string,
	guildID string,
	cmd *discordgo.ApplicationCommand,
	options ...discordgo.RequestOption,
) (*discordgo.ApplicationCommand, error) {
	s.commandNames = append(s.commandNames, cmd.Name)
	return cmd, nil
}

func (s *wave18Session) GetState() *discordgo.State {
	return s.state
}

type wave18UserMessageError struct {
	internal string
	user     string
}

func (e wave18UserMessageError) Error() string {
	return e.internal
}

func (e wave18UserMessageError) UserMessage() string {
	return e.user
}

func TestWave18ChannelMetadataApprovalAndStop(t *testing.T) {
	t.Parallel()

	session := newWave18Session()
	channel, err := New(Config{BotToken: "TEST_TOKEN", Session: session})
	require.NoError(t, err)

	assert.Equal(t, "discord", channel.Name())
	assert.NotNil(t, channel.GetApprovalProvider())

	require.NoError(t, channel.Stop(context.Background()))
	assert.Equal(t, 1, session.closeCalls)
}

func TestWave18GuildMentionAndContentHelpers(t *testing.T) {
	t.Parallel()

	channel := &Channel{
		botID: "bot-18",
		config: Config{
			AllowedGuilds: []string{"guild-allowed", "guild-other"},
		},
	}

	assert.True(t, (&Channel{}).isGuildAllowed("any-guild"))
	assert.True(t, channel.isGuildAllowed("guild-allowed"))
	assert.False(t, channel.isGuildAllowed("guild-denied"))

	msg := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Mentions: []*discordgo.User{
				{ID: "someone-else"},
				{ID: "bot-18"},
			},
		},
	}
	assert.True(t, channel.isBotMentioned(msg))

	msg.Mentions = []*discordgo.User{{ID: "someone-else"}}
	assert.False(t, channel.isBotMentioned(msg))

	tests := []struct {
		name string
		give string
		want string
	}{
		{
			name: "plain mention",
			give: "  <@bot-18> hello there  ",
			want: "hello there",
		},
		{
			name: "nickname mention",
			give: "<@!bot-18>\nplease help",
			want: "please help",
		},
		{
			name: "multiple bot mentions",
			give: "<@bot-18> summarize <@!bot-18>",
			want: "summarize",
		},
		{
			name: "other mentions preserved",
			give: "<@bot-18> ask <@user-1>",
			want: "ask <@user-1>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, channel.cleanContent(tt.give))
		})
	}
}

func TestWave18SplitMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		max  int
		want []string
	}{
		{
			name: "short message remains one chunk",
			text: "short",
			max:  10,
			want: []string{"short"},
		},
		{
			name: "exact limit remains one chunk",
			text: "12345",
			max:  5,
			want: []string{"12345"},
		},
		{
			name: "splits on last newline before limit",
			text: "alpha\nbeta gamma",
			max:  10,
			want: []string{"alpha", "beta gamma"},
		},
		{
			name: "hard splits when no newline exists",
			text: "abcdefghij",
			max:  4,
			want: []string{"abcd", "efgh", "ij"},
		},
		{
			name: "drops separator newline after split",
			text: "abcd\nef",
			max:  5,
			want: []string{"abcd", "ef"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, splitMessage(tt.text, tt.max))
		})
	}
}

func TestWave18SendTextSplitsAndStopsOnError(t *testing.T) {
	t.Parallel()

	session := newWave18Session()
	channel := &Channel{session: session}
	content := strings.Repeat("a", 1999) + "\n" + strings.Repeat("b", 10)

	err := channel.Send("chan-18", &OutgoingMessage{Content: content})
	require.NoError(t, err)

	assert.Equal(t, []string{"chan-18", "chan-18"}, session.sentChannels)
	assert.Equal(t, []string{strings.Repeat("a", 1999), strings.Repeat("b", 10)}, session.sentMessages)

	sendErr := errors.New("discord send failed")
	errorSession := newWave18Session()
	errorSession.sendErr = sendErr
	errorChannel := &Channel{session: errorSession}

	err = errorChannel.Send("chan-18", &OutgoingMessage{Content: "hello"})
	require.ErrorIs(t, err, sendErr)
	assert.Empty(t, errorSession.sentMessages)
}

func TestWave18SendEmbedMapsFieldsAndReturnsComplexError(t *testing.T) {
	t.Parallel()

	session := newWave18Session()
	channel := &Channel{session: session}

	msg := &OutgoingMessage{
		Content: "embed body",
		Embed: &Embed{
			Title:       "Status",
			Description: "Details",
			Color:       0x00ff99,
			Fields: []EmbedField{
				{Name: "Result", Value: "ok", Inline: true},
				{Name: "Next", Value: "none"},
			},
		},
	}

	require.NoError(t, channel.Send("chan-18", msg))
	require.Len(t, session.complexSends, 1)

	got := session.complexSends[0]
	require.NotNil(t, got.Embed)
	assert.Equal(t, "embed body", got.Content)
	assert.Equal(t, "Status", got.Embed.Title)
	assert.Equal(t, "Details", got.Embed.Description)
	assert.Equal(t, 0x00ff99, got.Embed.Color)
	require.Len(t, got.Embed.Fields, 2)
	assert.Equal(t, "Result", got.Embed.Fields[0].Name)
	assert.Equal(t, "ok", got.Embed.Fields[0].Value)
	assert.True(t, got.Embed.Fields[0].Inline)
	assert.Equal(t, "Next", got.Embed.Fields[1].Name)
	assert.Equal(t, "none", got.Embed.Fields[1].Value)
	assert.False(t, got.Embed.Fields[1].Inline)

	complexErr := errors.New("discord complex send failed")
	errorSession := newWave18Session()
	errorSession.complexSendErr = complexErr
	errorChannel := &Channel{session: errorSession}

	err := errorChannel.Send("chan-18", msg)
	require.ErrorIs(t, err, complexErr)
	assert.Empty(t, errorSession.complexSends)
}

func TestWave18FormatChannelErrorAndSendError(t *testing.T) {
	t.Parallel()

	assert.Equal(
		t,
		"Please retry with a smaller request.",
		formatChannelError(wave18UserMessageError{
			internal: "token budget exceeded",
			user:     "Please retry with a smaller request.",
		}),
	)
	assert.Equal(t, "Error: plain failure", formatChannelError(errors.New("plain failure")))

	session := newWave18Session()
	channel := &Channel{session: session}

	channel.sendError("chan-18", wave18UserMessageError{
		internal: "internal details",
		user:     "Safe user message",
	})

	require.Equal(t, []string{"chan-18"}, session.sentChannels)
	assert.Equal(t, []string{"❌ Safe user message"}, session.sentMessages)
}

func TestWave18TypingStartStopIsDeterministic(t *testing.T) {
	t.Parallel()

	session := newWave18Session()
	channel := &Channel{session: session}

	ctx, cancel := context.WithCancel(context.Background())
	stop := channel.StartTyping(ctx, "chan-public")
	cancel()
	stop()
	stop()

	assert.Equal(t, []string{"chan-public"}, session.typingCalls)

	privateSession := newWave18Session()
	privateSession.typingErr = errors.New("typing rejected")
	privateChannel := &Channel{session: privateSession}

	privateStop := privateChannel.startTyping("chan-private")
	privateStop()
	privateStop()

	assert.Equal(t, []string{"chan-private"}, privateSession.typingCalls)
}

func TestWave18StartRegistersCommandsWhenApplicationIDIsConfigured(t *testing.T) {
	t.Parallel()

	session := newWave18Session()
	channel, err := New(Config{
		BotToken:      "TEST_TOKEN",
		ApplicationID: "app-18",
		Session:       session,
	})
	require.NoError(t, err)
	channel.SetHandler(func(context.Context, *IncomingMessage) (*OutgoingMessage, error) {
		return nil, nil
	})

	require.NoError(t, channel.Start(context.Background()))

	assert.Equal(t, []string{"ask", "clear"}, session.commandNames)
	assert.Len(t, session.registeredFns, 2)
}

func TestWave18InteractionCreateRoutesOnlyMessageComponents(t *testing.T) {
	t.Parallel()

	session := newWave18Session()
	channel := &Channel{
		session:  session,
		approval: NewApprovalProvider(session, 0),
	}

	channel.onInteractionCreate(nil, &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{Type: discordgo.InteractionApplicationCommand},
	})
	assert.Zero(t, session.respondCalls)

	channel.onInteractionCreate(nil, &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionMessageComponent,
			Data: discordgo.MessageComponentInteractionData{
				CustomID: "approve:missing-request",
			},
		},
	})
	assert.Equal(t, 1, session.respondCalls)
}
