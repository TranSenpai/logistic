package entity

const (
	BidStatusPending   string = "PENDING"
	BidStatusMatched   string = "MATCHED"
	BidStatusCancelled string = "CANCELLED"
)

func IsValidBidStatus(status string) bool {
	switch status {
	case BidStatusPending, BidStatusMatched, BidStatusCancelled:
		return true
	default:
		return false
	}
}
