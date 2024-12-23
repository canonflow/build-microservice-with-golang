package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/canonflow/build-microservice-with-golang/model"
	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	Client *redis.Client
}

func orderIDKey(id uint64) string {
	return fmt.Sprintf("order:%d", id)
}

func (r *RedisRepo) Insert(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	key := orderIDKey(order.OrderID)

	// Start Transaction
	txn := r.Client.TxPipeline()

	// Will not allow any data that exists already
	res := txn.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		txn.Discard() // Rollback
		return fmt.Errorf("Failed to set: %w", err)
	}

	// Add OrderId to Orders Set on Redis (will use on FindAll methods)
	if err := txn.SAdd(ctx, "orders", key).Err(); err != nil {
		txn.Discard() // Rollback
		return fmt.Errorf("Failed to add to orders set: %w", err)
	}

	// Commit
	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("Failed to exec: %w", err)
	}
	return nil
}

// Custom Error
var ErrNotExist = errors.New("order does not exist")

func (r *RedisRepo) FindByID(ctx context.Context, id uint64) (model.Order, error) {
	key := orderIDKey(id)

	// Get From Redis
	value, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		// If not exist
		return model.Order{}, ErrNotExist
	} else if err != nil {
		// If Error
		return model.Order{}, fmt.Errorf("get order: %w", err)
	}

	// Convert JSON from Redis to Order Struct
	var order model.Order
	err = json.Unmarshal([]byte(value), &order)
	if err != nil {
		return model.Order{}, fmt.Errorf("failed to decode order json: %w", err)
	}

	return order, nil
}

func (r *RedisRepo) DeleteByID(ctx context.Context, id uint64) error {
	key := orderIDKey(id)

	// Start Transaction
	txn := r.Client.TxPipeline()
	err := txn.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return ErrNotExist
	} else if err != nil {
		txn.Discard() // Rollback
		return fmt.Errorf("delete order: %w", err)
	}

	// Remove OrderID from orders set
	if err := txn.SRem(ctx, "orders", key).Err(); err != nil {
		txn.Discard() // Rollback
		return fmt.Errorf("failed to remove from orders set: %w", err)
	}

	// Commit
	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	return nil
}

func (r *RedisRepo) Update(ctx context.Context, order model.Order) error {
	// Convert Struct to JSON
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to encode order: %w", err)
	}

	key := orderIDKey(order.OrderID)

	// Update the data
	err = r.Client.SetXX(ctx, key, string(data), 0).Err()
	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	} else if err != nil {
		return fmt.Errorf("update order: %w", err)
	}

	return nil
}

type FindAllPage struct {
	Size   uint64
	Offset uint64
}

type FindResult struct {
	Orders []model.Order
	Cursor uint64
}

func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	// Get OrderID from orders set
	res := r.Client.SScan(ctx, "orders", page.Offset, "*", int64(page.Size))

	// Get Keys (OrderID) and cursor
	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get order ids: %w", err)
	}

	// If the keys is empty
	if len(keys) == 0 {
		return FindResult{
			Orders: []model.Order{},
		}, nil
	}

	// Get All Data based on the keys (OrderID)
	data, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get orders: %w", err)
	}

	// Initialize
	orders := make([]model.Order, len(data))

	for i, d := range data {
		d := d.(string) // Convert single data into string
		var order model.Order

		// Parse JSON
		err := json.Unmarshal([]byte(d), &order)
		if err != nil {
			return FindResult{}, fmt.Errorf("failed to decode order json: %w", err)
		}
		orders[i] = order
	}

	return FindResult{
		Orders: orders,
		Cursor: cursor,
	}, nil
}
