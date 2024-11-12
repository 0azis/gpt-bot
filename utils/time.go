package utils

import (
	"time"
)

func AddMonth() string {
	t := time.Now().AddDate(0, 1, 0).Format(time.DateOnly)
	return t
}

func AddYear() string {
	t := time.Now().AddDate(1, 0, 0).Format(time.DateOnly)
	return t
}
