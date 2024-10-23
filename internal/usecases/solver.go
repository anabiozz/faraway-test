package usecases

import (
	"faraway/pkg/pow/argon2"
	"faraway/pkg/pow/hashcash"
	"fmt"
)

type SolverUsecase interface {
	FindCPUBoundSolution(challenge []byte) string
	FindMemoryBoundSolution(challenge []byte) (string, error)
}

type solverUsecaseImpl struct {
	hashcash *hashcash.HashCash
	argon2   *argon2.Argon2
}

// NewSolverUsecase
func NewSolverUsecase(difficulty uint64) (SolverUsecase, error) {
	hashcash, err := hashcash.NewHashCash(difficulty)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize hashcash: %w", err)
	}
	argon2, err := argon2.NewArgon2(difficulty)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize argon2: %w", err)
	}
	return &solverUsecaseImpl{
		hashcash: hashcash,
		argon2:   argon2,
	}, nil
}

func (s *solverUsecaseImpl) FindCPUBoundSolution(challenge []byte) string {
	return s.hashcash.FindSolution(challenge)
}

func (s *solverUsecaseImpl) FindMemoryBoundSolution(challenge []byte) (string, error) {
	return s.argon2.FindSolution(challenge)
}
