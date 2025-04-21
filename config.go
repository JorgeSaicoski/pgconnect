// Package pgconnect provides utilities for connecting to PostgreSQL
// using GORM and integrating with Gin framework.
package pgconnect

import (
	"gorm.io/gorm/logger"
)

// Config holds database connection configuration
type Config struct {
	Host         string
	Port         string
	User         string
	Password     string
	DatabaseName string
	SSLMode      string
	TimeZone     string
	MaxIdleConns int
	MaxOpenConns int
	LogLevel     logger.LogLevel
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		Host:         "localhost",
		Port:         "5432",
		User:         "postgres",
		Password:     "postgres",
		DatabaseName: "postgres",
		SSLMode:      "disable",
		TimeZone:     "UTC",
		MaxIdleConns: 10,
		MaxOpenConns: 100,
		LogLevel:     logger.Silent,
	}
}
