package biz

// type MatchingEngine struct {
// 	mu sync.RWMutex
// 	bidsByZone map[string]map[string]*entity.Bid
// 	asksByZone map[string]map[string]*entity.Ask

// 	matchChan chan *entity.MatchResult
// }
// func NewMatchingEngine() *MatchingEngine {
// 	return &MatchingEngine{
// 		bidsByZone: make(map[string]map[string]*entity.Bid),
// 		asksByZone: make(map[string]map[string]*entity.Ask),
// 		matchChan:  make(chan *entity.MatchResult, 1000)
// 	}
// }
// func (e *MatchingEngine) SubmitBid(ctx context.Context, bid *entity.Bid) error {
// 	e.mu.Lock()
// 	defer e.mu.Unlock()

// 	zone := bid.Origin.ZoneID
// 	if e.bidsByZone[zone] == nil {
// 		e.bidsByZone[zone] = make(map[string]*entity.Bid)
// 	}

// 	bid.Status = "PENDING"
// 	e.bidsByZone[zone][bid.ID] = bid

// 	e.tryMatch(zone)
// 	return nil
// }
// func (e *MatchingEngine) SubmitAsk(ctx context.Context, ask *entity.Ask) error {
// 	e.mu.Lock()
// 	defer e.mu.Unlock()

// 	zone := ask.CurrentLocation.ZoneID
// 	if e.asksByZone[zone] == nil {
// 		e.asksByZone[zone] = make(map[string]*entity.Ask)
// 	}

// 	ask.Status = "PENDING"
// 	e.asksByZone[zone][ask.ID] = ask

// 	e.tryMatch(zone)
// 	return nil
// }
// func (e *MatchingEngine) tryMatch(zone string) {
// 	bids := e.bidsByZone[zone]
// 	asks := e.asksByZone[zone]

// 	if len(bids) == 0 || len(asks) == 0 {
// 		return
// 	}
// needed.
// 	for bidID, bid := range bids {
// 		if bid.Status != "PENDING" {
// 			continue
// 		}

// 		for askID, ask := range asks {
// 			if ask.Status != "PENDING" {
// 				continue
// 			}

// 			if ask.AvailableVolumeM3 >= bid.VolumeM3 &&
// 				ask.AvailableWeightKg >= bid.WeightKg &&
// 				bid.MaxPrice >= ask.MinPrice {

// 				bid.Status = "MATCHED"
// 				ask.Status = "MATCHED"

// 				settlementPrice := ask.MinPrice

// 				matchResult := &entity.MatchResult{
// 					BidID:     bidID,
// 					AskID:     askID,
// 					Price:     settlementPrice,
// 					MatchedAt: time.Now(),
// 				}

// 				ask.AvailableVolumeM3 -= bid.VolumeM3
// 				ask.AvailableWeightKg -= bid.WeightKg

// 				select {
// 				case e.matchChan <- matchResult:
// 				default:

// 				}

// 				break
// 			}
// 		}
// 	}

// 	for id, b := range bids {
// 		if b.Status == "MATCHED" {
// 			delete(e.bidsByZone[zone], id)
// 		}
// 	}
// 	for id, a := range asks {
// 		if a.Status == "MATCHED" {
// 			delete(e.asksByZone[zone], id)
// 		}
// 	}
// }
// func (e *MatchingEngine) MatchStream() <-chan *entity.MatchResult {
// 	return e.matchChan
// }
