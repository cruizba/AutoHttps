package main

import "github.com/cruizba/autohttps/internal/autohttps"

func main() {
	config, err := autohttps.NewConfig()
	if err != nil {
		panic(err)
	}

	generator := autohttps.NewCaddyGenerator(config)
	err = generator.GenerateCaddyfile("Caddyfile")
	if err != nil {
		panic(err)
	}

}
