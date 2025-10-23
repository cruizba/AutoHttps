package autohttps

import (
	"context"
	"time"

	"github.com/cruizba/publicip"
)

// SSLIPService represents a service that provides URLs using sslip.io
type SSLIPService struct {
	publicip string
}

// Creates a new SSLIPService by autodiscovering the public IP address
func NewSSLIPService() (*SSLIPService, error) {
	client := publicip.NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ip, err := client.Discover(ctx)
	if err != nil {
		return nil, err
	}

	return &SSLIPService{
		publicip: ip.String(),
	}, nil
}

// Returns given a prefix, the Full URL to access the service via
// the autodiscovered public IP address using sslip.io
func (s *SSLIPService) GetSSLIPServiceDomain(prefix string) string {
	if prefix == "" {
		return parseIPForSSLIP(s.publicip) + ".sslip.io"
	}
	// Remove port if present
	for i := 0; i < len(prefix); i++ {
		if prefix[i] == ':' {
			prefix = prefix[:i]
			break
		}
	}
	return prefix + "-" + parseIPForSSLIP(s.publicip) + ".sslip.io"
}

// parseIPForSSLIP converts an IP address into a format suitable for sslip.io
func parseIPForSSLIP(ip string) string {
	// Replace dots and colons with dashes
	result := ""
	for _, ch := range ip {
		if ch == '.' || ch == ':' {
			result += "-"
		} else {
			result += string(ch)
		}
	}
	return result
}
