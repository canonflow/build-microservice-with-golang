package handler

import (
	"fmt"
	"net/http"
)

type Order struct{}

func (order *Order) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Creating an order")
}

func (order *Order) List(w http.ResponseWriter, r *http.Request) {
	fmt.Println("List all orders")
}

func (order *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Getting a single order by ID")
}

func (order *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Updating a single order by ID")
}

func (order *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Deleting a single order by ID")
}
