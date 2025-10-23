package autohttps

import (
	"fmt"
	"os"
	"text/template"
)

const CaddyfileTemplate = `
{
    admin off
}

{{ range $service, $url := .Services }}
{{ $url }} {
    reverse_proxy {{ $service }}
}
{{ end }}
`

type CaddyGenerator struct {
	Config *Config
}

func NewCaddyGenerator(config *Config) *CaddyGenerator {
	return &CaddyGenerator{
		Config: config,
	}
}

func (cg *CaddyGenerator) GenerateCaddyfile(path string) error {
	caddyfile := CaddyfileTemplate

	// Apply the template
	type templateData struct {
		Services map[string]string
	}
	data := templateData{
		Services: cg.Config.Services,
	}

	tmpl, err := template.New("caddyfile").Parse(caddyfile)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}

func (cg *CaddyGenerator) PrintServices() {
	fmt.Println("---------------------------------------")
	fmt.Println("Configured Services:")
	fmt.Println("---------------------------------------")
	fmt.Println("")
	for service, url := range cg.Config.Services {
		fmt.Println("    - Service:", service, "-> URL:", url)
	}
	fmt.Println("")
	fmt.Println("---------------------------------------")
}
