package autohttps

import (
	"testing"
)

func TestSSLIPService(t *testing.T) {
	t.Run("creates new service", func(t *testing.T) {
		service, err := NewSSLIPService()
		if err != nil {
			t.Fatalf("NewSSLIPService() error = %v", err)
		}
		if service == nil {
			t.Fatal("NewSSLIPService() returned nil")
		}
		if service.publicip == "" {
			t.Error("NewSSLIPService() returned empty public IP")
		}
	})
}

func TestGetSSLIPServiceDomain(t *testing.T) {
	service := &SSLIPService{
		publicip: "192.168.1.1",
	}

	tests := []struct {
		name   string
		prefix string
		want   string
	}{
		{
			name:   "valid prefix with port",
			prefix: "service:8080",
			want:   "service-192-168-1-1.sslip.io",
		},
		{
			name:   "empty prefix",
			prefix: "",
			want:   "192-168-1-1.sslip.io",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.GetSSLIPServiceDomain(tt.prefix)
			if got != tt.want {
				t.Errorf("GetSSLIPServiceDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseIPForSSLIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want string
	}{
		{
			name: "ipv4",
			ip:   "192.168.1.1",
			want: "192-168-1-1",
		},
		{
			name: "ipv6",
			ip:   "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			want: "2001-0db8-85a3-0000-0000-8a2e-0370-7334",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseIPForSSLIP(tt.ip); got != tt.want {
				t.Errorf("parseIPForSSLIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
