package application

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"
)

type App struct {
	router http.Handler
	rdb    *redis.Client
	config Config
}

func New(config Config) *App {
	app := &App{
		rdb: redis.NewClient(&redis.Options{
			Addr: config.RedisAddress,
		}),
	}

	app.loadRoutes()

	return app
}

func (app *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", app.config.ServerPort),
		Handler: app.router,
	}

	// Check Redis Connection
	err := app.rdb.Ping(ctx).Err()

	if err != nil {
		return fmt.Errorf("Failed to connect to redis: %w", err)
	}

	// Grateful shutdown the redis if there is an error
	defer func() {
		if err := app.rdb.Close(); err != nil {
			fmt.Printf("Failed to close redis connection")
		}
	}()

	fmt.Println("Server is running ...")

	// Make channel
	/* ===== 1 specifies the buffer size of the channel =====
	- The channel can hold 1 value in its buffer before it blocks further sends.
	- If a value is sent to the channel and the buffer is full,
	  the send operation (channel <- value) will block until there’s space in the buffer
	  (i.e., until a receiver takes the value).
	- If a value is received from the channel and the buffer is empty,
	  the receive operation (<-channel) will block until a value is available.
	*/
	channel := make(chan error, 1)

	// Init the goroutines
	go func() {
		// Start the Server
		err = server.ListenAndServe()
		if err != nil {
			// Publish the value into the channel
			channel <- fmt.Errorf("Failed to start server: %w", err)
		}
		// Close the channel
		close(channel)
	}()

	// Will block our code's execution until it receives a value of the channel is closed
	// Select -> similar to switch statement, but for channel
	select {
	case err = <-channel:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// If the shutdown exceeds 10 seconds, it will be forcibly stopped.
		return server.Shutdown(timeout)
	}

	return nil
}
