package configs

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
	}
}

func Get(key string) string {
	return os.Getenv(key)
}

func GetInt(key string, defaultValue int) int {
	if os.Getenv(key) == "" {
		return defaultValue
	}
	i, err := strconv.ParseInt(os.Getenv(key), 10, 32)
	if err != nil {
		return defaultValue
	}
	return int(i)
}

func GetInt64(key string) int64 {
	i, _ := strconv.ParseInt(os.Getenv(key), 10, 64)
	return i
}

func Gets(keys ...*string) {
	for i := range keys {
		*keys[i] = os.Getenv(*keys[i])
	}
}
