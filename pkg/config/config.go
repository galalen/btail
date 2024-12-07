package config

type Config struct {
	Lines        int
	Follow       bool
	UIBufferSize int
	BufferSize   int
}

func NewDefaultConfig() Config {
	return Config{
		Lines:        10,
		Follow:       false,
		UIBufferSize: 500,
		BufferSize:   500,
	}
}
