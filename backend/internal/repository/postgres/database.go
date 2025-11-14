// Package postgres contains minimal compatibility shims. The main
// Postgres initialization and connection logic lives in connection.go
// to avoid duplicate symbol declarations. See connection.go for the
// GORM + sqlx setup used by the application.
package postgres
