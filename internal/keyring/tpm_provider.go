//go:build linux

package keyring

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpm2/transport"
)

const (
	// tpmDevicePath is the TPM2 resource manager device on Linux.
	tpmDevicePath = "/dev/tpmrm0"
	// tpmSealedDir is the directory under ~/.lango/ for sealed blobs.
	tpmSealedDir = "tpm"
)

// TPMProvider stores secrets as TPM2-sealed blobs on disk.
// Only the same TPM chip can unseal the data, providing hardware-bound protection.
type TPMProvider struct {
	sealedDir string
}

var _ Provider   = (*TPMProvider)(nil)
var _ KeyChecker = (*TPMProvider)(nil)

// NewTPMProvider creates a new TPMProvider.
// Returns ErrTPMNotAvailable if the TPM2 device is not accessible.
func NewTPMProvider() (*TPMProvider, error) {
	t, err := transport.OpenTPM(tpmDevicePath)
	if err != nil {
		return nil, ErrTPMNotAvailable
	}
	t.Close()

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home directory: %w", err)
	}

	dir := filepath.Join(home, ".lango", tpmSealedDir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create sealed directory: %w", err)
	}

	return &TPMProvider{sealedDir: dir}, nil
}

// Get retrieves and unseals a secret from the TPM-sealed blob.
func (p *TPMProvider) Get(service, key string) (string, error) {
	blobPath := p.blobPath(service, key)
	blob, err := os.ReadFile(blobPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("read sealed blob: %w", err)
	}

	data, err := p.unseal(blob)
	if err != nil {
		return "", fmt.Errorf("tpm unseal: %w", err)
	}
	return string(data), nil
}

// Set seals a secret with the TPM and writes the sealed blob to disk.
func (p *TPMProvider) Set(service, key, value string) error {
	blob, err := p.seal([]byte(value))
	if err != nil {
		return fmt.Errorf("tpm seal: %w", err)
	}

	blobPath := p.blobPath(service, key)
	if err := os.WriteFile(blobPath, blob, 0600); err != nil {
		return fmt.Errorf("write sealed blob: %w", err)
	}
	return nil
}

// HasKey checks if a sealed blob file exists for the given key without unsealing.
func (p *TPMProvider) HasKey(service, key string) bool {
	_, err := os.Stat(p.blobPath(service, key))
	return err == nil
}

// Delete removes the sealed blob file.
func (p *TPMProvider) Delete(service, key string) error {
	blobPath := p.blobPath(service, key)
	if err := os.Remove(blobPath); err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return fmt.Errorf("remove sealed blob: %w", err)
	}
	return nil
}

// blobPath returns the file path for a sealed blob.
func (p *TPMProvider) blobPath(service, key string) string {
	return filepath.Join(p.sealedDir, service+"_"+key+".sealed")
}

// srkTemplate returns the Storage Root Key template (ECC P256).
func srkTemplate() tpm2.TPMTPublic {
	return tpm2.TPMTPublic{
		Type:    tpm2.TPMAlgECC,
		NameAlg: tpm2.TPMAlgSHA256,
		ObjectAttributes: tpm2.TPMAObject{
			FixedTPM:            true,
			FixedParent:         true,
			SensitiveDataOrigin: true,
			UserWithAuth:        true,
			NoDA:                true,
			Restricted:          true,
			Decrypt:             true,
		},
		Parameters: tpm2.NewTPMUPublicParms(
			tpm2.TPMAlgECC,
			&tpm2.TPMSECCParms{
				Symmetric: tpm2.TPMTSymDef{
					Algorithm: tpm2.TPMAlgAES,
					KeyBits:   tpm2.NewTPMUSymKeyBits(tpm2.TPMAlgAES, tpm2.TPMKeyBits(128)),
					Mode:      tpm2.NewTPMUSymMode(tpm2.TPMAlgAES, tpm2.TPMAlgCFB),
				},
				CurveID: tpm2.TPMECCNistP256,
			},
		),
	}
}

// sealedObjectTemplate returns the template for a sealed data object.
func sealedObjectTemplate() tpm2.TPMTPublic {
	return tpm2.TPMTPublic{
		Type:    tpm2.TPMAlgKeyedHash,
		NameAlg: tpm2.TPMAlgSHA256,
		ObjectAttributes: tpm2.TPMAObject{
			FixedTPM:     true,
			FixedParent:  true,
			UserWithAuth: true,
			NoDA:         true,
		},
	}
}

// createPrimaryKey creates an SRK under the owner hierarchy.
func createPrimaryKey(t transport.TPM) (*tpm2.CreatePrimaryResponse, error) {
	cmd := tpm2.CreatePrimary{
		PrimaryHandle: tpm2.TPMRHOwner,
		InPublic:      tpm2.New2B(srkTemplate()),
	}
	return cmd.Execute(t)
}

