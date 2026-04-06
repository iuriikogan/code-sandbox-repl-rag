package data

import (
	"os"
	"strings"
)

const baseChunk = `Employee John Doe took 3 days off in March. HR policy states max 15 days a year.
Employee Jane Smith requested a salary review.
Q1 Revenue was $2.4M, up 12% from last year.
Server costs spiked to $45k in February due to GPU usage.
Deployment to production failed on Mar 12 due to DB lock.
Migrated 45 microservices to new Kubernetes cluster.
`

// GenerateContext creates a simulated large unstructured dataset with the specified number of chunk repeats.
func GenerateContext(repeats int) string {
	return strings.Repeat(baseChunk, repeats)
}

// GenerateMassiveContext creates a simulated large unstructured dataset (50 repeats).
func GenerateMassiveContext() string {
	return GenerateContext(50)
}

// CreateContextFile writes the massive context to a temporary file.
// It returns the file path, a cleanup function, and an error.
func CreateContextFile(content string) (string, func(), error) {
	tmpFile, err := os.CreateTemp("", "context-*.txt")
	if err != nil {
		return "", nil, err
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", nil, err
	}

	tmpFile.Close()

	cleanup := func() {
		os.Remove(tmpFile.Name())
	}

	return tmpFile.Name(), cleanup, nil
}