package domain

// ProofOfWork defines the PoW entity, including the challenge and difficulty.
type ProofOfWork struct {
	Challenge  []byte
	Difficulty uint64
}

// Quote defines a simple quote structure.
type Quote struct {
	Text string
}
