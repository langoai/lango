//go:build kms_pkcs11 || kms_all

package security

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/logging"
	"github.com/miekg/pkcs11"
)

const (
	pkcs11GCMIVSize  = 12
	pkcs11GCMTagBits = 128
)

var pkcs11Logger = logging.SubsystemSugar("pkcs11")

// PKCS11Provider implements CryptoProvider using a PKCS#11 HSM.
type PKCS11Provider struct {
	mu         sync.Mutex
	ctx        *pkcs11.Ctx
	session    pkcs11.SessionHandle
	keyMap     map[string]pkcs11.ObjectHandle
	config     config.PKCS11Config
	maxRetries int
	timeout    time.Duration
}

var _ CryptoProvider = (*PKCS11Provider)(nil)

func newPKCS11Provider(kmsConfig config.KMSConfig) (CryptoProvider, error) {
	if kmsConfig.PKCS11.ModulePath == "" {
		return nil, fmt.Errorf("new PKCS#11 provider: module path is required")
	}

	p11ctx := pkcs11.New(kmsConfig.PKCS11.ModulePath)
	if p11ctx == nil {
		return nil, &KMSError{
			Provider: "pkcs11",
			Op:       "init",
			KeyID:    "",
			Err:      fmt.Errorf("%w: load module %s", ErrPKCS11Module, kmsConfig.PKCS11.ModulePath),
		}
	}

	if err := p11ctx.Initialize(); err != nil {
		p11ctx.Destroy()
		return nil, &KMSError{
			Provider: "pkcs11",
			Op:       "init",
			KeyID:    "",
			Err:      fmt.Errorf("%w: initialize: %v", ErrPKCS11Module, err),
		}
	}

	session, err := p11ctx.OpenSession(uint(kmsConfig.PKCS11.SlotID), pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION)
	if err != nil {
		p11ctx.Finalize()
		p11ctx.Destroy()
		return nil, &KMSError{
			Provider: "pkcs11",
			Op:       "open-session",
			KeyID:    "",
			Err:      fmt.Errorf("%w: slot %d: %v", ErrPKCS11Session, kmsConfig.PKCS11.SlotID, err),
		}
	}

	// PIN: environment variable takes priority over config.
	pin := os.Getenv("LANGO_PKCS11_PIN")
	if pin == "" {
		pin = kmsConfig.PKCS11.Pin
	}

	if pin != "" {
		if err := p11ctx.Login(session, pkcs11.CKU_USER, pin); err != nil {
			p11ctx.CloseSession(session)
			p11ctx.Finalize()
			p11ctx.Destroy()
			return nil, &KMSError{
				Provider: "pkcs11",
				Op:       "login",
				KeyID:    "",
				Err:      fmt.Errorf("%w: %v", ErrPKCS11Session, err),
			}
		}
	}

	maxRetries := kmsConfig.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}
	timeout := kmsConfig.TimeoutPerOperation
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	provider := &PKCS11Provider{
		ctx:        p11ctx,
		session:    session,
		keyMap:     make(map[string]pkcs11.ObjectHandle),
		config:     kmsConfig.PKCS11,
		maxRetries: maxRetries,
		timeout:    timeout,
	}

	// Discover key objects.
	if err := provider.findKeys(); err != nil {
		provider.Close()
		return nil, err
	}

	pkcs11Logger.Infow("PKCS#11 provider initialized",
		"module", kmsConfig.PKCS11.ModulePath,
		"slot", kmsConfig.PKCS11.SlotID,
		"keys", len(provider.keyMap),
	)

	return provider, nil
}

// findKeys enumerates key objects in the session and populates keyMap.
func (p *PKCS11Provider) findKeys() error {
	template := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_SECRET_KEY),
	}

	if err := p.ctx.FindObjectsInit(p.session, template); err != nil {
		return &KMSError{
			Provider: "pkcs11",
			Op:       "find-keys-init",
			KeyID:    "",
			Err:      fmt.Errorf("%w: %v", ErrPKCS11Session, err),
		}
	}

	objs, _, err := p.ctx.FindObjects(p.session, 100)
	if err != nil {
		p.ctx.FindObjectsFinal(p.session)
		return &KMSError{
			Provider: "pkcs11",
			Op:       "find-keys",
			KeyID:    "",
			Err:      fmt.Errorf("%w: %v", ErrPKCS11Session, err),
		}
	}

	if err := p.ctx.FindObjectsFinal(p.session); err != nil {
		return &KMSError{
			Provider: "pkcs11",
			Op:       "find-keys-final",
			KeyID:    "",
			Err:      fmt.Errorf("%w: %v", ErrPKCS11Session, err),
		}
	}

	// Also search for private keys (for signing).
	templatePriv := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_PRIVATE_KEY),
	}

	if err := p.ctx.FindObjectsInit(p.session, templatePriv); err == nil {
		privObjs, _, err := p.ctx.FindObjects(p.session, 100)
		if err == nil {
			objs = append(objs, privObjs...)
		}
		p.ctx.FindObjectsFinal(p.session)
	}

	for _, obj := range objs {
		attrs, err := p.ctx.GetAttributeValue(p.session, obj, []*pkcs11.Attribute{
			pkcs11.NewAttribute(pkcs11.CKA_LABEL, nil),
		})
		if err != nil {
			continue
		}
		for _, a := range attrs {
			if a.Type == pkcs11.CKA_LABEL && len(a.Value) > 0 {
				label := string(a.Value)
				p.keyMap[label] = obj
			}
		}
	}

	return nil
}

