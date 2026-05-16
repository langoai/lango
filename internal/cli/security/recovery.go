package security

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/prompt"
	"github.com/langoai/lango/internal/keyring"
	"github.com/langoai/lango/internal/security"
)

var executeRecoverySetup = func(cmd *cobra.Command, bootLoader func() (*bootstrap.Result, error)) error {
	if err := prompt.RequireInteractiveTerminal("recovery setup requires an interactive terminal"); err != nil {
		return err
	}
	boot, err := bootLoader()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	defer boot.Close()

	provider, ok := boot.Crypto.(*security.LocalCryptoProvider)
	if !ok {
		return fmt.Errorf("recovery setup is only available for the local crypto provider")
	}
	envelope := provider.Envelope()
	if envelope == nil {
		return fmt.Errorf("envelope not found — recovery requires envelope-based encryption")
	}
	if envelope.HasSlotType(security.KEKSlotMnemonic) {
		return fmt.Errorf("recovery mnemonic slot already exists; remove it first")
	}

	currentPass, err := prompt.Passphrase("Enter current passphrase to authorize setup: ")
	if err != nil {
		return err
	}
	mk, _, err := envelope.UnwrapFromPassphrase(currentPass)
	if err != nil {
		return fmt.Errorf("current passphrase is incorrect")
	}
	defer security.ZeroBytes(mk)

	mnemonic, err := security.GenerateRecoveryMnemonic()
	if err != nil {
		return fmt.Errorf("generate mnemonic: %w", err)
	}

	out := cmd.OutOrStdout()
	fmt.Fprintln(out)
	fmt.Fprintln(out, "============================================================")
	fmt.Fprintln(out, "RECOVERY MNEMONIC — write this down and store securely")
	fmt.Fprintln(out, "============================================================")
	words := strings.Fields(mnemonic)
	for i, w := range words {
		fmt.Fprintf(out, "%2d. %-10s", i+1, w)
		if (i+1)%4 == 0 {
			fmt.Fprintln(out)
		}
	}
	fmt.Fprintln(out, "============================================================")
	fmt.Fprintln(out)

	ok2, err := prompt.ConfirmDenyOnEOFIO(cmd.InOrStdin(), out, "Have you written down all 24 words?")
	if err != nil || !ok2 {
		return fmt.Errorf("setup aborted")
	}

	idx1, idx2 := pickConfirmationIndexes(len(words))
	if err := confirmWord(cmd.InOrStdin(), out, words, idx1); err != nil {
		return err
	}
	if err := confirmWord(cmd.InOrStdin(), out, words, idx2); err != nil {
		return err
	}

	if err := envelope.AddSlot(security.KEKSlotMnemonic, "recovery", mk, mnemonic, security.NewDefaultKDFParams()); err != nil {
		return fmt.Errorf("add mnemonic slot: %w", err)
	}
	if err := security.StoreEnvelopeFile(boot.LangoDir, envelope); err != nil {
		return fmt.Errorf("persist envelope: %w", err)
	}

	fmt.Fprintln(out, "Recovery mnemonic slot added successfully.")
	return nil
}

