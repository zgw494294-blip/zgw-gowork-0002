package domain

import "time"

type Book struct {
	ID    string
	Title string
	Author string
}

type Copy struct {
	ID       string
	BookID   string
	Library  string
	Status   CopyStatus
}

type CopyStatus string

const (
	CopyAvailable CopyStatus = "AVAILABLE"
	CopyBorrowed  CopyStatus = "BORROWED"
)

type Reader struct {
	ID    string
	Name  string
	Email string
}

type BorrowRequestStatus string

const (
	RequestApplied  BorrowRequestStatus = "APPLIED"
	RequestLocked   BorrowRequestStatus = "LOCKED"
	RequestShipped  BorrowRequestStatus = "SHIPPED"
	RequestReceived BorrowRequestStatus = "RECEIVED"
	RequestReturned BorrowRequestStatus = "RETURNED"
)

type BorrowRequest struct {
	ID          string
	CopyID      string
	ReaderID    string
	Status      BorrowRequestStatus
	RequestedAt time.Time
	DueDate     *time.Time
}

func (r *BorrowRequest) IsActive() bool {
	return r.Status != RequestReturned
}

type ActiveRequestView struct {
	RequestID string
	CopyID    string
	ReaderID  string
	Library   string
	Status    BorrowRequestStatus
	DueDate   *time.Time
}
