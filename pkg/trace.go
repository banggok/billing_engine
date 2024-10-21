// pkg/trace.go
package pkg

import "github.com/google/uuid"

func GenerateTraceID() string {
	return uuid.New().String()
}
