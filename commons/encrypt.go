package commons

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jxskiss/base62"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
	"golang.org/x/xerrors"
)

const (
	PgpEncryptedFileExt string = ".pgp.enc"
)

// EncryptionMode determines encryption mode
type EncryptionMode string

const (
	// EncryptionModeWinSCP is for WinSCP
	EncryptionModeWinSCP EncryptionMode = "WINSCP"
	// EncryptionModePGP is for PGP
	EncryptionModePGP EncryptionMode = "PGP"
	// EncryptionModeUnknown is for unknown mode
	EncryptionModeUnknown EncryptionMode = ""
)

// GetEncryptionMode returns encryption mode
func GetEncryptionMode(mode string) EncryptionMode {
	switch strings.ToUpper(mode) {
	case string(EncryptionModeWinSCP):
		return EncryptionModeWinSCP
	case string(EncryptionModePGP), "GPG", "OPENPGP":
		return EncryptionModePGP
	default:
		return EncryptionModeUnknown
	}
}

type EncryptionManager struct {
	mode            EncryptionMode
	encryptFilename bool
	password        string
}

// NewEncryptionManager creates a new EncryptionManager
func NewEncryptionManager(mode EncryptionMode, encryptFilename bool, password string) *EncryptionManager {
	manager := &EncryptionManager{
		mode:            mode,
		encryptFilename: encryptFilename,
		password:        password,
	}

	return manager
}

func (manager *EncryptionManager) EncryptFilename(filename string) (string, error) {
	switch manager.mode {
	case EncryptionModeWinSCP:
		return manager.encryptFilenameWinSCP(filename)
	case EncryptionModePGP:
		return manager.encryptFilenamePGP(filename)
	default:
		// EncryptionModeWinSCP
		return manager.encryptFilenameWinSCP(filename)
	}
}

func (manager *EncryptionManager) DecryptFilename(filename string) (string, error) {
	switch manager.mode {
	case EncryptionModeWinSCP:
		return manager.decryptFilenameWinSCP(filename)
	case EncryptionModePGP:
		return manager.decryptFilenamePGP(filename)
	default:
		// EncryptionModeWinSCP
		return manager.decryptFilenameWinSCP(filename)
	}
}

// EncryptFile encrypts local source file and returns encrypted file path
func (manager *EncryptionManager) EncryptFile(source string, target string) error {
	switch manager.mode {
	case EncryptionModeWinSCP:
		return manager.encryptFileWinSCP(source, target)
	case EncryptionModePGP:
		return manager.encryptFilePGP(source, target)
	default:
		// EncryptionModeWinSCP
		return manager.encryptFileWinSCP(source, target)
	}
}

// DecryptFile decrypts local source file and returns decrypted file path
func (manager *EncryptionManager) DecryptFile(source string, target string) error {
	switch manager.mode {
	case EncryptionModeWinSCP:
		return manager.decryptFileWinSCP(source, target)
	case EncryptionModePGP:
		return manager.decryptFilePGP(source, target)
	default:
		// EncryptionModeWinSCP
		return manager.decryptFileWinSCP(source, target)
	}
}

func (manager *EncryptionManager) encryptFilenameWinSCP(filename string) (string, error) {
	//TODO: Implement this
	return filename, nil
}

func (manager *EncryptionManager) decryptFilenameWinSCP(filename string) (string, error) {
	//TODO: Implement this
	return filename, nil
}

func (manager *EncryptionManager) encryptFilenamePGP(filename string) (string, error) {
	if !manager.encryptFilename {
		return filename, nil
	}

	encryptedFilename, err := manager.encryptAES([]byte(filename))
	if err != nil {
		xerrors.Errorf("failed to encrypt filename: %w", err)
	}

	b64EncodedFilename := base62.EncodeToString(encryptedFilename)
	newFilename := fmt.Sprintf("%s%s", b64EncodedFilename, PgpEncryptedFileExt)

	return newFilename, nil
}

func (manager *EncryptionManager) decryptFilenamePGP(filename string) (string, error) {
	if !manager.encryptFilename {
		return filename, nil
	}

	// trim file ext
	filename = strings.TrimSuffix(filename, PgpEncryptedFileExt)

	encryptedFilename, err := base62.DecodeString(string(filename))
	if err != nil {
		xerrors.Errorf("failed to base62 decode filename: %w", err)
	}

	decryptedFilename, err := manager.decryptAES(encryptedFilename)
	if err != nil {
		xerrors.Errorf("failed to decrypt filename: %w", err)
	}

	return string(decryptedFilename), nil
}

