package main

import (
	"log"
	"net/http"

	"library/internal/handler"
	"library/internal/service"
	"library/internal/storage"
)

func main() {
	store := storage.New()
	svc := service.New(store)
	h := handler.New(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /readers", h.CreateReader)
	mux.HandleFunc("POST /books", h.CreateBook)
	mux.HandleFunc("POST /copies", h.CreateCopy)
	mux.HandleFunc("POST /requests", h.CreateRequest)
	mux.HandleFunc("POST /requests/{id}/lock", h.LockRequest)
	mux.HandleFunc("POST /requests/{id}/ship", h.ShipRequest)
	mux.HandleFunc("POST /requests/{id}/receive", h.ReceiveRequest)
	mux.HandleFunc("POST /requests/{id}/return", h.ReturnRequest)
	mux.HandleFunc("GET /active-requests", h.ActiveRequests)

	log.Println("server listening on :8080")
	http.ListenAndServe(":8080", mux)
}
