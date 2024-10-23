package hashcash

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const (
	tokenLength   = 16
	maxDifficulty = 64 // Maximum possible difficulty (SHA-256 output length)
)

var (
	ErrDifficultyRange = errors.New("difficulty out of acceptable range")
	ErrGenerateRandom  = errors.New("failed to generate random challenge")
	ErrTimeout         = errors.New("solution computation timed out")
)

// ProofOfWork encapsulates a proof-of-work mechanism.
type ProofOfWork struct {
	difficultyLevel uint64
}

// NewProofOfWork initializes a ProofOfWork with a specified difficulty.
func NewProofOfWork(difficulty uint64) (*ProofOfWork, error) {
	if difficulty < 1 || difficulty > maxDifficulty {
		return nil, fmt.Errorf("%w: difficulty must be between 1 and %d", ErrDifficultyRange, maxDifficulty)
	}

	return &ProofOfWork{
		difficultyLevel: difficulty,
	}, nil
}

// GenerateChallenge creates a new challenge using cryptographically secure random numbers.
func (pow *ProofOfWork) GenerateChallenge() ([]byte, error) {
	bytes := make([]byte, tokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGenerateRandom, err)
	}
	return bytes, nil
}

// Verify checks if the provided solution satisfies the challenge.
func (pow *ProofOfWork) Verify(challengeBytes []byte, solutionBytes []byte) bool {
	hash := sha256.Sum256([]byte(string(challengeBytes) + string(solutionBytes)))
	hashStr := hex.EncodeToString(hash[:])
	return strings.HasPrefix(hashStr, strings.Repeat("0", int(pow.difficultyLevel)))
}

func (pow *ProofOfWork) GetDifficulty() uint64 {
	return pow.difficultyLevel
}

// FindSolution attempts to compute a valid solution for the challenge.
func (pow *ProofOfWork) FindSolution(challenge []byte, difficulty uint64) string {
	return computeSolution(challenge, difficulty)
}

// computeSolution iterates through possible nonces to find a valid solution for the challenge.
func computeSolution(challenge []byte, difficulty uint64) string {
	zerosPrefix := strings.Repeat("0", int(difficulty))
	var nonce int

	for {
		// Concatenate the challenge and the current nonce
		data := fmt.Sprintf("%s%d", challenge, nonce)

		// Compute the SHA-256 hash
		hash := sha256.Sum256([]byte(data))
		hashStr := hex.EncodeToString(hash[:])

		// Check if the hash has the required number of leading zeros
		if strings.HasPrefix(hashStr, zerosPrefix) {
			return fmt.Sprintf("%d", nonce)
		}

		nonce++
	}
}
