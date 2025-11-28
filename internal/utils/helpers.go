package utils

import (
	"log"
	"os"
)

func MustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("Fatal Error: %s environment variable not set.\n", k)
	}
	return v
}
