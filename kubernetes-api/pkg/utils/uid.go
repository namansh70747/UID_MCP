package utils

import (
	"crypto/rand"
	"fmt"
	"strings"
)

func GenerateUID() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

func GeneratePodName(baseName string) string {
	uid := GenerateUID()
	return fmt.Sprintf("%s-%s", baseName, uid)
}

func SanitizeName(name string) string {
	// Kubernetes names must be DNS-1123 compliant
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "_", "-")
	return name
}
