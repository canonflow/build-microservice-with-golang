package main

import (
	"context"
	"fmt"
	"github.com/canonflow/build-microservice-with-golang/application"
)

func main() {
	app := application.New()
	err := app.Start(context.TODO())

	if err != nil {
		fmt.Println("Failed to start application:", err)
	}
}
