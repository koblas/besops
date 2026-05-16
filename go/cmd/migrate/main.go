package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/koblas/besops/internal/database"
	"github.com/spf13/pflag"
)

type config struct {
	DatabaseURL string `koanf:"database_url"`
	Direction   string `koanf:"direction"`
}

func main() {
	fs := pflag.CommandLine
	fs.String("database-url", "sqlite://./data/besops.db", "database connection URL")
	fs.String("direction", "up", "migration direction: up or down")
	pflag.Parse()

	k := koanf.New(".")

	if err := k.Load(env.Provider("", "_", func(s string) string {
		return strings.ToLower(s)
	}), nil); err != nil {
		log.Fatalf("loading env: %v", err)
	}

	if err := k.Load(posflag.Provider(fs, ".", k), nil); err != nil {
		log.Fatalf("loading flags: %v", err)
	}

	var cfg config
	if err := k.Unmarshal("", &cfg); err != nil {
		log.Fatalf("unmarshalling config: %v", err)
	}

	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = "sqlite://./data/besops.db"
	}

	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	switch cfg.Direction {
	case "up":
		if err := database.Migrate(db, cfg.DatabaseURL); err != nil {
			log.Fatalf("migrate up: %v", err)
		}
		fmt.Println("migrations applied successfully")
	case "down":
		if err := database.MigrateDown(db, cfg.DatabaseURL); err != nil {
			log.Fatalf("migrate down: %v", err)
		}
		fmt.Println("migrations rolled back successfully")
	default:
		log.Fatalf("unknown direction: %s (use 'up' or 'down')", cfg.Direction)
	}
}
