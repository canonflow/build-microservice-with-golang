package main

import (
	"context"
	"fmt"
	"github.com/canonflow/build-microservice-with-golang/application"
	"os"
	"os/signal"
)

func main() {
	app := application.New()

	// Root level context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	err := app.Start(ctx)
	// Kill the root and its children context
	// defer -> will run at the end of the file
	defer cancel()

	if err != nil {
		fmt.Println("Failed to start application:", err)
	}

}