// seal encrypts data under the TPM's SRK.
// Returns a blob containing the marshaled public and private parts.
func (p *TPMProvider) seal(data []byte) ([]byte, error) {
	t, err := transport.OpenTPM(tpmDevicePath)
	if err != nil {
		return nil, fmt.Errorf("open tpm: %w", err)
	}
	defer t.Close()

	primary, err := createPrimaryKey(t)
	if err != nil {
		return nil, fmt.Errorf("create primary key: %w", err)
	}
	defer func() {
		flush := tpm2.FlushContext{FlushHandle: primary.ObjectHandle}
		flush.Execute(t)
	}()

	createCmd := tpm2.Create{
		ParentHandle: tpm2.NamedHandle{
			Handle: primary.ObjectHandle,
			Name:   primary.Name,
		},
		InSensitive: tpm2.TPM2BSensitiveCreate{
			Sensitive: &tpm2.TPMSSensitiveCreate{
				Data: tpm2.NewTPMUSensitiveCreate(
					&tpm2.TPM2BSensitiveData{Buffer: data},
				),
			},
		},
		InPublic: tpm2.New2B(sealedObjectTemplate()),
	}

	createRsp, err := createCmd.Execute(t)
	if err != nil {
		return nil, fmt.Errorf("create sealed object: %w", err)
	}

	return marshalSealedBlob(createRsp.OutPublic, createRsp.OutPrivate)
}

// unseal decrypts a sealed blob using the TPM's SRK.
func (p *TPMProvider) unseal(blob []byte) ([]byte, error) {
	pub, priv, err := unmarshalSealedBlob(blob)
	if err != nil {
		return nil, fmt.Errorf("unmarshal blob: %w", err)
	}

	t, err := transport.OpenTPM(tpmDevicePath)
	if err != nil {
		return nil, fmt.Errorf("open tpm: %w", err)
	}
	defer t.Close()

	primary, err := createPrimaryKey(t)
	if err != nil {
		return nil, fmt.Errorf("create primary key: %w", err)
	}
	defer func() {
		flush := tpm2.FlushContext{FlushHandle: primary.ObjectHandle}
		flush.Execute(t)
	}()

	loadCmd := tpm2.Load{
		ParentHandle: tpm2.NamedHandle{
			Handle: primary.ObjectHandle,
			Name:   primary.Name,
		},
		InPublic:  pub,
		InPrivate: priv,
	}
	loadRsp, err := loadCmd.Execute(t)
	if err != nil {
		return nil, fmt.Errorf("load sealed object: %w", err)
	}
	defer func() {
		flush := tpm2.FlushContext{FlushHandle: loadRsp.ObjectHandle}
		flush.Execute(t)
	}()

	unsealCmd := tpm2.Unseal{
		ItemHandle: tpm2.NamedHandle{
			Handle: loadRsp.ObjectHandle,
			Name:   loadRsp.Name,
		},
	}
	unsealRsp, err := unsealCmd.Execute(t)
	if err != nil {
		return nil, fmt.Errorf("unseal: %w", err)
	}

	return unsealRsp.OutData.Buffer, nil
}

// marshalSealedBlob encodes TPM public and private parts into a single byte slice.
// Format: [4-byte pubLen][pubBytes][4-byte privLen][privBytes]
func marshalSealedBlob(pub tpm2.TPM2BPublic, priv tpm2.TPM2BPrivate) ([]byte, error) {
	pubBytes, err := tpm2.Marshal(pub)
	if err != nil {
		return nil, fmt.Errorf("marshal public: %w", err)
	}
	privBytes, err := tpm2.Marshal(priv)
	if err != nil {
		return nil, fmt.Errorf("marshal private: %w", err)
	}

	buf := make([]byte, 4+len(pubBytes)+4+len(privBytes))
	binary.BigEndian.PutUint32(buf[0:4], uint32(len(pubBytes)))
	copy(buf[4:], pubBytes)
	offset := 4 + len(pubBytes)
	binary.BigEndian.PutUint32(buf[offset:offset+4], uint32(len(privBytes)))
	copy(buf[offset+4:], privBytes)

	return buf, nil
}

// unmarshalSealedBlob decodes a sealed blob back into TPM public and private parts.
func unmarshalSealedBlob(blob []byte) (tpm2.TPM2BPublic, tpm2.TPM2BPrivate, error) {
	var pub tpm2.TPM2BPublic
	var priv tpm2.TPM2BPrivate

	if len(blob) < 8 {
		return pub, priv, fmt.Errorf("sealed blob too short")
	}

	pubLen := binary.BigEndian.Uint32(blob[0:4])
	if uint32(len(blob)) < 4+pubLen+4 {
		return pub, priv, fmt.Errorf("sealed blob truncated at public")
	}

	pubBytes := blob[4 : 4+pubLen]
	if _, err := tpm2.Unmarshal(pubBytes, &pub); err != nil {
		return pub, priv, fmt.Errorf("unmarshal public: %w", err)
	}

	offset := 4 + pubLen
	privLen := binary.BigEndian.Uint32(blob[offset : offset+4])
	if uint32(len(blob)) < offset+4+privLen {
		return pub, priv, fmt.Errorf("sealed blob truncated at private")
	}

	privBytes := blob[offset+4 : offset+4+privLen]
	if _, err := tpm2.Unmarshal(privBytes, &priv); err != nil {
		return pub, priv, fmt.Errorf("unmarshal private: %w", err)
	}

	return pub, priv, nil
}
