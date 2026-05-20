package provenance

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	provenancepkg "github.com/langoai/lango/internal/provenance"
)

func TestRenderSessionTreeAndEmptyDashHelpers(t *testing.T) {
	t.Parallel()

	nodes := []provenancepkg.SessionNode{
		{SessionKey: "root", Status: provenancepkg.SessionStatusCompleted},
		{SessionKey: "child-a", ParentKey: "root", Status: provenancepkg.SessionStatusActive},
		{SessionKey: "grandchild", ParentKey: "child-a", Status: provenancepkg.SessionStatusMerged},
		{SessionKey: "child-b", ParentKey: "root", Status: provenancepkg.SessionStatusDiscarded},
		{SessionKey: "orphan-child", ParentKey: "missing", Status: provenancepkg.SessionStatusActive},
		{SessionKey: "second-root", Status: provenancepkg.SessionStatusActive},
	}

	lines := renderSessionTree(nodes)

	assert.Equal(t, []string{
		"root [completed]",
		"  child-a [active]",
		"    grandchild [merged]",
		"  child-b [discarded]",
		"second-root [active]",
	}, lines)
	assert.Equal(t, "-", emptyDash(""))
	assert.Equal(t, "parent", emptyDash("parent"))
}

func TestLoadSignerRejectsUnavailableBootstrapPrerequisites(t *testing.T) {
	t.Parallel()

	did, signer, err := loadSigner(context.Background(), nil)
	require.Error(t, err)
	assert.Empty(t, did)
	assert.Nil(t, signer)
	assert.EqualError(t, err, "signed provenance export requires initialized bootstrap crypto")

	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = false
	did, signer, err = loadSigner(context.Background(), &bootstrap.Result{Config: cfg})
	require.Error(t, err)
	assert.Empty(t, did)
	assert.Nil(t, signer)
	assert.EqualError(t, err, "signed provenance export requires initialized bootstrap crypto")
}

func TestLoadSignerRejectsDisabledPaymentAfterCryptoInitialization(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = false
	boot := &bootstrap.Result{
		Config: cfg,
		Crypto: provenanceCryptoProvider{},
	}

	did, signer, err := loadSigner(context.Background(), boot)

	require.Error(t, err)
	assert.Empty(t, did)
	assert.Nil(t, signer)
	assert.EqualError(t, err, "signed provenance export requires payment.enabled=true")
}

func TestLoadSignerRejectsMissingPersistentSecurityStores(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = true
	cfg.Payment.WalletProvider = "local"
	boot := &bootstrap.Result{
		Config: cfg,
		Crypto: provenanceCryptoProvider{},
	}

	did, signer, err := loadSigner(context.Background(), boot)

	require.Error(t, err)
	assert.Empty(t, did)
	assert.Nil(t, signer)
	assert.EqualError(t, err, "signed provenance export requires persistent security stores")
}

type provenanceCryptoProvider struct{}

func (provenanceCryptoProvider) Sign(_ context.Context, _ string, payload []byte) ([]byte, error) {
	return append([]byte(nil), payload...), nil
}

func (provenanceCryptoProvider) Encrypt(_ context.Context, _ string, plaintext []byte) ([]byte, error) {
	return append([]byte(nil), plaintext...), nil
}

func (provenanceCryptoProvider) Decrypt(_ context.Context, _ string, ciphertext []byte) ([]byte, error) {
	return append([]byte(nil), ciphertext...), nil
}
