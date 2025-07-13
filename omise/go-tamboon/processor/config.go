package processor

import (
	"os"
	"strconv"
)

var (
	maxRecords      = defaultMaxRecords
	expYearIncrease = defaultExpYearIncrease
)

func InitConfig() {
	maxRecords = getEnvInt("MAX_RECORDS", defaultMaxRecords)
	expYearIncrease = getEnvInt("EXP_YEAR_INCREASE", defaultExpYearIncrease)
}

func getEnvInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
