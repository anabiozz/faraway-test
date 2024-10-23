package config

type Client struct {
	ServerAddr string `envconfig:"SERVER_ADDR" required:"true"`
	Name       string `envconfig:"NAME" required:"true"`
}
