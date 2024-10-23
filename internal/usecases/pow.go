package usecases

import (
	"faraway/internal/domain"
	"faraway/pkg/pow/argon2"
	"faraway/pkg/pow/hashcash"
	"fmt"
	"log"
)

// PowUsecase defines the interface for Proof of Work usecase.
type PowUsecase interface {
	GenerateCPUBoundChallenge() (*domain.ProofOfWork, error)
	GenerateMemoryBoundChallenge() (*domain.ProofOfWork, error)

	ValidateCPUBoundSolution(challenge, nonce []byte) bool
	ValidateMemoryBoundSolution(challenge, nonce []byte) (bool, error)
}

type powUsecaseImpl struct {
	hashcash *hashcash.HashCash
	argon2   *argon2.Argon2
}

// NewPowUsecase initializes the powUsecaseImpl with the specified difficulty.
func NewPowUsecase(difficulty uint64) (PowUsecase, error) {
	hashcash, err := hashcash.NewHashCash(difficulty)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize hashcash: %w", err)
	}
	argon2, err := argon2.NewArgon2(difficulty)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize argon2: %w", err)
	}
	return &powUsecaseImpl{
		hashcash: hashcash,
		argon2:   argon2,
	}, nil
}

// GenerateCPUBoundChallenge creates a new challenge using the hashcash package.
func (p *powUsecaseImpl) GenerateCPUBoundChallenge() (*domain.ProofOfWork, error) {
	challenge, err := p.hashcash.GenerateChallenge()
	if err != nil {
		return nil, fmt.Errorf("failed to generate challenge: %w", err)
	}
	return &domain.ProofOfWork{
		Challenge:  challenge,
		Difficulty: p.hashcash.GetDifficulty(),
	}, nil
}

// ValidateCPUBoundSolution checks if the provided solution (nonce) is valid for the given challenge.
// It returns false if the solution is invalid or if any error occurs during verification.
func (p *powUsecaseImpl) ValidateCPUBoundSolution(challenge, nonce []byte) bool {
	if len(challenge) == 0 || len(nonce) == 0 {
		log.Printf("Invalid input: challenge length=%d, nonce length=%d", len(challenge), len(nonce))
		return false
	}

	return p.hashcash.Verify(challenge, nonce)
}

// GenerateMemoryBoundChallenge creates a new challenge using the argon2 package.
func (p *powUsecaseImpl) GenerateMemoryBoundChallenge() (*domain.ProofOfWork, error) {
	challenge, err := p.argon2.GenerateChallenge()
	if err != nil {
		return nil, fmt.Errorf("failed to generate challenge: %w", err)
	}
	return &domain.ProofOfWork{
		Challenge:  challenge,
		Difficulty: p.argon2.GetDifficulty(),
	}, nil
}

// ValidateMemoryBoundSolution checks if the provided solution (nonce) is valid for the given challenge.
// It returns false if the solution is invalid or if any error occurs during verification.
func (p *powUsecaseImpl) ValidateMemoryBoundSolution(challenge, nonce []byte) (bool, error) {
	if len(challenge) == 0 || len(nonce) == 0 {
		log.Printf("Invalid input: challenge length=%d, nonce length=%d", len(challenge), len(nonce))
		return false, nil
	}

	isVerified, err := p.argon2.Verify(challenge, string(nonce))
	if err != nil {
		return false, fmt.Errorf("failed to verify argon2 solution: %w", err)
	}

	return isVerified, nil
}
