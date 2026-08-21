package biz

import (
	"context"
	"errors"
	"sync"
	"testing"

	"matching_service/internal/entity"

	"github.com/google/uuid"
)

type fakeNotifier struct {
	mu               sync.Mutex
	driverCandidates int
	matchFound       int
	offerReceived    int
	offerRejected    int
	cargoSuggested   int

	lastAsks     []entity.Ask
	lastBids     []entity.Bid
	lastContract *entity.MatchContract
	lastBid      *entity.Bid
	lastAsk      *entity.Ask

	err error
}

func (f *fakeNotifier) NotifyDriverCandidates(_ context.Context, bid *entity.Bid, asks []entity.Ask) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.driverCandidates++
	f.lastBid = bid
	f.lastAsks = asks
	return f.err
}

func (f *fakeNotifier) NotifyMatchFound(_ context.Context, c *entity.MatchContract, bid *entity.Bid, ask *entity.Ask) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.matchFound++
	f.lastContract = c
	f.lastBid = bid
	f.lastAsk = ask
	return f.err
}

func (f *fakeNotifier) NotifyOfferReceived(_ context.Context, bid *entity.Bid, ask *entity.Ask, _ float64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.offerReceived++
	return f.err
}

func (f *fakeNotifier) NotifyOfferRejected(_ context.Context, _ *entity.Bid, _ *entity.Ask, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.offerRejected++
	return f.err
}

func (f *fakeNotifier) NotifyCargoSuggested(_ context.Context, ask *entity.Ask, bids []entity.Bid) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cargoSuggested++
	f.lastAsk = ask
	f.lastBids = bids
	return f.err
}

func (f *fakeNotifier) counts() (int, int, int, int, int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.driverCandidates, f.matchFound, f.offerReceived, f.offerRejected, f.cargoSuggested
}

type fakeRepo struct {
	mu   sync.Mutex
	bids map[uuid.UUID]*entity.Bid
	asks map[uuid.UUID]*entity.Ask

	asksForBid []entity.Ask
	bidsForAsk []entity.Bid
	contracts  []*entity.MatchContract
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		bids: make(map[uuid.UUID]*entity.Bid),
		asks: make(map[uuid.UUID]*entity.Ask),
	}
}

func (r *fakeRepo) CreateBid(_ context.Context, bid *entity.Bid) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	clone := *bid
	r.bids[bid.ID] = &clone
	return nil
}

func (r *fakeRepo) CreateAsk(_ context.Context, ask *entity.Ask) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	clone := *ask
	r.asks[ask.ID] = &clone
	return nil
}

func (r *fakeRepo) FindAskForBid(context.Context, *entity.Bid) ([]entity.Ask, error) {
	return r.asksForBid, nil
}

func (r *fakeRepo) FindBidForAsk(context.Context, *entity.Ask) ([]entity.Bid, error) {
	return r.bidsForAsk, nil
}

func (r *fakeRepo) UpdateBid(_ context.Context, bid *entity.Bid) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	clone := *bid
	r.bids[bid.ID] = &clone
	return nil
}

func (r *fakeRepo) UpdateAsk(_ context.Context, ask *entity.Ask) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	clone := *ask
	r.asks[ask.ID] = &clone
	return nil
}

func (r *fakeRepo) DeleteAsk(context.Context, uuid.UUID) error { return nil }
func (r *fakeRepo) DeleteBid(context.Context, uuid.UUID) error { return nil }

func (r *fakeRepo) GetBid(_ context.Context, id uuid.UUID) (*entity.Bid, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if b, ok := r.bids[id]; ok {
		clone := *b
		return &clone, nil
	}
	return nil, errors.New("bid not found")
}

func (r *fakeRepo) GetAsk(_ context.Context, id uuid.UUID) (*entity.Ask, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if a, ok := r.asks[id]; ok {
		clone := *a
		return &clone, nil
	}
	return nil, errors.New("ask not found")
}

func (r *fakeRepo) CreateMatchContract(_ context.Context, c *entity.MatchContract) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.contracts = append(r.contracts, c)
	return nil
}

type fakeSpatial struct{}

func (fakeSpatial) GetZoneId(context.Context, float64, float64) (string, error) {
	return "VN-HCM-Q1", nil
}
func (fakeSpatial) GetNeighborZones(context.Context, string) ([]string, error) {
	return []string{"VN-HCM-Q1"}, nil
}

type fakePublisher struct {
	mu    sync.Mutex
	count int
}

func (p *fakePublisher) Publish(context.Context, *EventMessage) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.count++
	return nil
}

func newTestEngine(repo MatchingRepo, notifier Notifier) MatchingEngine {
	return NewMatchingEngine(repo, fakeSpatial{}, NewMockWalletClient(), &fakePublisher{}, &fakePublisher{}, notifier)
}

