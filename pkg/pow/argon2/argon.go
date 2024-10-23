package argon2

/*
Key Concepts of Argon2:

Challenge-Response Mechanism:
Argon2 is not typically used in challenge-response mechanisms but is a key derivation function (KDF)
designed for secure password hashing and proof-of-work applications.
In a password hashing context, the challenge is typically a password or data (e.g., some fixed challenge bytes),
and the response is the resulting hashed value derived using Argon2. This derived hash is compared against the
stored hash for verification.

Memory-Hard Function:
Argon2 is designed to be "memory-hard," meaning that it requires a significant amount of memory to compute,
making it resistant to specialized hardware attacks (e.g., FPGA or ASIC).
The difficulty in deriving a hash is controlled by specifying the amount of memory, time cost, and degree of parallelism.

Difficulty:
Argon2 provides adjustable parameters for difficulty:
Memory cost: The amount of memory Argon2 uses during hashing.
Time cost: The number of iterations or rounds performed by Argon2.
Parallelism (threads): The number of computational threads used for parallel processing.
These parameters allow customization of how computationally intensive it is to derive a hash, balancing between CPU
and memory usage.
The difficulty increases with larger memory and time costs, making it harder for attackers to compute hashes efficiently,
especially on low-memory devices.

Key Derivation and Password Hashing:
Argon2 is widely used for securely hashing passwords by incorporating both a password (or key material) and a salt
(a unique random value per hash) into the hash computation.
The use of a unique salt ensures that identical passwords result in different hashes, preventing precomputed attacks
(e.g., rainbow tables).
Proof-of-Work (PoW):

While not originally designed as a PoW algorithm, Argon2 can be adapted for proof-of-work systems where the solution
to a challenge requires intensive computation (e.g., cryptographic puzzles that require solving under memory and time constraints).
The client presents a "solution" that is verifiable by anyone with minimal effort, but finding the solution requires
a significant computational effort and memory resources.

Security Against Hardware Attacks:
Argon2 is designed to resist GPU and ASIC optimizations that can easily exploit traditional password hashing algorithms
like SHA-256. Its memory-hard design ensures that attackers need both significant CPU power and large amounts of memory
to perform brute-force attacks.
This makes it ideal for applications where password security is crucial, as attackers are forced to spend considerably more
resources to break the hash.

Pros and Cons of Argon2:

Pros:
Memory-hardness: Argon2 is highly resistant to brute-force attacks on specialized hardware like GPUs and ASICs.
Customizable difficulty: Users can adjust memory, time, and parallelism settings to tailor Argon2 to their security needs.
Security: Argon2 is highly regarded for its security, winning the Password Hashing Competition (PHC) and becoming a
standard for password hashing.
Efficiency: Despite its security, Argon2 is still reasonably fast for legitimate users, especially when using parallelism.

Cons:
Resource-intensive: Argon2's memory-hard design may not be suitable for environments with limited computational power,
such as embedded systems or mobile devices.
Complexity: Setting up the correct parameters for memory, time, and parallelism requires some understanding of the system's
security needs and resource availability.
Not as widely supported: While growing in popularity, Argon2 is not yet as universally supported across all systems and
languages as simpler algorithms like bcrypt or PBKDF2.
*/

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

const (
	// Parameters for Argon2
	argon2Time        = 1                // Number of iterations (time cost)
	argon2Memory      = 64 * 1024        // Memory usage (64MB)
	argon2Threads     = 4                // Number of threads to use
	argon2KeyLength   = 32               // Length of the generated key
	argon2SaltLength  = 16               // Length of the salt
	argon2TokenLength = 16               // Length of the random challenge token
	argon2MaxTime     = 10 * time.Second // Maximum time allowed to compute the solution
)

var (
	ErrDifficultyRange = errors.New("difficulty out of acceptable range")
	ErrGenerateRandom  = errors.New("failed to generate random challenge")
	ErrArgon2Timeout   = errors.New("argon2 solution computation timed out")
	ErrInvalidSolution = errors.New("invalid argon2 solution")
	ErrInvalidFormat   = errors.New("invalid solution format")
)

// Argon2 encapsulates the Argon2-based proof-of-work mechanism.
type Argon2 struct {
	difficultyLevel uint64
}

// Solution represents an Argon2 proof-of-work solution
type Solution struct {
	Hash string
	Salt string
}

// NewArgon2 initializes a new Argon2 proof-of-work with a specified difficulty.
func NewArgon2(difficulty uint64) (*Argon2, error) {
	if difficulty < 1 || difficulty > 10 {
		return nil, fmt.Errorf("%w: difficulty must be between 1 and 10", ErrDifficultyRange)
	}
	return &Argon2{
		difficultyLevel: difficulty,
	}, nil
}

// GenerateChallenge creates a new cryptographically secure random challenge token.
func (pow *Argon2) GenerateChallenge() ([]byte, error) {
	bytes := make([]byte, argon2TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGenerateRandom, err)
	}
	return bytes, nil
}

// FindSolution computes a valid Argon2 solution for the challenge.
// Returns a solution string in the format "hash$salt" for verification.
func (pow *Argon2) FindSolution(challenge []byte) (string, error) {
	// Generate a random salt
	salt := make([]byte, argon2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("%w: %v", ErrGenerateRandom, err)
	}

	// Start a timeout for Argon2 computation
	done := make(chan string, 1)

	go func() {
		// Derive key using Argon2 with memory constraints
		key := argon2.IDKey(challenge, salt, uint32(pow.difficultyLevel), argon2Memory, argon2Threads, argon2KeyLength)

		// Encode both the key and salt in base64
		hashStr := base64.StdEncoding.EncodeToString(key)
		saltStr := base64.StdEncoding.EncodeToString(salt)

		// Combine hash and salt with a separator
		solution := fmt.Sprintf("%s$%s", hashStr, saltStr)
		done <- solution
	}()

	select {
	case solution := <-done:
		return solution, nil
	case <-time.After(argon2MaxTime):
		return "", ErrArgon2Timeout
	}
}

// Verify checks if the provided solution satisfies the challenge.
// Solution should be in the format "hash$salt" where both are base64 encoded.
func (pow *Argon2) Verify(challenge []byte, solutionStr string) (bool, error) {
	// Split the solution string to get hash and salt
	parts := strings.Split(solutionStr, "$")
	if len(parts) != 2 {
		return false, ErrInvalidFormat
	}

	// Decode the hash and salt from base64
	hash, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return false, fmt.Errorf("invalid hash encoding: %v", err)
	}

	salt, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, fmt.Errorf("invalid salt encoding: %v", err)
	}

	// Derive the key using the same parameters and salt
	computedKey := argon2.IDKey(challenge, salt, uint32(pow.difficultyLevel), argon2Memory, argon2Threads, argon2KeyLength)

	// Debugging output
	fmt.Printf("Challenge: %s\n", base64.StdEncoding.EncodeToString(challenge))
	fmt.Printf("Solution: %s\n", solutionStr)
	fmt.Printf("Computed Key: %s\n", base64.StdEncoding.EncodeToString(computedKey))
	fmt.Printf("Provided Hash: %s\n", base64.StdEncoding.EncodeToString(hash))
	fmt.Printf("Salt: %s\n", base64.StdEncoding.EncodeToString(salt))

	// Compare the computed key with the provided hash
	if len(computedKey) != len(hash) || subtle.ConstantTimeCompare(computedKey, hash) != 1 {
		return false, nil
	}

	return true, nil
}

// GetDifficulty returns the current difficulty level
func (pow *Argon2) GetDifficulty() uint64 {
	return pow.difficultyLevel
}
