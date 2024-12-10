package config

import (
	"fmt"
	"log/slog"
	"os"
)

// Config
type Config struct {
	Server   Server
	Database Database
	Api      Api
	Telegram Telegram
}

func New() Config {
	server := newServer()
	db := newDatabase()
	api := newApi()
	telegram := newTelegram()
	return Config{
		Server:   server,
		Database: db,
		Api:      api,
		Telegram: telegram,
	}
}

func (c Config) IsValid() bool {
	if !c.Server.IsValid() {
		attr := slog.Any("server", c.Server)
		slog.Info("", attr)
		return false
	}

	if !c.Database.IsValid() {
		attr := slog.Any("database", c.Database)
		slog.Info("", attr)
		return false
	}

	if !c.Api.IsValid() {
		attr := slog.Any("api", c.Api)
		slog.Info("", attr)
		return false
	}

	if !c.Telegram.IsValid() {
		attr := slog.Any("telegram", c.Telegram)
		slog.Info("", attr)
		return false
	}

	return true
}

// server config
type Server struct {
	host     string
	port     string
	savePath string
}

func (s Server) IsValid() bool {
	return s.host != "" && s.port != "" && string(JwtSecretHash()) != "" && s.savePath != ""
}

func (s Server) Addr() string {
	return s.host + ":" + s.port
}

func (s Server) SavePath() string {
	return s.savePath
}

func newServer() Server {
	return Server{
		host:     getEnv("HTTP_HOST"),
		port:     getEnv("HTTP_PORT"),
		savePath: getEnv("HTTP_SAVE_PATH"),
	}
}

// database config
type Database struct {
	host     string
	port     string
	name     string
	user     string
	password string
}

func (d Database) IsValid() bool {
	return d.host != "" && d.port != "" && d.name != "" && d.user != "" && d.password != ""
}
func (d Database) Addr() string {
	uri := fmt.Sprintf("%s:%s@(%s:%s)/%s?parseTime=true", d.user, d.password, d.host, d.port, d.name)
	return uri
}

func newDatabase() Database {
	return Database{
		host:     getEnv("DATABASE_HOST"),
		port:     getEnv("DATABASE_PORT"),
		name:     getEnv("DATABASE_NAME"),
		user:     getEnv("DATABASE_USER"),
		password: getEnv("DATABASE_PASSWORD"),
	}
}

type Api struct {
	openai    string
	runware   string
	cryptoBot string
	yooKassa  string
}

func (a Api) IsValid() bool {
	return a.openai != "" && a.runware != "" && a.cryptoBot != "" && a.yooKassa != ""
}

func (a Api) OpenAI() string {
	return a.openai
}

func (a Api) Runware() string {
	return a.runware
}

func (a Api) CryptoBot() string {
	return a.cryptoBot
}

func (a Api) YooKassa() string {
	return a.yooKassa
}

func newApi() Api {
	return Api{
		openai:    getEnv("API_OPENAI_TOKEN"),
		runware:   getEnv("API_RUNWARE_TOKEN"),
		cryptoBot: getEnv("API_CRYPTOBOT_TOKEN"),
		yooKassa:  getEnv("API_YOOKASSA_TOKEN"),
	}
}

type Telegram struct {
	token         string
	adminPassword string
	webAppUrl     string
}

func (t Telegram) IsValid() bool {
	return t.token != "" && t.adminPassword != "" && t.webAppUrl != ""
}
func (t Telegram) GetToken() string {
	return t.token
}

func (t Telegram) GetAdminPassword() string {
	return t.adminPassword
}

func (t Telegram) GetWebAppUrl() string {
	return t.webAppUrl
}

func newTelegram() Telegram {
	return Telegram{
		token:         getEnv("TELEGRAM_TOKEN"),
		adminPassword: getEnv("TELEGRAM_PASSWORD"),
		webAppUrl:     getEnv("TELEGRAM_WEB_APP_URL"),
	}
}

func JwtSecretHash() []byte {
	return []byte(getEnv("HTTP_JWT_SECRET_HASH"))
}

// helper function
func getEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return ""
}