var executeRecoveryRestore = func(cmd *cobra.Command, errWriter io.Writer) error {
	if err := prompt.RequireInteractiveTerminal("recovery restore requires an interactive terminal"); err != nil {
		return err
	}

	langoDir := defaultLangoDir()
	if langoDir == "" {
		return fmt.Errorf("resolve home directory: cannot determine lango data directory")
	}

	envelope, err := security.LoadEnvelopeFile(langoDir)
	if err != nil {
		return fmt.Errorf("load envelope: %w", err)
	}
	if envelope == nil {
		return fmt.Errorf("envelope not found — recovery requires local encryption mode")
	}
	if !envelope.HasSlotType(security.KEKSlotMnemonic) {
		return fmt.Errorf("no recovery mnemonic slot on this envelope")
	}

	mnemonic, err := prompt.Passphrase("Enter 24-word recovery mnemonic: ")
	if err != nil {
		return err
	}
	if err := security.ValidateMnemonic(mnemonic); err != nil {
		return fmt.Errorf("invalid mnemonic: %w", err)
	}
	mk, _, err := envelope.UnwrapFromMnemonic(mnemonic)
	if err != nil {
		return fmt.Errorf("mnemonic did not match any recovery slot")
	}
	defer security.ZeroBytes(mk)

	newPass, err := prompt.PassphraseConfirm("Enter NEW passphrase: ", "Confirm NEW passphrase: ")
	if err != nil {
		return err
	}
	if err := envelope.ChangePassphraseSlot(mk, newPass); err != nil {
		return fmt.Errorf("replace passphrase slot: %w", err)
	}
	if err := security.StoreEnvelopeFile(langoDir, envelope); err != nil {
		return fmt.Errorf("persist envelope: %w", err)
	}

	keyfilePath := filepath.Join(langoDir, "keyfile")
	if _, statErr := os.Stat(keyfilePath); statErr == nil {
		if writeErr := os.WriteFile(keyfilePath, []byte(newPass), 0600); writeErr != nil {
			fmt.Fprintf(errWriter, "warning: update keyfile: %v\n", writeErr)
		} else {
			fmt.Fprintln(errWriter, "Keyfile updated with new passphrase.")
		}
	}
	if secureProvider, _ := keyring.DetectSecureProvider(); secureProvider != nil {
		if setErr := secureProvider.Set(keyring.Service, keyring.KeyMasterPassphrase, newPass); setErr != nil {
			fmt.Fprintf(errWriter, "warning: keyring update failed: %v\n", setErr)
			fmt.Fprintf(errWriter, "  If a stale passphrase is stored, next headless bootstrap may fail.\n")
			fmt.Fprintf(errWriter, "  Fix: run `lango security keyring set` or clear the keyring entry manually.\n")
		} else {
			fmt.Fprintln(errWriter, "Keyring updated with new passphrase.")
		}
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Recovery complete. The new passphrase is now active.")
	return nil
}

// newRecoveryCmd bundles `lango security recovery setup` and `lango security recovery restore`.
func newRecoveryCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recovery",
		Short: "Manage recovery mnemonic for the Master Key envelope",
	}
	cmd.AddCommand(newRecoverySetupCmd(bootLoader))
	cmd.AddCommand(newRecoveryRestoreCmd())
	return cmd
}

func newRecoverySetupCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Generate a BIP39 recovery mnemonic and add it as a KEK slot",
		Long: `Generate a 24-word BIP39 recovery mnemonic and install it as a new KEK slot
on the Master Key envelope.

The mnemonic is displayed exactly once. You MUST write it down and store it
securely — losing both your passphrase and your mnemonic means permanent loss
of access to all encrypted data.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeRecoverySetup(cmd, bootLoader)
		},
	}
}

// newRecoveryRestoreCmd creates the restore command that loads the envelope
// directly from the filesystem without running the full bootstrap pipeline.
// This allows recovery even when the user has lost their passphrase.
func newRecoveryRestoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restore",
		Short: "Recover access using the BIP39 recovery mnemonic",
		Long: `Recover access to the Master Key envelope using the BIP39 recovery mnemonic.

You will be prompted for the 24-word mnemonic and then for a new passphrase.
The new passphrase replaces the existing passphrase slot; the recovery slot
and any other slots are unchanged.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeRecoveryRestore(cmd, cmd.ErrOrStderr())
		},
	}
}

func pickConfirmationIndexes(n int) (int, int) {
	var b [8]byte
	_, _ = rand.Read(b[:])
	u := binary.BigEndian.Uint64(b[:])
	i := int(u%uint64(n)) + 1
	j := int((u>>32)%uint64(n)) + 1
	if j == i {
		j = (j % n) + 1
	}
	return i, j
}

func confirmWord(in io.Reader, out io.Writer, words []string, index int) error {
	got, err := prompt.ReadLineIO(in, out, fmt.Sprintf("Enter word %d to confirm: ", index))
	if err != nil && !(errors.Is(err, io.EOF) && strings.TrimSpace(got) != "") {
		return fmt.Errorf("read confirmation word: %w", err)
	}
	if strings.TrimSpace(strings.ToLower(got)) != words[index-1] {
		return fmt.Errorf("confirmation word %d did not match", index)
	}
	return nil
}