func (manager *EncryptionManager) encryptFileWinSCP(source string, target string) error {
	return nil
}

func (manager *EncryptionManager) decryptFileWinSCP(source string, target string) error {
	return nil
}

func (manager *EncryptionManager) encryptFilePGP(source string, target string) error {
	sourceFileHandle, err := os.Open(source)
	if err != nil {
		return xerrors.Errorf("failed to open file %s: %w", source, err)
	}

	defer sourceFileHandle.Close()

	targetFileHandle, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return xerrors.Errorf("failed to create file %s: %w", target, err)
	}

	defer targetFileHandle.Close()

	encryptionConfig := &packet.Config{
		DefaultCipher: packet.CipherAES256,
	}

	writeHandle, err := openpgp.SymmetricallyEncrypt(targetFileHandle, []byte(manager.password), nil, encryptionConfig)
	if err != nil {
		return xerrors.Errorf("failed to create a encrypt writer for %s: %w", target, err)
	}

	defer writeHandle.Close()

	_, err = io.Copy(writeHandle, sourceFileHandle)
	if err != nil {
		return xerrors.Errorf("failed to encrypt data: %w", err)
	}

	return nil
}

func (manager *EncryptionManager) decryptFilePGP(source string, target string) error {
	sourceFileHandle, err := os.Open(source)
	if err != nil {
		return xerrors.Errorf("failed to open file %s: %w", source, err)
	}

	defer sourceFileHandle.Close()

	targetFileHandle, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return xerrors.Errorf("failed to create file %s: %w", target, err)
	}

	defer targetFileHandle.Close()

	encryptionConfig := &packet.Config{
		DefaultCipher: packet.CipherAES256,
	}

	failed := false
	prompt := func(keys []openpgp.Key, symmetric bool) ([]byte, error) {
		if failed {
			return nil, xerrors.New("decryption failed")
		}
		failed = true
		return []byte(manager.password), nil
	}

	messageDetail, err := openpgp.ReadMessage(sourceFileHandle, nil, prompt, encryptionConfig)
	if err != nil {
		return xerrors.Errorf("failed to decrypt for %s: %w", source, err)
	}

	_, err = io.Copy(targetFileHandle, messageDetail.UnverifiedBody)
	if err != nil {
		return xerrors.Errorf("failed to decrypt data: %w", err)
	}

	return nil
}

func (manager *EncryptionManager) padAesKey(key string) string {
	// AES key must be 16bytes len
	paddedKey := fmt.Sprintf("%s%s", key, aesPadding)
	return paddedKey[:16]
}

func (manager *EncryptionManager) padPkcs7(data []byte, blocksize int) []byte {
	n := blocksize - (len(data) % blocksize)
	pb := make([]byte, len(data)+n)
	copy(pb, data)
	copy(pb[len(data):], bytes.Repeat([]byte{byte(n)}, n))
	return pb
}

func (manager *EncryptionManager) decryptAES(data []byte) ([]byte, error) {
	key := manager.padAesKey(manager.password)
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, xerrors.Errorf("failed to create AES cipher: %w", err)
	}

	decrypter := cipher.NewCBCDecrypter(block, []byte(aesIV))
	contentLength := binary.LittleEndian.Uint32(data[:4])

	dest := make([]byte, len(data[4:]))
	decrypter.CryptBlocks(dest, data[4:])

	return dest[:contentLength], nil
}

func (manager *EncryptionManager) encryptAES(data []byte) ([]byte, error) {
	key := manager.padAesKey(manager.password)
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, xerrors.Errorf("failed to create AES cipher: %w", err)
	}

	encrypter := cipher.NewCBCEncrypter(block, []byte(aesIV))

	contentLength := uint32(len(data))
	padData := manager.padPkcs7(data, block.BlockSize())

	dest := make([]byte, len(padData)+4)

	// add size header
	binary.LittleEndian.PutUint32(dest, contentLength)
	encrypter.CryptBlocks(dest[4:], padData)

	return dest, nil
}
