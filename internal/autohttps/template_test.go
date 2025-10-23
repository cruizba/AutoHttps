package autohttps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCaddyGenerator(t *testing.T) {
	config := &Config{
		Services: map[string]string{
			"service1:8080": "https://example.com",
			"service2:8081": "https://test.com",
		},
	}

	generator := NewCaddyGenerator(config)
	if generator == nil {
		t.Fatal("NewCaddyGenerator returned nil")
	}

	t.Run("generates valid caddyfile", func(t *testing.T) {
		tempDir := t.TempDir()
		caddyfilePath := filepath.Join(tempDir, "Caddyfile")

		err := generator.GenerateCaddyfile(caddyfilePath)
		if err != nil {
			t.Fatalf("GenerateCaddyfile() error = %v", err)
		}

		// Read generated file
		content, err := os.ReadFile(caddyfilePath)
		if err != nil {
			t.Fatalf("Failed to read generated Caddyfile: %v", err)
		}

		// Check content
		contentStr := string(content)
		if !strings.Contains(contentStr, "admin off") {
			t.Error("Generated Caddyfile missing admin off directive")
		}

		// Check each service is properly configured
		for service, url := range config.Services {
			if !strings.Contains(contentStr, url) {
				t.Errorf("Generated Caddyfile missing URL %s", url)
			}
			if !strings.Contains(contentStr, service) {
				t.Errorf("Generated Caddyfile missing service %s", service)
			}
		}
	})

	t.Run("fails with invalid path", func(t *testing.T) {
		err := generator.GenerateCaddyfile("/nonexistent/directory/Caddyfile")
		if err == nil {
			t.Error("GenerateCaddyfile() expected error for invalid path")
		}
	})
}
