package config

type Pow struct {
	Difficulty uint64 `envconfig:"DIFFICULTY" required:"true"`
}
