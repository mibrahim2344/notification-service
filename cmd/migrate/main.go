package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/mibrahim2344/notification-service/internal/infrastructure/db"
)

func main() {
	// Define command line flags
	var (
		command = flag.String("command", "", "Migration command (up/down/version/force/steps)")
		steps   = flag.Int("steps", 0, "Number of migration steps (for steps command)")
		version = flag.Int("version", 0, "Target version (for force command)")
	)

	flag.Parse()

	// Get database configuration from environment variables
	config := db.PostgresConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvAsInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "notification_service"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Create migration manager
	manager, err := db.NewMigrationManager(db.MigrationConfig{
		MigrationsPath: "migrations",
		DBConfig:       config,
	})
	if err != nil {
		fmt.Printf("Failed to create migration manager: %v\n", err)
		os.Exit(1)
	}

	// Execute command
	switch *command {
	case "up":
		if err := manager.Up(); err != nil {
			fmt.Printf("Failed to run migrations: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Successfully ran all migrations")

	case "down":
		if err := manager.Down(); err != nil {
			fmt.Printf("Failed to rollback migrations: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Successfully rolled back all migrations")

	case "version":
		version, dirty, err := manager.Version()
		if err != nil {
			fmt.Printf("Failed to get version: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Current version: %d (dirty: %v)\n", version, dirty)

	case "force":
		if err := manager.Force(*version); err != nil {
			fmt.Printf("Failed to force version: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully forced version to %d\n", *version)

	case "steps":
		if err := manager.Steps(*steps); err != nil {
			fmt.Printf("Failed to run migration steps: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully ran %d migration steps\n", *steps)

	default:
		fmt.Println("Invalid command. Available commands: up, down, version, force, steps")
		os.Exit(1)
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
