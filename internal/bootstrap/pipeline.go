package bootstrap

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/keyring"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/security/passphrase"
)

// State carries data between pipeline phases.
// Each phase can read from and write to State.
type State struct {
	Options Options
	Result  Result

	// Internal state passed between phases.
	Home    string
	LangoDir string

	// Encryption detection.
	DBEncrypted bool
	NeedsDBKey  bool

	// Passphrase acquisition.
	Passphrase     string
	PassSource     passphrase.Source
	SecureProvider keyring.Provider
	SecurityTier   keyring.SecurityTier
	FirstRunGuess  bool

	// Database handles (set by phaseOpenDatabase).
	Client *ent.Client
	RawDB  *sql.DB

	// Security state from DB.
	Salt     []byte
	Checksum []byte
	FirstRun bool

	// Crypto.
	DBKey    string
	Crypto   security.CryptoProvider
}

// Phase represents a single step in the bootstrap pipeline.
type Phase struct {
	Name    string
	Run     func(ctx context.Context, state *State) error
	Cleanup func(state *State) // called in reverse order if a later phase fails
}

// Pipeline executes phases sequentially. If a phase fails,
// cleanup functions of all previously completed phases are called in reverse order.
type Pipeline struct {
	phases []Phase
}

// NewPipeline creates a pipeline from the given phases.
func NewPipeline(phases ...Phase) *Pipeline {
	return &Pipeline{phases: phases}
}

// Execute runs all phases. On failure, cleans up in reverse order.
func (p *Pipeline) Execute(ctx context.Context, opts Options) (*Result, error) {
	state := &State{Options: opts}

	var completed []int // indices of completed phases

	for i, phase := range p.phases {
		if err := phase.Run(ctx, state); err != nil {
			// Cleanup in reverse order.
			for j := len(completed) - 1; j >= 0; j-- {
				idx := completed[j]
				if p.phases[idx].Cleanup != nil {
					p.phases[idx].Cleanup(state)
				}
			}
			return nil, fmt.Errorf("%s: %w", phase.Name, err)
		}
		completed = append(completed, i)
	}

	return &state.Result, nil
}
