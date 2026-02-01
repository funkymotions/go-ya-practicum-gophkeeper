package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/proto/storage"
	"golang.org/x/crypto/scrypt"
)

type ExtractedKey struct {
	Key  []byte
	Salt []byte
}

type ScryptParams struct {
	N         int
	R         int
	P         int
	KeyLength int
	Algo      string
}

type ScryptProfile string

const (
	ProfileLow    ScryptProfile = "PROFILE_V1"
	ProfileMedium ScryptProfile = "PROFILE_V2"
	ProfileHigh   ScryptProfile = "PROFILE_V3"
)

var ScryptProfiles = map[ScryptProfile]ScryptParams{
	"PROFILE_V1": {N: 1 << 14, R: 8, P: 1, KeyLength: 32, Algo: "scrypt"},
	"PROFILE_V2": {N: 1 << 15, R: 8, P: 1, KeyLength: 32, Algo: "scrypt"},
	"PROFILE_V3": {N: 1 << 16, R: 8, P: 1, KeyLength: 32, Algo: "scrypt"},
}

func ProfileToProto(profile ScryptProfile) storage.EncProfile {
	switch profile {
	case ProfileLow:
		return storage.EncProfile_PROFILE_V1
	case ProfileMedium:
		return storage.EncProfile_PROFILE_V2
	case ProfileHigh:
		return storage.EncProfile_PROFILE_V3
	default:
		return storage.EncProfile_PROFILE_V1
	}
}

func ExtractKeyFromPassword(password string, profile ScryptProfile) (*ExtractedKey, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	scryptProfile, ok := ScryptProfiles[profile]
	if !ok {
		scryptProfile = ScryptProfiles["PROFILE_V1"]
	}

	key, err := scrypt.Key([]byte(password), salt, scryptProfile.N, scryptProfile.R, scryptProfile.P, scryptProfile.KeyLength)
	if err != nil {
		return nil, err
	}

	return &ExtractedKey{Key: key, Salt: salt}, nil
}

func EncryptWithPassword(
	passsword []byte,
	payload []byte,
	key *ExtractedKey,
) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key.Key)
	if err != nil {
		return nil, nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}

	ciphertext := aead.Seal(nil, nonce, payload, nil)

	return ciphertext, nonce, nil
}

func DecryptWithPassword(
	ciphertext []byte,
	nonce []byte,
	password []byte,
	salt []byte,
	profile ScryptProfile,
) ([]byte, error) {
	p := ScryptProfiles[profile]
	key, err := scrypt.Key(password, salt, p.N, p.R, p.P, p.KeyLength)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	pt, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return pt, nil
}
