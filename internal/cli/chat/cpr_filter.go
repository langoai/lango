package chat

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// cprState tracks the CPR (Cursor Position Report) filter state machine.
type cprState int

const (
	cprIdle     cprState = iota
	cprGotEsc
	cprGotBracket
	cprInParams
	cprGotOSC
	cprInOSC
	cprOscEsc
)

// cprTimeoutMsg is sent when a CPR detection window expires.
type cprTimeoutMsg struct{}

// cprTimeout is how long we wait after ESC before deciding it's not a CPR sequence.
const cprTimeout = 50 * time.Millisecond

// cprFilter is a standalone state machine that intercepts terminal response
// sequences (CPR, OSC) before they reach the input composer. It buffers
// ambiguous key messages until the sequence is identified as a terminal
// response (discarded) or a normal key sequence (flushed for replay).
type cprFilter struct {
	state cprState
	buf   []tea.KeyMsg
}

// Filter processes a single key message through the CPR state machine.
// It returns filtered=true when the key was consumed (buffered or discarded).
// Any returned cmds (e.g. timeout tick) should be batched by the caller.
//
// When the state machine determines the buffered sequence is NOT a terminal
// response, it returns filtered=false and leaves previously buffered keys in
// buf for the caller to retrieve via Flush. The caller should check
// len(f.buf) > 0 after a filtered=false return to detect this case.
func (f *cprFilter) Filter(msg tea.KeyMsg) (filtered bool, cmds []tea.Cmd) {
	switch f.state {
	case cprIdle:
		if msg.Type == tea.KeyEscape {
			f.state = cprGotEsc
			f.buf = append(f.buf[:0], msg)
			return true, []tea.Cmd{tea.Tick(cprTimeout, func(time.Time) tea.Msg {
				return cprTimeoutMsg{}
			})}
		}
		return false, nil

	case cprGotEsc:
		if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == '[' {
			f.state = cprGotBracket
			f.buf = append(f.buf, msg)
			return true, nil
		}
		if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == ']' {
			f.state = cprGotOSC
			f.buf = append(f.buf, msg)
			return true, nil
		}
		// Not a CSI or OSC sequence — needs flush and current msg pass-through.
		f.markNeedsFlush()
		return false, nil

	case cprGotBracket, cprInParams:
		if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
			r := msg.Runes[0]
			if (r >= '0' && r <= '9') || r == ';' {
				f.state = cprInParams
				f.buf = append(f.buf, msg)
				return true, nil
			}
			if r == 'R' && f.state == cprInParams {
				// Complete CPR sequence — discard everything.
				f.state = cprIdle
				f.buf = f.buf[:0]
				return true, nil
			}
		}
		// Not a valid CPR continuation — needs flush.
		f.markNeedsFlush()
		return false, nil

	case cprGotOSC, cprInOSC:
		if msg.Type == tea.KeyCtrlG || (msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == '\a') {
			// OSC terminated with BEL — discard.
			f.state = cprIdle
			f.buf = f.buf[:0]
			return true, nil
		}
		if msg.Type == tea.KeyEscape {
			f.state = cprOscEsc
			f.buf = append(f.buf, msg)
			return true, nil
		}
		f.state = cprInOSC
		f.buf = append(f.buf, msg)
		return true, nil

	case cprOscEsc:
		if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == '\\' {
			// OSC terminated with ST (ESC \) — discard.
			f.state = cprIdle
			f.buf = f.buf[:0]
			return true, nil
		}
		f.state = cprInOSC
		f.buf = append(f.buf, msg)
		return true, nil
	}

	return false, nil
}

// markNeedsFlush is called internally when Filter determines the buffered
// sequence is not a terminal response. It does NOT clear the buffer — the
// ChatModel wrapper calls Flush to retrieve and replay buffered keys.
func (f *cprFilter) markNeedsFlush() {
	f.state = cprIdle
	// buf is intentionally left intact for the caller to Flush.
}

// Flush returns the buffered keys and resets state to cprIdle.
// The caller (ChatModel) is responsible for replaying the returned keys.
func (f *cprFilter) Flush() []tea.KeyMsg {
	f.state = cprIdle
	keys := make([]tea.KeyMsg, len(f.buf))
	copy(keys, f.buf)
	f.buf = f.buf[:0]
	return keys
}

// HandleTimeout handles a cprTimeoutMsg. If the state machine is mid-sequence,
// it flushes the buffer; otherwise it returns nil.
func (f *cprFilter) HandleTimeout() []tea.KeyMsg {
	if f.state != cprIdle {
		return f.Flush()
	}
	return nil
}
