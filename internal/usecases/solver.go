package usecases

import (
	"faraway/pkg/pow/hashcash"
)

// SolverUsecase
type SolverUsecase interface {
	FindSolution(challenge []byte, difficulty uint64) string
}

type solverUsecaseImpl struct {
	pow *hashcash.ProofOfWork
}

// NewSolverUsecase
func NewSolverUsecase() SolverUsecase {
	return &solverUsecaseImpl{}
}

// GenerateChallenge creates a new challenge using the hashcash package.
func (s *solverUsecaseImpl) FindSolution(challenge []byte, difficulty uint64) string {
	return s.pow.FindSolution(challenge, difficulty)
}
