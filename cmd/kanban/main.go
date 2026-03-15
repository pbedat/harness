package main

import (
	"fmt"
	"os"

	"github.com/pbedat/harness/adapters/in/cli"
	"github.com/pbedat/harness/service"
)

func main() {
	application := service.NewApplication()
	err := cli.Create(application).Execute()

	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
