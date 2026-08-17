package service

import (
	"testing"
	"time"

	"library/internal/domain"
	"library/internal/storage"
)

func TestNormalFlow(t *testing.T) {
	store := storage.New()
	svc := New(store)

	// 创建书目、副本、读者
	if err := svc.CreateBook(domain.Book{ID: "b1", Title: "Go Programming", Author: "Alan"}); err != nil {
		t.Fatal(err)
	}
	if err := svc.CreateCopy(domain.Copy{ID: "c1", BookID: "b1", Library: "LibA", Status: domain.CopyAvailable}); err != nil {
		t.Fatal(err)
	}
	if err := svc.CreateReader(domain.Reader{ID: "r1", Name: "Bob", Email: "bob@example.com"}); err != nil {
		t.Fatal(err)
	}

	// 创建申请
	req := domain.BorrowRequest{
		ID:          "req1",
		CopyID:      "c1",
		ReaderID:    "r1",
		Status:      domain.RequestApplied,
		RequestedAt: time.Now(),
	}
	if err := svc.CreateRequest(req); err != nil {
		t.Fatal(err)
	}

	// 状态链：APPLIED -> LOCKED -> SHIPPED -> RECEIVED -> RETURNED
	if err := svc.LockRequest("req1"); err != nil {
		t.Fatal(err)
	}
	if err := svc.ShipRequest("req1"); err != nil {
		t.Fatal(err)
	}
	if err := svc.ReceiveRequest("req1", time.Now().Add(7*24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := svc.ReturnRequest("req1"); err != nil {
		t.Fatal(err)
	}

	// 副本释放
	copy, _ := store.GetCopy("c1")
	if copy.Status != domain.CopyAvailable {
		t.Errorf("copy should be available after return, got %s", copy.Status)
	}
}

func TestFailureKeepsState(t *testing.T) {
	store := storage.New()
	svc := New(store)

	svc.CreateBook(domain.Book{ID: "b1", Title: "T", Author: "A"})
	svc.CreateCopy(domain.Copy{ID: "c1", BookID: "b1", Library: "LibA", Status: domain.CopyAvailable})
	svc.CreateReader(domain.Reader{ID: "r1", Name: "N", Email: "e@e.com"})

	req := domain.BorrowRequest{ID: "req1", CopyID: "c1", ReaderID: "r1", Status: domain.RequestApplied, RequestedAt: time.Now()}
	svc.CreateRequest(req)

	// 尝试从 APPLIED 直接 SHIP 应该失败
	if err := svc.ShipRequest("req1"); err == nil {
		t.Fatal("expected error shipping from APPLIED")
	}
	// 状态不变
	reqGot, _ := store.GetRequest("req1")
	if reqGot.Status != domain.RequestApplied {
		t.Errorf("status changed, got %s", reqGot.Status)
	}
}

func TestOverdueReaderCannotBorrow(t *testing.T) {
	store := storage.New()
	svc := New(store)

	svc.CreateBook(domain.Book{ID: "b1", Title: "T", Author: "A"})
	svc.CreateCopy(domain.Copy{ID: "c1", BookID: "b1", Library: "LibA", Status: domain.CopyAvailable})
	svc.CreateCopy(domain.Copy{ID: "c2", BookID: "b1", Library: "LibA", Status: domain.CopyAvailable})
	svc.CreateReader(domain.Reader{ID: "r1", Name: "N", Email: "e@e.com"})

	// 第一本正常借出，但设置为已逾期
	req1 := domain.BorrowRequest{ID: "req1", CopyID: "c1", ReaderID: "r1", Status: domain.RequestApplied, RequestedAt: time.Now()}
	svc.CreateRequest(req1)
	svc.LockRequest("req1")
	svc.ShipRequest("req1")
	svc.ReceiveRequest("req1", time.Now().Add(-24*time.Hour)) // 已逾期

	// 尝试借第二本，应该失败
	req2 := domain.BorrowRequest{ID: "req2", CopyID: "c2", ReaderID: "r1", Status: domain.RequestApplied, RequestedAt: time.Now()}
	if err := svc.CreateRequest(req2); err == nil {
		t.Fatal("expected error for overdue reader")
	}
}

func TestActiveRequestsQuery(t *testing.T) {
	store := storage.New()
	svc := New(store)

	svc.CreateBook(domain.Book{ID: "b1", Title: "T", Author: "A"})
	svc.CreateCopy(domain.Copy{ID: "c1", BookID: "b1", Library: "LibA", Status: domain.CopyAvailable})
	svc.CreateCopy(domain.Copy{ID: "c2", BookID: "b1", Library: "LibB", Status: domain.CopyAvailable})
	svc.CreateReader(domain.Reader{ID: "r1", Name: "N", Email: "e@e.com"})

	// 创建两个申请，分别进入 SHIPPED 和 RECEIVED
	req1 := domain.BorrowRequest{ID: "req1", CopyID: "c1", ReaderID: "r1", Status: domain.RequestApplied, RequestedAt: time.Now()}
	req2 := domain.BorrowRequest{ID: "req2", CopyID: "c2", ReaderID: "r1", Status: domain.RequestApplied, RequestedAt: time.Now()}
	svc.CreateRequest(req1)
	svc.CreateRequest(req2)
	svc.LockRequest("req1")
	svc.ShipRequest("req1")
	svc.LockRequest("req2")
	svc.ShipRequest("req2")
	svc.ReceiveRequest("req2", time.Now().Add(-48*time.Hour)) // 逾期

	views := svc.ActiveRequests()
	if len(views) != 2 {
		t.Fatalf("expected 2 active requests, got %d", len(views))
	}
	// 按馆排序：LibA, LibB
	if views[0].Library != "LibA" || views[1].Library != "LibB" {
		t.Errorf("not sorted by library: %v", views)
	}
}

func TestDuplicateRequestFailureDoesNotReserveAnotherCopy(t *testing.T) {
	store := storage.New()
	svc := New(store)

	svc.CreateBook(domain.Book{ID: "b1", Title: "T", Author: "A"})
	svc.CreateCopy(domain.Copy{ID: "c1", BookID: "b1", Library: "LibA", Status: domain.CopyAvailable})
	svc.CreateCopy(domain.Copy{ID: "c2", BookID: "b1", Library: "LibA", Status: domain.CopyAvailable})
	svc.CreateReader(domain.Reader{ID: "r1", Name: "N1", Email: "r1@example.com"})
	svc.CreateReader(domain.Reader{ID: "r2", Name: "N2", Email: "r2@example.com"})
	svc.CreateRequest(domain.BorrowRequest{ID: "req1", CopyID: "c1", ReaderID: "r1", Status: domain.RequestApplied, RequestedAt: time.Now()})

	err := svc.CreateRequest(domain.BorrowRequest{ID: "req1", CopyID: "c2", ReaderID: "r2", Status: domain.RequestApplied, RequestedAt: time.Now()})
	if err == nil {
		t.Fatal("expected duplicate request id to fail")
	}
	err = svc.CreateRequest(domain.BorrowRequest{ID: "req2", CopyID: "c2", ReaderID: "r2", Status: domain.RequestApplied, RequestedAt: time.Now()})
	if err != nil {
		t.Fatalf("failed duplicate attempt reserved the unrelated copy: %v", err)
	}
}
