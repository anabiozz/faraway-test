package config

import "time"

type Server struct {
	Addr      string        `envconfig:"ADDR" required:"true"`
	Name      string        `envconfig:"NAME" required:"true"`
	Deadline  time.Duration `envconfig:"DEADLINE" required:"true"`
	KeepAlive time.Duration `envconfig:"SERVER_KEEP_ALIVE,default=15s"`
}
