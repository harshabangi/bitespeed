package service

import (
	"fmt"
	"github.com/harshabangi/bitespeed/internal/storage"
	"github.com/labstack/echo/v4"
)

type Config struct {
	UserName   string `json:"user"`
	Password   string `json:"password"`
	Database   string `json:"database"`
	Host       string `json:"host"`
	ListenAddr string `json:"listen_addr"`
}

func NewConfig() *Config {
	return &Config{}
}

type Service struct {
	conf    *Config
	storage *storage.Store
}

func NewService(c *Config) (*Service, error) {
	store, err := storage.New(c.UserName, c.Password, c.Host, c.Database)
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}

	return &Service{
		conf:    c,
		storage: store,
	}, nil
}

func (s *Service) Run() {
	e := echo.New()

	// Register app (*App) to be injected into all HTTP handlers.
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("service", s)
			return next(c)
		}
	})

	e.POST("/identify", identify)

	e.Logger.Fatal(e.Start(s.conf.ListenAddr))
}
