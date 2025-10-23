package autohttps

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Services map[string]string
}

// NewConfig creates a new Config by reading the SERVICES environment variable
// The SERVICES variable should be a comma-separated list of services in the format:
// "serviceName:port=domain.com,anotherService:port"
// If the domain is omitted, sslip.io will be used with the autodiscovered public IP.
func NewConfig() (*Config, error) {
	sslipService, err := NewSSLIPService()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSLIPService: %v", err)
	}

	services := make(map[string]string)
	envServices := os.Getenv("SERVICES")
	if envServices == "" {
		return nil, fmt.Errorf("SERVICES environment variable is not set")
	}

	serviceList := strings.Split(envServices, ",")
	for _, service := range serviceList {
		// If not contains '=', use sslip.io as default service
		if !strings.Contains(service, "=") {
			// Check that service includes colon to separate name and port
			if !strings.Contains(service, ":") {
				return nil, fmt.Errorf("invalid service format (missing port): %s", service)
			}
			serviceDomain := sslipService.GetSSLIPServiceDomain(service)
			services[service] = "https://" + serviceDomain
			continue
		}
		parts := strings.SplitN(service, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid service format: %s", service)
		}
		serviceName := parts[0]
		serviceDomain := parts[1]
		// Check that serviceName includes colon to separate name and port
		if !strings.Contains(serviceName, ":") {
			return nil, fmt.Errorf("invalid service format (missing port): %s", serviceName)
		}
		services[serviceName] = "https://" + serviceDomain
	}

	return &Config{
		Services: services,
	}, nil
}
