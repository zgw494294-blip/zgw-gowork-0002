package storage

import (
	"errors"
	"sync"

	"library/internal/domain"
)

type Store struct {
	mu sync.RWMutex

	Books     map[string]domain.Book
	Copies    map[string]domain.Copy
	Readers   map[string]domain.Reader
	Requests  map[string]domain.BorrowRequest
	copyReqID map[string]string // copyID -> requestID (active request)
}

func New() *Store {
	return &Store{
		Books:     make(map[string]domain.Book),
		Copies:    make(map[string]domain.Copy),
		Readers:   make(map[string]domain.Reader),
		Requests:  make(map[string]domain.BorrowRequest),
		copyReqID: make(map[string]string),
	}
}

func (s *Store) CreateBook(b domain.Book) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.Books[b.ID]; exists {
		return errors.New("book already exists")
	}
	s.Books[b.ID] = b
	return nil
}

func (s *Store) CreateCopy(c domain.Copy) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.Copies[c.ID]; exists {
		return errors.New("copy already exists")
	}
	if _, ok := s.Books[c.BookID]; !ok {
		return errors.New("book not found")
	}
	s.Copies[c.ID] = c
	return nil
}

func (s *Store) CreateReader(r domain.Reader) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.Readers[r.ID]; exists {
		return errors.New("reader already exists")
	}
	s.Readers[r.ID] = r
	return nil
}

func (s *Store) CreateRequest(req domain.BorrowRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.Requests[req.ID]; exists {
		return errors.New("request already exists")
	}
	// check copy exists and is available
	copy, ok := s.Copies[req.CopyID]
	if !ok {
		return errors.New("copy not found")
	}
	if copy.Status != domain.CopyAvailable {
		return errors.New("copy is not available")
	}
	// check reader exists
	if _, ok := s.Readers[req.ReaderID]; !ok {
		return errors.New("reader not found")
	}
	// check reader has overdue active request
	for _, r := range s.Requests {
		if r.ReaderID == req.ReaderID && r.Status != domain.RequestReturned {
			if r.DueDate != nil && r.DueDate.Before(req.RequestedAt) {
				return errors.New("reader has overdue request")
			}
		}
	}
	// lock copy immediately (we'll represent availability via copyReqID)
	s.Copies[req.CopyID] = domain.Copy{ID: copy.ID, BookID: copy.BookID, Library: copy.Library, Status: domain.CopyBorrowed}
	s.copyReqID[req.CopyID] = req.ID
	s.Requests[req.ID] = req
	return nil
}

func (s *Store) GetRequest(id string) (domain.BorrowRequest, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.Requests[id]
	return r, ok
}

func (s *Store) UpdateRequest(r domain.BorrowRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Requests[r.ID] = r
}

func (s *Store) GetCopy(id string) (domain.Copy, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.Copies[id]
	return c, ok
}

func (s *Store) ReleaseCopy(copyID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy, ok := s.Copies[copyID]
	if ok {
		copy.Status = domain.CopyAvailable
		s.Copies[copyID] = copy
	}
	delete(s.copyReqID, copyID)
}

func (s *Store) ActiveRequestsByLibraryAndDue() []domain.ActiveRequestView {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var views []domain.ActiveRequestView
	for _, req := range s.Requests {
		if req.Status != domain.RequestShipped && req.Status != domain.RequestReceived {
			continue
		}
		copy := s.Copies[req.CopyID]
		views = append(views, domain.ActiveRequestView{
			RequestID: req.ID,
			CopyID:    req.CopyID,
			ReaderID:  req.ReaderID,
			Library:   copy.Library,
			Status:    req.Status,
			DueDate:   req.DueDate,
		})
	}
	// sort by library then due date
	// we'll use sort.Slice
	return views
}
