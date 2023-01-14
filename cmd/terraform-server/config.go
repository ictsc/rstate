package main

type Config struct {
	AdminPass string          `yaml:"adminpass"`
	RootURL   string          `yaml:"rootURL"`
	LogLevel  string          `yaml:"logLevel"`
	Worker    WorkerConfig    `yaml:"worker"`
	Terraform TerraformConfig `yaml:"terraform"`
}

type WorkerConfig struct {
	MaxThread int `yaml:"maxThread"`
}

type TerraformConfig struct {
	Options TerraformOptions `yaml:"options"`
	Secrets TerraformSecrets `yaml:"secrets"`
}

type TerraformOptions struct {
	Path             string `yaml:"path"`
	WorkingDirectory string `yaml:"workingDirectory"`
	Parallelism      int    `yaml:"parallelism"`
}

type TerraformSecrets struct {
	SakuraCloudAccessToken       string `yaml:"sakuraCloudAccessToken"`
	SakuraCloudAccessTokenSecret string `yaml:"sakuraCloudAccessTokenSecret"`

	AwsAccessKeyID     string `yaml:"awsAccessKeyId"`
	AwsSecretAccessKey string `yaml:"awsSecretAccessKey"`
	CloudflareAPIToken string `yaml:"cloudflareApiToken"`
}
