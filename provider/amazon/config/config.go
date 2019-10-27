package config

type Amazon struct {
	Region    string   `yaml:"region"`
	IdKey     string   `yaml:"id_key"`
	SecretKey string   `yaml:"secret_key"`
	Instance  Instance `yaml:"instance"`
}

type Instance struct {
	PrefixName     string `yaml:"prefix_name"`
	ImageID        string `yaml:"image_id"`
	Type           string `yaml:"type"`
	SecurityGroups struct {
		Id      string `yaml:"id"`
		KeyPair string `yaml:"key_pair"`
	} `yaml:"security_groups"`
}
