package autohttps

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Services map[string]string
}

// parsedService holds a service target (name:port) and its optional domain.
type parsedService struct {
	target string // e.g. "web:3000"
	domain string // explicit domain, empty when none was provided
}

// NewConfig creates a new Config by reading the SERVICES environment variable.
//
// SERVICES is a comma-separated list of services in the format:
//
//	serviceName:port[=domain.com][,anotherService:port]
//
// When a service has no domain, a sslip.io domain is generated from the
// autodiscovered public IP. As a convenience for single-service deployments
// (e.g. a docker-compose template published in a registry), the DOMAIN
// environment variable can set the domain without having to edit SERVICES.
func NewConfig() (*Config, error) {
	envServices := os.Getenv("SERVICES")
	if envServices == "" {
		return nil, fmt.Errorf("SERVICES environment variable is not set")
	}

	parsed, err := parseServices(envServices)
	if err != nil {
		return nil, err
	}

	if err := applyDomainEnv(parsed, os.Getenv("DOMAIN")); err != nil {
		return nil, err
	}

	services, err := resolveDomains(parsed, NewSSLIPService)
	if err != nil {
		return nil, err
	}

	return &Config{
		Services: services,
	}, nil
}

// parseServices splits the SERVICES value into target/domain pairs.
//
// A trailing "=" with an empty domain is treated as "no domain" so that
// docker-compose substitutions like SERVICES=web:3000=${DOMAIN} gracefully
// fall back to sslip.io when the variable is unset, instead of producing an
// invalid configuration.
func parseServices(envServices string) ([]parsedService, error) {
	var result []parsedService
	for _, service := range strings.Split(envServices, ",") {
		target := service
		domain := ""
		if i := strings.Index(service, "="); i >= 0 {
			target = service[:i]
			domain = service[i+1:]
		}
		if !strings.Contains(target, ":") {
			return nil, fmt.Errorf("invalid service format (missing port): %s", service)
		}
		result = append(result, parsedService{target: target, domain: domain})
	}
	return result, nil
}

// applyDomainEnv applies the optional DOMAIN environment variable. It is a
// convenience for the common single-service deployment; multi-service setups
// must specify each domain inline in SERVICES.
func applyDomainEnv(parsed []parsedService, domain string) error {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return nil
	}
	if len(parsed) != 1 {
		return fmt.Errorf("DOMAIN can only be used with a single service in SERVICES; specify each domain inline (service:port=domain) when using multiple services")
	}
	if parsed[0].domain != "" {
		return fmt.Errorf("domain specified twice: via the DOMAIN variable and inline in SERVICES (%q); use only one", parsed[0].domain)
	}
	parsed[0].domain = domain
	return nil
}

// resolveDomains turns parsed services into a target -> URL map, generating
// sslip.io domains only for the services that still lack one. The sslip service
// is created lazily via newSSLIP, so configurations where every service has an
// explicit domain do not require network access (and tests can inject a fixed
// public IP).
func resolveDomains(parsed []parsedService, newSSLIP func() (*SSLIPService, error)) (map[string]string, error) {
	var sslipService *SSLIPService
	services := make(map[string]string)
	for _, p := range parsed {
		domain := p.domain
		if domain == "" {
			if sslipService == nil {
				var err error
				sslipService, err = newSSLIP()
				if err != nil {
					return nil, fmt.Errorf("failed to create SSLIPService: %v", err)
				}
			}
			domain = sslipService.GetSSLIPServiceDomain(p.target)
		}
		services[p.target] = "https://" + domain
	}
	return services, nil
}
