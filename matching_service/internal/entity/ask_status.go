package entity

const (
	AskStatusPending int8 = iota
	AskStatusMatched
	AskStatusCancelled
)

func IsValidAskStatus(status int8) bool {
	switch status {
	case AskStatusPending, AskStatusMatched, AskStatusCancelled:
		return true
	default:
		return false
	}
}
