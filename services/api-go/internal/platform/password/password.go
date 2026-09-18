package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonVersion = argon2.Version
	memoryKiB    = 19 * 1024
	iterations   = 2
	parallelism  = 1
	saltLength   = 16
	keyLength    = 32
)

var (
	ErrEmptyPassword = errors.New("password must not be empty")
	ErrInvalidHash   = errors.New("invalid password hash")
)

type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

var defaultParams = params{
	memory:      memoryKiB,
	iterations:  iterations,
	parallelism: parallelism,
	saltLength:  saltLength,
	keyLength:   keyLength,
}

func Hash(plain string) (string, error) {
	if plain == "" {
		return "", ErrEmptyPassword
	}

	salt, err := randomBytes(defaultParams.saltLength)
	if err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	hash := argon2.IDKey([]byte(plain), salt, defaultParams.iterations, defaultParams.memory, defaultParams.parallelism, defaultParams.keyLength)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argonVersion,
		defaultParams.memory,
		defaultParams.iterations,
		defaultParams.parallelism,
		encodedSalt,
		encodedHash,
	), nil
}

func Verify(plain string, encodedHash string) (bool, error) {
	if plain == "" {
		return false, ErrEmptyPassword
	}

	params, salt, expectedHash, err := parse(encodedHash)
	if err != nil {
		return false, err
	}

	actualHash := argon2.IDKey([]byte(plain), salt, params.iterations, params.memory, params.parallelism, params.keyLength)
	if subtle.ConstantTimeCompare(actualHash, expectedHash) == 1 {
		return true, nil
	}

	return false, nil
}

func randomBytes(length uint32) ([]byte, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}

	return buf, nil
}

func parse(encodedHash string) (params, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return params{}, nil, nil, ErrInvalidHash
	}

	version, err := parseVersion(parts[2])
	if err != nil || version != argonVersion {
		return params{}, nil, nil, ErrInvalidHash
	}

	params, err := parseParams(parts[3])
	if err != nil {
		return params, nil, nil, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return params, nil, nil, ErrInvalidHash
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return params, nil, nil, ErrInvalidHash
	}

	if len(salt) == 0 || len(hash) == 0 {
		return params, nil, nil, ErrInvalidHash
	}

	params.saltLength = uint32(len(salt))
	params.keyLength = uint32(len(hash))

	return params, salt, hash, nil
}

func parseVersion(value string) (int, error) {
	versionText, ok := strings.CutPrefix(value, "v=")
	if !ok {
		return 0, ErrInvalidHash
	}

	return strconv.Atoi(versionText)
}

func parseParams(value string) (params, error) {
	items := strings.Split(value, ",")
	if len(items) != 3 {
		return params{}, ErrInvalidHash
	}

	parsed := make(map[string]uint64, len(items))
	for _, item := range items {
		key, rawValue, ok := strings.Cut(item, "=")
		if !ok {
			return params{}, ErrInvalidHash
		}

		number, err := strconv.ParseUint(rawValue, 10, 32)
		if err != nil || number == 0 {
			return params{}, ErrInvalidHash
		}

		parsed[key] = number
	}

	memory, ok := parsed["m"]
	if !ok {
		return params{}, ErrInvalidHash
	}

	iterations, ok := parsed["t"]
	if !ok {
		return params{}, ErrInvalidHash
	}

	parallelismValue, ok := parsed["p"]
	if !ok || parallelismValue > 255 {
		return params{}, ErrInvalidHash
	}

	return params{
		memory:      uint32(memory),
		iterations:  uint32(iterations),
		parallelism: uint8(parallelismValue),
	}, nil
}
