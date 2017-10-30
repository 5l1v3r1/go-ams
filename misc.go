package ams

import (
	"fmt"
	"time"
)

func toResource(name, id string) string {
	return fmt.Sprintf("%s('%s')", name, id)
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
