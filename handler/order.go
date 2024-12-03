package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/canonflow/build-microservice-with-golang/model"
	"github.com/canonflow/build-microservice-with-golang/repository/order"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type Order struct {
	Repo *order.RedisRepo
}

func (h *Order) Create(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("Creating an order")
	var body struct {
		CustomerID uuid.UUID        `json:"customer_id"`
		LineItems  []model.LineItem `json:"line_items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Init the order
	now := time.Now()
	payload := model.Order{
		CustomerID: body.CustomerID,
		OrderID:    rand.Uint64(),
		LineItems:  body.LineItems,
		CreatedAt:  &now,
	}

	// Insert the order
	err := h.Repo.Insert(r.Context(), payload)
	if err != nil {
		fmt.Println("Failed to insert order:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return the response
	res, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Failed to marshal order:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(res)
	w.WriteHeader(http.StatusCreated)
}

func (h *Order) List(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}

	// Convert cursor str to uint
	const decimal = 10
	const bitSize = 64
	cursor, err := strconv.ParseUint(cursorStr, decimal, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Find all orders
	const size = 50
	res, err := h.Repo.FindAll(r.Context(), order.FindAllPage{
		Offset: cursor,
		Size:   size,
	})
	if err != nil {
		fmt.Println("Failed to find all:", err)
		w.WriteHeader(http.StatusBadRequest)
	}

	// Init response
	var response struct {
		Items []model.Order `json:"items"`
		Next  uint64        `json:"next,omitempty"`
	}
	response.Items = res.Orders
	response.Next = res.Cursor
	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Failed to marshal order:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
	w.WriteHeader(http.StatusOK)
}

func (h *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("Getting a single order by ID")
	idParam := chi.URLParam(r, "id")
	const base = 10
	const bitSize = 64
	orderId, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get order
	o, err := h.Repo.FindByID(r.Context(), orderId)
	if errors.Is(err, order.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("Failed to find order:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send the response
	if err := json.NewEncoder(w).Encode(o); err != nil {
		fmt.Println("Failed to marshal order:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("Updating a single order by ID")
	var body struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	idParam := chi.URLParam(r, "id")
	const base = 10
	const bitSize = 64
	orderId, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Find The Order
	theOrder, err := h.Repo.FindByID(r.Context(), orderId)
	if errors.Is(err, order.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("Failed to find order:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	const completedStatus = "completed"
	const shippedStatus = "shipped"
	now := time.Now()
	switch body.Status {
	case shippedStatus:
		if theOrder.ShippedAt != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		theOrder.ShippedAt = &now
	case completedStatus:
		if theOrder.CompletedAt != nil || theOrder.ShippedAt == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		theOrder.CompletedAt = &now
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.Repo.Update(r.Context(), theOrder)
	if err != nil {
		fmt.Println("Failed to update order:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send the response
	if err := json.NewEncoder(w).Encode(theOrder); err != nil {
		fmt.Println("Failed to marshal order:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("Deleting a single order by ID")
	idParam := chi.URLParam(r, "id")
	const base = 10
	const bitSize = 64
	orderId, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.Repo.DeleteByID(r.Context(), orderId)
	if errors.Is(err, order.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("Failed to delete order:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send response
	var response struct {
		Status string `json:"status"`
	}
	response.Status = "success"
	res, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Failed to marshal order:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(res)
	w.WriteHeader(http.StatusOK)
}