func TestSubmitBidNotifiesCandidateDrivers(t *testing.T) {
	repo := newFakeRepo()
	repo.asksForBid = []entity.Ask{
		availableTruck(100),
		availableTruck(120),
	}

	notifier := &fakeNotifier{}
	engine := newTestEngine(repo, notifier)

	bid := &entity.Bid{
		ShipperID: uuid.New(),
		Origin:    entity.Location{Latitude: 10.7769, Longitude: 106.7009},
		WeightKg:  500,
		VolumeM3:  3,
		MaxPrice:  200,
	}

	if _, err := engine.SubmitBid(context.Background(), bid); err != nil {
		t.Fatalf("SubmitBid lỗi: %v", err)
	}

	candidates, _, _, _, _ := notifier.counts()
	if candidates != 1 {
		t.Fatalf("mong đợi 1 lần báo tài xế tiềm năng, nhận %d", candidates)
	}
	if len(notifier.lastAsks) != 2 {
		t.Errorf("danh sách tài xế gửi đi có %d mục, mong đợi 2", len(notifier.lastAsks))
	}
	if notifier.lastBid == nil || notifier.lastBid.ID == uuid.Nil {
		t.Error("bid gửi cho notifier phải có ID đã sinh")
	}
}

func TestSubmitBidWithNoCandidatesDoesNotNotify(t *testing.T) {
	repo := newFakeRepo()
	repo.asksForBid = nil

	notifier := &fakeNotifier{}
	engine := newTestEngine(repo, notifier)

	_, err := engine.SubmitBid(context.Background(), &entity.Bid{
		ShipperID: uuid.New(),
		Origin:    entity.Location{Latitude: 10.7769, Longitude: 106.7009},
	})
	if err != nil {
		t.Fatalf("SubmitBid lỗi: %v", err)
	}

	if candidates, _, _, _, _ := notifier.counts(); candidates != 0 {
		t.Errorf("không có tài xế nào mà vẫn gọi notifier %d lần", candidates)
	}
}

func TestSubmitBidSucceedsWhenNotifierFails(t *testing.T) {
	repo := newFakeRepo()
	repo.asksForBid = []entity.Ask{availableTruck(100)}

	notifier := &fakeNotifier{err: errors.New("rabbitmq: connection refused")}
	engine := newTestEngine(repo, notifier)

	bid, err := engine.SubmitBid(context.Background(), &entity.Bid{
		ShipperID: uuid.New(),
		Origin:    entity.Location{Latitude: 10.7769, Longitude: 106.7009},
	})
	if err != nil {
		t.Fatalf("notifier hỏng không được làm SubmitBid thất bại, nhận: %v", err)
	}
	if bid == nil || bid.ID == uuid.Nil {
		t.Fatal("bid vẫn phải được tạo dù notifier hỏng")
	}
	if _, ok := repo.bids[bid.ID]; !ok {
		t.Error("bid phải nằm trong repo dù notifier hỏng")
	}
}

func TestAcceptOfferNotifiesMatchFound(t *testing.T) {
	repo := newFakeRepo()
	notifier := &fakeNotifier{}
	engine := newTestEngine(repo, notifier)
	ctx := context.Background()

	bidID := uuid.Must(uuid.NewV7())
	askID := uuid.Must(uuid.NewV7())

	repo.bids[bidID] = &entity.Bid{ID: bidID, ShipperID: uuid.New(), Status: entity.BidStatusNegotiating}
	repo.asks[askID] = &entity.Ask{ID: askID, DriverID: uuid.New(), VehicleID: uuid.New(), MinPrice: 1000}

	contract, err := engine.AcceptOffer(ctx, bidID, askID)
	if err != nil {
		t.Fatalf("AcceptOffer lỗi: %v", err)
	}

	_, matchFound, _, _, _ := notifier.counts()
	if matchFound != 1 {
		t.Fatalf("mong đợi 1 lần báo ghép đơn, nhận %d", matchFound)
	}
	if notifier.lastContract == nil || notifier.lastContract.ID != contract.ID {
		t.Error("contract gửi cho notifier không khớp contract trả về")
	}

	if notifier.lastBid == nil || notifier.lastAsk == nil {
		t.Error("notifier phải nhận được cả bid và ask để báo cho hai phía")
	}
}

func TestAcceptOfferRejectsNonNegotiatingBid(t *testing.T) {
	repo := newFakeRepo()
	notifier := &fakeNotifier{}
	engine := newTestEngine(repo, notifier)

	bidID := uuid.Must(uuid.NewV7())
	askID := uuid.Must(uuid.NewV7())
	repo.bids[bidID] = &entity.Bid{ID: bidID, Status: entity.BidStatusPending}
	repo.asks[askID] = &entity.Ask{ID: askID}

	if _, err := engine.AcceptOffer(context.Background(), bidID, askID); err == nil {
		t.Fatal("không được chốt một bid chưa qua bước thương lượng")
	}
	if _, matchFound, _, _, _ := notifier.counts(); matchFound != 0 {
		t.Errorf("chốt thất bại mà vẫn báo ghép đơn %d lần", matchFound)
	}
}

