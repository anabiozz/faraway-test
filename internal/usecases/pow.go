package usecases

import (
	"faraway/internal/domain"
	"faraway/pkg/pow/hashcash"
	"fmt"
	"log"
)

// PowUsecase defines the interface for Proof of Work usecase.
type PowUsecase interface {
	GenerateChallenge() (*domain.ProofOfWork, error)
	ValidateSolution(challenge, nonce []byte) bool
}

type powUsecaseImpl struct {
	pow *hashcash.ProofOfWork
}

// NewPowUsecase initializes the powUsecaseImpl with the specified difficulty.
func NewPowUsecase(difficulty uint64) (PowUsecase, error) {
	pow, err := hashcash.NewProofOfWork(difficulty)
	if err != nil {
		return nil, err
	}
	return &powUsecaseImpl{
		pow: pow,
	}, nil
}

// GenerateChallenge creates a new challenge using the hashcash package.
func (p *powUsecaseImpl) GenerateChallenge() (*domain.ProofOfWork, error) {
	challenge, err := p.pow.GenerateChallenge()
	if err != nil {
		return nil, fmt.Errorf("failed to generate challenge: %w", err)
	}
	return &domain.ProofOfWork{
		Challenge:  challenge,
		Difficulty: p.pow.GetDifficulty(),
	}, nil
}

// ValidateSolution checks if the provided solution (nonce) is valid for the given challenge.
// It returns false if the solution is invalid or if any error occurs during verification.
func (p *powUsecaseImpl) ValidateSolution(challenge, nonce []byte) bool {
	if len(challenge) == 0 || len(nonce) == 0 {
		log.Printf("Invalid input: challenge length=%d, nonce length=%d", len(challenge), len(nonce))
		return false
	}

	return p.pow.Verify(challenge, nonce)
}
