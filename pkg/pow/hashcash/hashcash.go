package hashcash

/*
	Key Concepts of Hashcash:

	Challenge-Response Mechanism:
	The system generates a challenge (e.g., a random string or nonce).
	The client (sender) needs to solve this challenge by finding a value (typically called a nonce)
	such that the hash of the challenge combined with the nonce meets a difficulty target (usually a
	certain number of leading zeros in the hash).

	Hash Function:
	The hash function used is typically a cryptographic hash function like SHA-1 or SHA-256.
	The goal is to find a combination of the challenge and a nonce such that the resulting hash
	meets the required difficulty level, such as starting with a specific number of leading zeros.

	Difficulty:
	The difficulty is adjustable and determines how much computational effort is required to solve the challenge.
	Higher difficulty levels require more resources to find a valid solution.
	The difficulty is often specified by the number of leading zeros in the binary or hexadecimal representation of the hash.

	Proof-of-Work:
	The client proves they expended computational effort by presenting the nonce that, when combined with
	the challenge, produces a valid hash.
	Once a valid hash is found, it is sent back to the server (or verifier), which can easily verify whether
	the hash meets the difficulty requirement.

	Anti-Spam and DDOS Protection:
	Hashcash can be used to prevent spam in email systems by requiring senders to perform computational
	work before sending each email. Legitimate users can solve the PoW easily for a few emails, but spammers
	would need substantial computational resources to send bulk messages.
	It can also be used to mitigate Denial of Service (DDoS) attacks, as attackers would need to solve PoW
	puzzles for every connection request, slowing down the attack.

	Pros and Cons of Hashcash:

	Pros:
	Lightweight and easy to implement.
	Scalable by adjusting the difficulty.
	Provides an effective mechanism for rate-limiting abuse (like spam or DDoS).

	Cons:
	Requires computational resources, which may be costly for legitimate users in resource-constrained
	environments (e.g., mobile devices).
	Does not completely eliminate the possibility of attacks, but makes them more expensive.
*/

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
