package config

type DBConfig struct {
	URL      string `mapstructure:"url"`
	Database string `mapstructure:"database"`
}

func (c *DBConfig) Check() error {
	return nil
}
