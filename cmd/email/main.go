package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/spf13/afero"
	"github.com/spf13/viper"

	events "github.com/pbedat/harness/common/event"
	"github.com/pbedat/harness/modules/email/adapters/in/cli"
	"github.com/pbedat/harness/modules/email/service"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}
	viper.SetEnvPrefix("harness_email")
	viper.AutomaticEnv()

	bus := events.NewBus()
	app := service.NewApplication(bus, afero.NewOsFs(), ".maildata")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	bus.Run(ctx)

	cmd := cli.Create(app)
	cmd.SetContext(ctx)
	if err := cmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
