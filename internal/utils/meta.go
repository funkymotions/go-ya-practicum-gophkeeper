package utils

import (
	"fmt"
	"strings"
)

func GetAppMetaString(version, date string) string {
	sb := strings.Builder{}
	if version == "" {
		version = "N/A"
	}
	if date == "" {
		date = "N/A"
	}

	sb.WriteString(fmt.Sprintf("\nBuild version: %s\n", version))
	sb.WriteString(fmt.Sprintf("Build date: %s\n", date))

	return sb.String()
}
