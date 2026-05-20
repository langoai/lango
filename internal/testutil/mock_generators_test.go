package testutil

import (
	"context"
	"errors"
	"testing"
)

func TestMockTextGeneratorRecordsCallsAndErrors(t *testing.T) {
	gen := NewMockTextGenerator("response")

	got, err := gen.GenerateText(context.Background(), "system", "user")
	if err != nil {
		t.Fatalf("GenerateText() returned error: %v", err)
	}
	if got != "response" {
		t.Fatalf("GenerateText() = %q, want response", got)
	}
	if calls := gen.Calls(); calls != 1 {
		t.Fatalf("Calls() = %d, want 1", calls)
	}
	system, user := gen.LastArgs()
	if system != "system" || user != "user" {
		t.Fatalf("LastArgs() = (%q, %q), want (system, user)", system, user)
	}

	gen.Err = errors.New("generator failed")
	if _, err := gen.GenerateText(context.Background(), "s2", "u2"); !errors.Is(err, gen.Err) {
		t.Fatalf("GenerateText() error = %v, want %v", err, gen.Err)
	}
}

func TestMockAgentRunnerRecordsSessionAndErrors(t *testing.T) {
	runner := NewMockAgentRunner("done")

	got, err := runner.Run(context.Background(), "session-1", "prompt")
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
	if got != "done" {
		t.Fatalf("Run() = %q, want done", got)
	}
	if calls := runner.Calls(); calls != 1 {
		t.Fatalf("Calls() = %d, want 1", calls)
	}
	if session := runner.LastSessionKey(); session != "session-1" {
		t.Fatalf("LastSessionKey() = %q, want session-1", session)
	}

	runner.Err = errors.New("runner failed")
	if _, err := runner.Run(context.Background(), "session-2", "prompt"); !errors.Is(err, runner.Err) {
		t.Fatalf("Run() error = %v, want %v", err, runner.Err)
	}
}

func TestMockChannelSenderRecordsMessagesAndReturnsCopies(t *testing.T) {
	sender := NewMockChannelSender()

	if err := sender.SendMessage(context.Background(), "telegram", "hello"); err != nil {
		t.Fatalf("SendMessage() returned error: %v", err)
	}
	if calls := sender.Calls(); calls != 1 {
		t.Fatalf("Calls() = %d, want 1", calls)
	}
	messages := sender.Messages()
	if len(messages) != 1 || messages[0].Channel != "telegram" || messages[0].Message != "hello" {
		t.Fatalf("Messages() = %#v, want recorded telegram message", messages)
	}

	messages[0].Message = "mutated"
	if got := sender.Messages()[0].Message; got != "hello" {
		t.Fatalf("Messages() exposed internal slice, got %q", got)
	}

	sender.Err = errors.New("send failed")
	if err := sender.SendMessage(context.Background(), "slack", "fail"); !errors.Is(err, sender.Err) {
		t.Fatalf("SendMessage() error = %v, want %v", err, sender.Err)
	}
}
