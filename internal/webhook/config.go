package webhook

// Config holds the webhook server settings.
type Config struct {
	Port            int
	TLSCertPath     string
	TLSKeyPath      string
	PolicyPath      string
	SchemaDir       string
	RegistryPrefix  string
	FailOpen        bool
	CacheTTLSeconds int
}

// DefaultConfig returns the default webhook configuration.
func DefaultConfig() Config {
	return Config{
		Port:            8443,
		SchemaDir:       "schemas/v1",
		FailOpen:        false,
		CacheTTLSeconds: 300,
	}
}
