// Package smartcontact is the application entry point for the Smart Contact
// service. It is the Go equivalent of Spring Boot's @SpringBootApplication main
// class: it wires together the HTTP router, middleware, and (eventually) the
// handler/service/repository dependencies.
//
// MIGRATION_NOTE: The original Java class merely delegated to
// SpringApplication.run(...), which performed component scanning,
// auto-configuration, and embedded-server startup by convention. Go has no
// equivalent "magic" bootstrap: dependency wiring is explicit. The embedded
// server lifecycle (listen + graceful shutdown) lives in cmd/server/main.go,
// which calls buildRouter() from its own package to obtain the fully-configured
// http.Handler.
package smartcontact
