package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"library/internal/domain"
	"library/internal/service"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func pathID(r *http.Request) string {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

func (h *Handler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     string `json:"id"`
		Title  string `json:"title"`
		Author string `json:"author"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.svc.CreateBook(domain.Book{ID: req.ID, Title: req.Title, Author: req.Author}); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) CreateCopy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID      string `json:"id"`
		BookID  string `json:"book_id"`
		Library string `json:"library"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	copy := domain.Copy{ID: req.ID, BookID: req.BookID, Library: req.Library, Status: domain.CopyAvailable}
	if err := h.svc.CreateCopy(copy); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) CreateReader(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.svc.CreateReader(domain.Reader{ID: req.ID, Name: req.Name, Email: req.Email}); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID       string `json:"id"`
		CopyID   string `json:"copy_id"`
		ReaderID string `json:"reader_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	borrowReq := domain.BorrowRequest{
		ID:          req.ID,
		CopyID:      req.CopyID,
		ReaderID:    req.ReaderID,
		Status:      domain.RequestApplied,
		RequestedAt: time.Now(),
	}
	if err := h.svc.CreateRequest(borrowReq); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) LockRequest(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if err := h.svc.LockRequest(id); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ShipRequest(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if err := h.svc.ShipRequest(id); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ReceiveRequest(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var req struct {
		DueDate string `json:"due_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	due, err := time.Parse(time.RFC3339, req.DueDate)
	if err != nil {
		http.Error(w, "invalid due_date", http.StatusBadRequest)
		return
	}
	if err := h.svc.ReceiveRequest(id, due); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ReturnRequest(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if err := h.svc.ReturnRequest(id); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ActiveRequests(w http.ResponseWriter, r *http.Request) {
	views := h.svc.ActiveRequests()
	writeJSON(w, views)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
