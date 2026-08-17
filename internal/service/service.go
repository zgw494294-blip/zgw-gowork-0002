package service

import (
	"errors"
	"sort"
	"time"

	"library/internal/domain"
	"library/internal/storage"
)

type Service struct {
	store *storage.Store
}

func New(store *storage.Store) *Service {
	return &Service{store: store}
}

func (s *Service) CreateBook(b domain.Book) error {
	return s.store.CreateBook(b)
}

func (s *Service) CreateCopy(c domain.Copy) error {
	return s.store.CreateCopy(c)
}

func (s *Service) CreateReader(r domain.Reader) error {
	return s.store.CreateReader(r)
}

func (s *Service) CreateRequest(req domain.BorrowRequest) error {
	if req.Status != domain.RequestApplied {
		return errors.New("new request must be APPLIED")
	}
	return s.store.CreateRequest(req)
}

func (s *Service) LockRequest(requestID string) error {
	req, ok := s.store.GetRequest(requestID)
	if !ok {
		return errors.New("request not found")
	}
	if req.Status != domain.RequestApplied {
		return errors.New("request status must be APPLIED")
	}
	req.Status = domain.RequestLocked
	s.store.UpdateRequest(req)
	return nil
}

func (s *Service) ShipRequest(requestID string) error {
	req, ok := s.store.GetRequest(requestID)
	if !ok {
		return errors.New("request not found")
	}
	if req.Status != domain.RequestLocked {
		return errors.New("request status must be LOCKED")
	}
	req.Status = domain.RequestShipped
	s.store.UpdateRequest(req)
	return nil
}

func (s *Service) ReceiveRequest(requestID string, dueDate time.Time) error {
	req, ok := s.store.GetRequest(requestID)
	if !ok {
		return errors.New("request not found")
	}
	if req.Status != domain.RequestShipped {
		return errors.New("request status must be SHIPPED")
	}
	req.Status = domain.RequestReceived
	req.DueDate = &dueDate
	s.store.UpdateRequest(req)
	return nil
}

func (s *Service) ReturnRequest(requestID string) error {
	req, ok := s.store.GetRequest(requestID)
	if !ok {
		return errors.New("request not found")
	}
	s.store.ReleaseCopy(req.CopyID)
	if req.Status != domain.RequestReceived {
		return errors.New("request status must be RECEIVED")
	}
	req.Status = domain.RequestReturned
	s.store.UpdateRequest(req)
	return nil
}

func (s *Service) ActiveRequests() []domain.ActiveRequestView {
	views := s.store.ActiveRequestsByLibraryAndDue()
	sort.Slice(views, func(i, j int) bool {
		if views[i].Library != views[j].Library {
			return views[i].Library < views[j].Library
		}
		if views[i].DueDate == nil || views[j].DueDate == nil {
			return views[i].DueDate == nil
		}
		return views[i].DueDate.Before(*views[j].DueDate)
	})
	return views
}
