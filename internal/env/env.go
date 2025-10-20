package env

import (
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	godotenv.Load(".env")
	return nil
}

func ValidateEnv() error {
	requiredEnvs := []string{
		"DATABASE_URL",
	}

	for _, env := range requiredEnvs {
		if _, exists := lookupEnv(env); !exists {
			return &MissingEnvError{env}
		}
	}

	return nil
}

func lookupEnv(key string) (string, bool) {
	if env := os.Getenv(key); env != "" {
		return env, true
	}

	return "", false
}

type MissingEnvError struct {
	VarName string
}

func (e *MissingEnvError) Error() string {
	return "environment variable not set: " + e.VarName
}
