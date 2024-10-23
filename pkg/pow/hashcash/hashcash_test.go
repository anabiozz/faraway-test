package hashcash

import (
	"strings"
	"testing"
)

func TestNewProofOfWork(t *testing.T) {
	// Valid difficulty test
	pow, err := NewProofOfWork(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pow.GetDifficulty() != 5 {
		t.Fatalf("expected difficulty 5, got %d", pow.GetDifficulty())
	}

	// Invalid difficulty tests
	_, err = NewProofOfWork(0)
	if err == nil || !strings.Contains(err.Error(), ErrDifficultyRange.Error()) {
		t.Fatalf("expected ErrDifficultyRange for difficulty 0, got %v", err)
	}

	_, err = NewProofOfWork(65)
	if err == nil || !strings.Contains(err.Error(), ErrDifficultyRange.Error()) {
		t.Fatalf("expected ErrDifficultyRange for difficulty 65, got %v", err)
	}
}

func TestGenerateChallenge(t *testing.T) {
	pow, err := NewProofOfWork(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	challenge, err := pow.GenerateChallenge()
	if err != nil {
		t.Fatalf("unexpected error generating challenge: %v", err)
	}
	if len(challenge) != tokenLength {
		t.Fatalf("expected challenge length %d, got %d", tokenLength, len(challenge))
	}
}

func TestVerify(t *testing.T) {
	pow, err := NewProofOfWork(2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Generate a random challenge
	challenge, err := pow.GenerateChallenge()
	if err != nil {
		t.Fatalf("unexpected error generating challenge: %v", err)
	}

	// Use FindSolution to compute a solution that meets the difficulty requirement
	solution := pow.FindSolution(challenge, pow.GetDifficulty())

	// Verify if the solution is valid
	if !pow.Verify(challenge, []byte(solution)) {
		t.Fatalf("expected valid proof-of-work verification")
	}
}

func TestFindSolution(t *testing.T) {
	pow, err := NewProofOfWork(2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	challenge := []byte("challenge")
	solution := pow.FindSolution(challenge, pow.GetDifficulty())

	// Verify the solution
	if !pow.Verify(challenge, []byte(solution)) {
		t.Fatalf("expected valid solution but verification failed")
	}
}

func TestFindSolutionWithHigherDifficulty(t *testing.T) {
	pow, err := NewProofOfWork(4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	challenge := []byte("challenge")
	solution := pow.FindSolution(challenge, pow.GetDifficulty())

	// Verify the solution
	if !pow.Verify(challenge, []byte(solution)) {
		t.Fatalf("expected valid solution but verification failed")
	}
}
