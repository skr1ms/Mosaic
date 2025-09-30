// To adapt the config to the ConfigInterface interface

package config

func (c *Config) GetRecaptchaConfig() RecaptchaConfig {
	return c.RecaptchaConfig
}

func (c *Config) GetServerConfig() ServerConfig {
	return c.ServerConfig
}

func (c *Config) GetAlfaBankConfig() AlphaBankConfig {
	return c.AlphaBankConfig
}

func (c *Config) GetS3MinioConfig() S3MinioConfig {
	return c.S3MinioConfig
}

func (c *Config) GetStableDiffusionConfig() StableDiffusionConfig {
	return c.StableDiffusionConfig
}

func (c *Config) GetMosaicGeneratorConfig() MosaicGeneratorConfig {
	return c.MosaicGeneratorConfig
}

func (c *Config) GetGitLabConfig() GitLabConfig {
	return c.GitLabConfig
}
