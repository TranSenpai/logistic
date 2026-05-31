package entity

const (
	AskStatusPending   string = "PENDING"
	AskStatusMatched   string = "MATCHED"
	AskStatusCancelled string = "CANCELLED"
)

func IsValidAskStatus(status string) bool {
	switch status {
	case AskStatusPending, AskStatusMatched, AskStatusCancelled:
		return true
	default:
		return false
	}
}
