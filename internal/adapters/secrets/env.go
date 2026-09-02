package secrets

import (
	"context"
	"fmt"
	"os"
)

// Env is the MVP stand-in for OS Keychain. Production keychain is a non-goal.
type Env struct{}

func New() *Env { return &Env{} }

func (e *Env) Get(_ context.Context, name string) (string, error) {
	if v := os.Getenv(name); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("secret %s not set", name)
}