// Sign generates a signature using PKCS#11 ECDSA.
func (p *PKCS11Provider) Sign(ctx context.Context, keyID string, payload []byte) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	handle, err := p.resolveKey(keyID)
	if err != nil {
		return nil, err
	}

	var result []byte
	retryErr := withRetry(ctx, p.maxRetries, func() error {
		mechanism := []*pkcs11.Mechanism{pkcs11.NewMechanism(pkcs11.CKM_ECDSA, nil)}
		if err := p.ctx.SignInit(p.session, mechanism, handle); err != nil {
			return &KMSError{
				Provider: "pkcs11",
				Op:       "sign-init",
				KeyID:    keyID,
				Err:      fmt.Errorf("%w: %v", ErrPKCS11Session, err),
			}
		}

		sig, err := p.ctx.Sign(p.session, payload)
		if err != nil {
			return &KMSError{
				Provider: "pkcs11",
				Op:       "sign",
				KeyID:    keyID,
				Err:      fmt.Errorf("%w: %v", ErrPKCS11Session, err),
			}
		}
		result = sig
		return nil
	})
	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

// Encrypt encrypts plaintext using PKCS#11 AES-GCM.
// The IV is randomly generated and prepended to the ciphertext.
func (p *PKCS11Provider) Encrypt(ctx context.Context, keyID string, plaintext []byte) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	handle, err := p.resolveKey(keyID)
	if err != nil {
		return nil, err
	}

	// Generate random IV.
	iv := make([]byte, pkcs11GCMIVSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("generate IV: %w", err)
	}

	var result []byte
	retryErr := withRetry(ctx, p.maxRetries, func() error {
		params := pkcs11.NewGCMParams(iv, nil, pkcs11GCMTagBits)
		mechanism := []*pkcs11.Mechanism{pkcs11.NewMechanism(pkcs11.CKM_AES_GCM, params)}
		if err := p.ctx.EncryptInit(p.session, mechanism, handle); err != nil {
			return &KMSError{
				Provider: "pkcs11",
				Op:       "encrypt-init",
				KeyID:    keyID,
				Err:      fmt.Errorf("%w: %v", ErrPKCS11Session, err),
			}
		}

		ciphertext, err := p.ctx.Encrypt(p.session, plaintext)
		if err != nil {
			return &KMSError{
				Provider: "pkcs11",
				Op:       "encrypt",
				KeyID:    keyID,
				Err:      fmt.Errorf("%w: %v", ErrPKCS11Session, err),
			}
		}
		// Prepend IV to ciphertext.
		result = make([]byte, pkcs11GCMIVSize+len(ciphertext))
		copy(result, iv)
		copy(result[pkcs11GCMIVSize:], ciphertext)
		return nil
	})
	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

// Decrypt decrypts ciphertext using PKCS#11 AES-GCM.
// The IV is extracted from the first 12 bytes of the ciphertext.
func (p *PKCS11Provider) Decrypt(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < pkcs11GCMIVSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	handle, err := p.resolveKey(keyID)
	if err != nil {
		return nil, err
	}

	iv := ciphertext[:pkcs11GCMIVSize]
	encData := ciphertext[pkcs11GCMIVSize:]

	var result []byte
	retryErr := withRetry(ctx, p.maxRetries, func() error {
		params := pkcs11.NewGCMParams(iv, nil, pkcs11GCMTagBits)
		mechanism := []*pkcs11.Mechanism{pkcs11.NewMechanism(pkcs11.CKM_AES_GCM, params)}
		if err := p.ctx.DecryptInit(p.session, mechanism, handle); err != nil {
			return &KMSError{
				Provider: "pkcs11",
				Op:       "decrypt-init",
				KeyID:    keyID,
				Err:      fmt.Errorf("%w: %v", ErrPKCS11Session, err),
			}
		}

		plaintext, err := p.ctx.Decrypt(p.session, encData)
		if err != nil {
			return &KMSError{
				Provider: "pkcs11",
				Op:       "decrypt",
				KeyID:    keyID,
				Err:      fmt.Errorf("%w: %v", ErrPKCS11Session, err),
			}
		}
		result = plaintext
		return nil
	})
	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

// resolveKey looks up the key handle by label, falling back to the default key label.
func (p *PKCS11Provider) resolveKey(keyID string) (pkcs11.ObjectHandle, error) {
	resolved := keyID
	if resolved == "local" || resolved == "default" || resolved == "" {
		resolved = p.config.KeyLabel
	}

	handle, ok := p.keyMap[resolved]
	if !ok {
		return 0, &KMSError{
			Provider: "pkcs11",
			Op:       "resolve-key",
			KeyID:    resolved,
			Err:      ErrKMSInvalidKey,
		}
	}
	return handle, nil
}

// Close releases PKCS#11 resources.
func (p *PKCS11Provider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ctx == nil {
		return nil
	}

	p.ctx.Logout(p.session)
	p.ctx.CloseSession(p.session)
	p.ctx.Finalize()
	p.ctx.Destroy()
	p.ctx = nil
	return nil
}
