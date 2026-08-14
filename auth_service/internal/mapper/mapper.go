package mapper

import (
	"time"

	"auth_service/ent"
	"auth_service/internal/entity"
	pb "github.com/logistic/api/logistic/auth_service/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// goverter:converter
// goverter:useZeroValueOnPointerInconsistency
// goverter:ignoreUnexported
//
//go:generate go run github.com/jmattheis/goverter/cmd/goverter@latest gen ./
type AuthMapper interface {
	// goverter:map ID Id | IntToInt64
	// goverter:map CreatedAt CreatedAt | TimeToTimePtr
	// goverter:map UpdatedAt UpdatedAt | TimeToTimePtr
	ToUserProfile(source *ent.Users) *entity.UserProfile

	// goverter:map CreatedAt CreatedAt | TimeToTimestamp
	// goverter:map UpdatedAt UpdatedAt | TimeToTimestamp
	ToUserProfileProto(source *entity.UserProfile) *pb.UserProfile

	ToAuthTokenPairProto(source *entity.AuthTokenPair) *pb.AuthTokenPair
}

func TimeToTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
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
