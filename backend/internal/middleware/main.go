// This file was previously a duplicate server main placed inside the middleware
// package which created an import cycle (routes -> middleware -> routes).
//
// Keep this file minimal to provide middleware package-level helpers and avoid
// importing other high-level packages. All server startup logic lives in
// `cmd/server/main.go`.

package middleware

// No-op placeholder to ensure package compiles and does not introduce import
// cycles. All middleware implementations are in other files inside this
// package (e.g. auth.go, logger.go, cors.go, graceful_shutdown.go).
