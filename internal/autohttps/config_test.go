package autohttps

import (
	"os"
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
