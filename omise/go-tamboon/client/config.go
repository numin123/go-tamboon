package client

import (
	"os"
	"strconv"
)

var (
	maxRetries            = defaultMaxRetries
	maxDonationGoroutines = defaultMaxDonationGoroutines
)

func InitConfig() {
	maxRetries = getEnvInt("MAX_RETRIES", defaultMaxRetries)
	maxDonationGoroutines = getEnvInt("MAX_DONATION_GOROUTINES", defaultMaxDonationGoroutines)
}

func getEnvInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
