package utils

import (
	"regexp"
	"strings"
	"time"
)

func ParseTimestampToUTC7(timestamp int) (time.Time, error) {
	t := time.Unix(int64(timestamp), 0)

	// Define the UTC+7 time zone
	utcPlus7 := time.FixedZone("UTC+7", 7*60*60)

	// Convert the time to UTC+7
	t = t.In(utcPlus7)

	return t, nil
}

func ConvertToSnakeCase(input string) string {
	// Convert to lowercase
	lowercase := strings.ToLower(input)

	// Replace spaces with underscores
	underscored := strings.ReplaceAll(lowercase, " ", "_")

	// Use regular expression to remove non-alphanumeric characters
	re := regexp.MustCompile("[^a-z0-9_]")
	snakeCase := re.ReplaceAllString(underscored, "")

	return snakeCase
}
