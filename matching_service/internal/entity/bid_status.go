package entity

const (
	BidStatusPending int8 = iota
	BidStatusMatched
	BidStatusCancelled
)

func IsValidBidStatus(status int8) bool {
	switch status {
	case BidStatusPending, BidStatusMatched, BidStatusCancelled:
		return true
	default:
		return false
	}
}
