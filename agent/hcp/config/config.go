package config

import (
	"crypto/tls"

	hcpcfg "github.com/hashicorp/hcp-sdk-go/config"
)

// CloudConfig defines configuration for connecting to HCP services
type CloudConfig struct {
	ResourceID   string
	ClientID     string
	ClientSecret string
	Hostname     string
	AuthURL      string
	ScadaAddress string

	// Management token used by HCP management plane.
	// Cannot be set via config files.
	ManagementToken string

	// TlsConfig for testing.
	TLSConfig *tls.Config
}

func (c *CloudConfig) WithTLSConfig(cfg *tls.Config) {
	c.TLSConfig = cfg
}

func (c *CloudConfig) HCPConfig(opts ...hcpcfg.HCPConfigOption) (hcpcfg.HCPConfig, error) {
	if c.TLSConfig == nil {
		c.TLSConfig = &tls.Config{}
	}
	if c.ClientID != "" && c.ClientSecret != "" {
		opts = append(opts, hcpcfg.WithClientCredentials(c.ClientID, c.ClientSecret))
	}
	if c.AuthURL != "" {
		opts = append(opts, hcpcfg.WithAuth(c.AuthURL, c.TLSConfig))
	}
	if c.Hostname != "" {
		opts = append(opts, hcpcfg.WithAPI(c.Hostname, c.TLSConfig))
	}
	if c.ScadaAddress != "" {
		opts = append(opts, hcpcfg.WithSCADA(c.ScadaAddress, c.TLSConfig))
	}
	opts = append(opts, hcpcfg.FromEnv())
	return hcpcfg.NewHCPConfig(opts...)
}