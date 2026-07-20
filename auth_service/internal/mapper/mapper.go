package mapper

import (
	"time"

	"auth_service/internal/entity"
	dto "goBackend/api/logistics/v1/gen"
	"auth_service/ent"
)

// goverter:converter
// goverter:useZeroValueOnPointerInconsistency
type AuthMapper interface {
	// goverter:map ID Id | IntToInt64
	// goverter:map CreatedAt CreatedAt | TimeToTimePtr
	// goverter:map UpdatedAt UpdatedAt | TimeToTimePtr
	ToUserProfile(source *ent.Users) *entity.UserProfile

	// goverter:map CreatedAt CreatedAt | TimeToUnixPtr
	ToUserProfileResponse(source *entity.UserProfile) dto.UserProfileResponse
}

func TimeToUnixPtr(t *time.Time) *int64 {
	if t == nil {
		return nil
	}
	val := t.Unix()
	return &val
}

func TimeToTimePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func IntToInt64(i int) int64 {
	return int64(i)
}

// TODO: Thêm LogisticsMapper sau khi có entity và dto cho Logistics
