package config

type KeyValuePair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Configuration struct {
	ID         string            `json:"id"`
	Version    string            `json:"version"`
	Labels     map[string]string `json:"labels"`
	Parameters []KeyValuePair    `json:"parameters"`
}

type ConfigurationGroup struct {
	ID             string            `json:"id"`
	Version        string            `json:"version"`
	Labels         map[string]string `json:"labels"`
	Configurations []Configuration   `json:"configurations"`
}