func TestSubmitAskNotifiesCargoSuggested(t *testing.T) {
	repo := newFakeRepo()
	repo.bidsForAsk = []entity.Bid{
		matchableCargo(),
		matchableCargo(),
		matchableCargo(),
	}

	notifier := &fakeNotifier{}
	engine := newTestEngine(repo, notifier)

	truck := availableTruck(1_000_000)
	_, err := engine.SubmitAsk(context.Background(), &truck)
	if err != nil {
		t.Fatalf("SubmitAsk lỗi: %v", err)
	}

	_, _, _, _, cargo := notifier.counts()
	if cargo != 1 {
		t.Fatalf("mong đợi 1 lần gợi ý đơn hàng, nhận %d", cargo)
	}
	if len(notifier.lastBids) != 3 {
		t.Errorf("gợi ý %d đơn, mong đợi 3", len(notifier.lastBids))
	}
}

func TestRejectOfferNotifiesDriverAndReopensBid(t *testing.T) {
	repo := newFakeRepo()
	notifier := &fakeNotifier{}
	engine := newTestEngine(repo, notifier)

	bidID := uuid.Must(uuid.NewV7())
	askID := uuid.Must(uuid.NewV7())
	repo.bids[bidID] = &entity.Bid{ID: bidID, ShipperID: uuid.New(), Status: entity.BidStatusNegotiating}
	repo.asks[askID] = &entity.Ask{ID: askID, DriverID: uuid.New()}

	if err := engine.RejectOffer(context.Background(), bidID, askID); err != nil {
		t.Fatalf("RejectOffer lỗi: %v", err)
	}

	if _, _, _, rejected, _ := notifier.counts(); rejected != 1 {
		t.Errorf("mong đợi 1 lần báo từ chối, nhận %d", rejected)
	}

	if repo.bids[bidID].Status != entity.BidStatusPending {
		t.Errorf("bid sau khi từ chối có status %d, mong đợi PENDING (%d)",
			repo.bids[bidID].Status, entity.BidStatusPending)
	}
}

func TestRejectOfferWithMissingAskDoesNotPanic(t *testing.T) {
	repo := newFakeRepo()
	notifier := &fakeNotifier{}
	engine := newTestEngine(repo, notifier)

	bidID := uuid.Must(uuid.NewV7())
	repo.bids[bidID] = &entity.Bid{ID: bidID, ShipperID: uuid.New(), Status: entity.BidStatusNegotiating}

	err := engine.RejectOffer(context.Background(), bidID, uuid.Must(uuid.NewV7()))
	if err != nil {
		t.Fatalf("không đọc được ask không phải lỗi chí mạng, nhận: %v", err)
	}
	if repo.bids[bidID].Status != entity.BidStatusPending {
		t.Error("bid vẫn phải được mở lại kể cả khi không đọc được ask")
	}
}

func TestNewMatchingEngineAcceptsNilNotifier(t *testing.T) {
	repo := newFakeRepo()

	engine := NewMatchingEngine(repo, fakeSpatial{}, NewMockWalletClient(), &fakePublisher{}, &fakePublisher{}, nil)

	repo.asksForBid = []entity.Ask{availableTruck(100)}
	if _, err := engine.SubmitBid(context.Background(), &entity.Bid{
		ShipperID: uuid.New(),
		Origin:    entity.Location{Latitude: 10.7769, Longitude: 106.7009},
	}); err != nil {
		t.Fatalf("engine với notifier nil phải chạy được, nhận: %v", err)
	}
}

func availableTruck(minPrice float64) entity.Ask {
	return entity.Ask{
		ID:                uuid.New(),
		DriverID:          uuid.New(),
		VehicleID:         uuid.New(),
		CurrentLocation:   entity.Location{Latitude: 10.7769, Longitude: 106.7009},
		Destination:       entity.Location{Latitude: 10.9804, Longitude: 106.6519},
		AvailableWeightKg: 8000,
		AvailableVolumeM3: 30,
		MinPrice:          minPrice,
	}
}

func matchableCargo() entity.Bid {
	return entity.Bid{
		ID:          uuid.New(),
		ShipperID:   uuid.New(),
		Origin:      entity.Location{Latitude: 10.7769, Longitude: 106.7009},
		Destination: entity.Location{Latitude: 10.9804, Longitude: 106.6519},
		WeightKg:    2000,
		VolumeM3:    8,
		MaxPrice:    5_000_000,
	}
}
