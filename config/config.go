package config

import "os"

type Config struct {
	MovieBaseURL string
	HTTPAddr     string
	GRPCAddr     string
}

// Load читает настройки из окружения, если их нет - берет значения по умолчанию
func Load() Config {
	return Config{
		MovieBaseURL: getenv("MOVIE_BASE_URL", "https://homeworksite.site"),
		HTTPAddr:     getenv("HTTP_ADDR", ":8080"),
		GRPCAddr:     getenv("GRPC_ADDR", ":9090"),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
