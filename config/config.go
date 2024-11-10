package config

import (
	"fmt"
	"os"
)

// Config
type Config struct {
	Server    server
	Db        database
	Tokens    tokens
	WebAppUrl string
}

func New() Config {
	server := newServer()
	db := newDatabase()
	tokens := newTokens()
	return Config{
		Server:    server,
		Db:        db,
		Tokens:    tokens,
		WebAppUrl: getEnv("WEB_APP_URL", ""),
	}
}

// server config
type server struct {
	host string
	port string
}

func (s server) Addr() string {
	return s.host + ":" + s.port
}

func newServer() server {
	return server{
		host: getEnv("HTTP_HOST", "localhost"),
		port: getEnv("HTTP_PORT", "5000"),
	}
}

// database config
type database struct {
	host     string
	port     string
	name     string
	user     string
	password string
}

func (d database) Addr() string {
	uri := fmt.Sprintf("%s:%s@(%s:%s)/%s?parseTime=true", d.user, d.password, d.host, d.port, d.name)
	return uri
}

func newDatabase() database {
	return database{
		host:     getEnv("DATABASE_HOST", "localhost"),
		port:     getEnv("DATABASE_POST", "3306"),
		name:     getEnv("DATABASE_NAME", ""),
		user:     getEnv("DATABASE_USER", ""),
		password: getEnv("DATABASE_PASSWORD", ""),
	}
}

// tokens config
type tokens struct {
	telegram string
	openai   string
	runware  string
}

func (t tokens) Telegram() string {
	return t.telegram
}

func (t tokens) OpenAI() string {
	return t.openai
}

func (t tokens) Runware() string {
	return t.runware
}

func newTokens() tokens {
	return tokens{
		telegram: getEnv("TELEGRAM_TOKEN", ""),
		openai:   getEnv("OPENAI_TOKEN", ""),
		runware:  getEnv("RUNWARE_TOKEN", ""),
	}
}

// type JwtSecretHash string

func JwtSecretHash() []byte {
	return []byte(getEnv("JWT_SECRET_HASH", ""))
}

// helper function
func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
