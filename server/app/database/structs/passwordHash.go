package structs

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"golang.org/x/exp/rand"
	"strings"
)

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

const algoName = "argon2id"

type Params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

type PasswordHash struct {
	hash   []byte
	salt   []byte
	params Params
}

func (h *PasswordHash) String() string {
	b64Salt := base64.RawStdEncoding.EncodeToString(h.salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(h.hash)
	return fmt.Sprintf(
		"$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		algoName,
		argon2.Version,
		h.params.memory,
		h.params.iterations,
		h.params.parallelism,
		b64Salt,
		b64Hash,
	)
}

func (h *PasswordHash) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	params, salt, hash, err := decodeHash(v)
	if err != nil {
		return err
	}
	h.params = params
	h.salt = salt
	h.hash = hash
	return nil
}

func (h *PasswordHash) Compare(otherHash PasswordHash) bool {
	return subtle.ConstantTimeCompare(h.hash, otherHash.hash) == 1
}

func (h *PasswordHash) CompareHash(hash string) (bool, error) {
	password, err := LoadPasswordHash(hash)
	if err != nil {
		return false, err
	}
	return h.Compare(password), nil
}

func (h *PasswordHash) ComparePlainText(plainText string) (bool, error) {
	password, err := NewPasswordHash(plainText, h.params)
	if err != nil {
		return false, nil
	}
	return h.Compare(password), nil
}

func NewPasswordHash(plainText string, params Params) (PasswordHash, error) {
	salt, err := generateRandomBytes(int(params.saltLength))
	if err != nil {
		return PasswordHash{}, err
	}

	hash := argon2.IDKey(
		[]byte(plainText),
		salt,
		params.iterations,
		params.memory,
		params.parallelism,
		params.keyLength,
	)

	return PasswordHash{
		hash:   hash,
		salt:   salt,
		params: params,
	}, nil
}

func LoadPasswordHash(hashedValue string) (PasswordHash, error) {
	params, salt, hash, err := decodeHash(hashedValue)
	if err != nil {
		return PasswordHash{}, err
	}
	return PasswordHash{
		hash:   hash,
		salt:   salt,
		params: params,
	}, nil
}

func NewPasswordParams(
	memory uint32,
	iterations uint32,
	parallelism uint8,
	saltLength uint32,
	keyLength uint32,
) Params {
	return Params{
		memory:      memory,
		iterations:  iterations,
		parallelism: parallelism,
		saltLength:  saltLength,
		keyLength:   keyLength,
	}
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func decodeHash(encodedHash string) (p Params, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return Params{}, nil, nil, ErrInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return Params{}, nil, nil, err
	}
	if version != argon2.Version {
		return Params{}, nil, nil, ErrIncompatibleVersion
	}

	p = Params{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return Params{}, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return Params{}, nil, nil, err
	}
	p.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return Params{}, nil, nil, err
	}
	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}
