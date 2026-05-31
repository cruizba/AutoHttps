package autohttps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name        string
		envServices string
		wantErr     bool
	}{
		{
			name:        "valid services with domain",
			envServices: "service1:8080=example.com,service2:8081=test.com",
			wantErr:     false,
		},
		{
			name:        "valid services with sslip",
			envServices: "service1:8080,service2:8081",
			wantErr:     false,
		},
		{
			name:        "invalid service format missing port",
			envServices: "service1=example.com",
			wantErr:     true,
		},
		{
			name:        "empty services",
			envServices: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("SERVICES", tt.envServices)
			defer os.Unsetenv("SERVICES")

			got, err := NewConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Error("NewConfig() returned nil but expected a Config")
			}

			if !tt.wantErr {
				// Check that services were properly parsed
				for service, url := range got.Services {
					if url == "" {
						t.Errorf("NewConfig() service %s has empty URL", service)
					}
				}
			}
		})
	}
}

func TestParseServices(t *testing.T) {
	tests := []struct {
		name        string
		envServices string
		want        []parsedService
		wantErr     bool
	}{
		{
			name:        "single service without domain",
			envServices: "web:3000",
			want:        []parsedService{{target: "web:3000", domain: ""}},
		},
		{
			name:        "single service with domain",
			envServices: "web:3000=example.com",
			want:        []parsedService{{target: "web:3000", domain: "example.com"}},
		},
		{
			name:        "empty inline domain falls back to no domain",
			envServices: "web:3000=",
			want:        []parsedService{{target: "web:3000", domain: ""}},
		},
		{
			name:        "missing port",
			envServices: "web=example.com",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseServices(tt.envServices)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseServices() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("parseServices() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseServices()[%d] = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestApplyDomainEnv(t *testing.T) {
	tests := []struct {
		name       string
		parsed     []parsedService
		domain     string
		wantDomain string // expected domain on the single service (when no error)
		wantErr    bool
	}{
		{
			name:       "applies domain to single domain-less service",
			parsed:     []parsedService{{target: "web:3000"}},
			domain:     "example.com",
			wantDomain: "example.com",
		},
		{
			name:       "trims whitespace",
			parsed:     []parsedService{{target: "web:3000"}},
			domain:     "  example.com  ",
			wantDomain: "example.com",
		},
		{
			name:       "empty domain is a no-op (falls back to sslip)",
			parsed:     []parsedService{{target: "web:3000"}},
			domain:     "",
			wantDomain: "",
		},
		{
			name:    "errors with multiple services",
			parsed:  []parsedService{{target: "a:80"}, {target: "b:81"}},
			domain:  "example.com",
			wantErr: true,
		},
		{
			name:    "errors when domain set both inline and via DOMAIN",
			parsed:  []parsedService{{target: "web:3000", domain: "inline.com"}},
			domain:  "example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := applyDomainEnv(tt.parsed, tt.domain)
			if (err != nil) != tt.wantErr {
				t.Fatalf("applyDomainEnv() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if tt.parsed[0].domain != tt.wantDomain {
				t.Errorf("applyDomainEnv() domain = %q, want %q", tt.parsed[0].domain, tt.wantDomain)
			}
		})
	}
}

func TestNewConfigWithDomainEnv(t *testing.T) {
	os.Setenv("SERVICES", "web:3000")
	os.Setenv("DOMAIN", "www.example.com")
	defer os.Unsetenv("SERVICES")
	defer os.Unsetenv("DOMAIN")

	got, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	if url := got.Services["web:3000"]; url != "https://www.example.com" {
		t.Errorf("NewConfig() URL = %q, want %q", url, "https://www.example.com")
	}
}

func TestNewConfigDomainEnvMultipleServices(t *testing.T) {
	os.Setenv("SERVICES", "a:80,b:81")
	os.Setenv("DOMAIN", "www.example.com")
	defer os.Unsetenv("SERVICES")
	defer os.Unsetenv("DOMAIN")

	if _, err := NewConfig(); err == nil {
		t.Error("NewConfig() expected error when DOMAIN is set with multiple services")
	}
}

// TestCaddyfileEndToEnd exercises the full path SERVICES/DOMAIN -> NewConfig ->
// GenerateCaddyfile and asserts the actual Caddyfile content for each
// domain-based case. These cases are deterministic and require no network
// (sslip.io is only contacted when a service has no domain).
func TestCaddyfileEndToEnd(t *testing.T) {
	tests := []struct {
		name     string
		services string
		domain   string
		want     []string // substrings that must appear in the Caddyfile
	}{
		{
			name:     "single service with DOMAIN variable",
			services: "web:3000",
			domain:   "www.example.com",
			want: []string{
				"https://www.example.com {",
				"reverse_proxy web:3000",
			},
		},
		{
			name:     "single service with inline domain",
			services: "web:3000=inline.example.com",
			want: []string{
				"https://inline.example.com {",
				"reverse_proxy web:3000",
			},
		},
		{
			name:     "multiple services with inline domains",
			services: "app1:3000=one.example.com,app2:8080=two.example.com",
			want: []string{
				"https://one.example.com {",
				"reverse_proxy app1:3000",
				"https://two.example.com {",
				"reverse_proxy app2:8080",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("SERVICES", tt.services)
			defer os.Unsetenv("SERVICES")
			if tt.domain != "" {
				os.Setenv("DOMAIN", tt.domain)
				defer os.Unsetenv("DOMAIN")
			}

			config, err := NewConfig()
			if err != nil {
				t.Fatalf("NewConfig() error = %v", err)
			}

			path := filepath.Join(t.TempDir(), "Caddyfile")
			if err := NewCaddyGenerator(config).GenerateCaddyfile(path); err != nil {
				t.Fatalf("GenerateCaddyfile() error = %v", err)
			}

			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile() error = %v", err)
			}
			got := string(content)

			for _, w := range tt.want {
				if !strings.Contains(got, w) {
					t.Errorf("Caddyfile missing %q\n--- generated ---\n%s", w, got)
				}
			}
			// A domain-based config must never emit an empty host (the old
			// footgun where an empty domain produced "https:// {").
			if strings.Contains(got, "https:// {") {
				t.Errorf("Caddyfile contains an empty host block\n--- generated ---\n%s", got)
			}
		})
	}
}

// TestCaddyfileSSLIPDomain verifies the generated sslip.io Caddyfile for the
// no-domain cases. A fixed public IP is injected so the output is deterministic
// and the exact sslip.io domain format can be asserted without network access.
func TestCaddyfileSSLIPDomain(t *testing.T) {
	tests := []struct {
		name     string
		services string
	}{
		{name: "service without domain", services: "web:3000"},
		{name: "service with empty inline domain", services: "web:3000="},
	}

	fixedIP := func() (*SSLIPService, error) {
		return &SSLIPService{publicip: "1.2.3.4"}, nil
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parseServices(tt.services)
			if err != nil {
				t.Fatalf("parseServices() error = %v", err)
			}

			services, err := resolveDomains(parsed, fixedIP)
			if err != nil {
				t.Fatalf("resolveDomains() error = %v", err)
			}

			path := filepath.Join(t.TempDir(), "Caddyfile")
			if err := NewCaddyGenerator(&Config{Services: services}).GenerateCaddyfile(path); err != nil {
				t.Fatalf("GenerateCaddyfile() error = %v", err)
			}

			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile() error = %v", err)
			}
			got := string(content)

			want := []string{
				"https://web-1-2-3-4.sslip.io {",
				"reverse_proxy web:3000",
			}
			for _, w := range want {
				if !strings.Contains(got, w) {
					t.Errorf("Caddyfile missing %q\n--- generated ---\n%s", w, got)
				}
			}
			if strings.Contains(got, "https:// {") {
				t.Errorf("Caddyfile contains an empty host block\n--- generated ---\n%s", got)
			}
		})
	}
}
