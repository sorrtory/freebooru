package config

import "fmt"

type Config struct{}

func Load() {
	fmt.Println("Loading config...")
}
