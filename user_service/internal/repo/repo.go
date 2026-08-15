package repo

import (
	"context"

	"user_service/ent"
	"user_service/ent/driverprofile"
	"user_service/ent/shipperprofile"
	"user_service/ent/user"
	"github.com/google/uuid"
)

type userRepoImpl struct {
	client *ent.Client
}

func NewUserRepo(client *ent.Client) *userRepoImpl {
	return &userRepoImpl{client: client}
}

func (r *userRepoImpl) CreateUser(ctx context.Context, u *ent.User) (*ent.User, error) {
	return r.client.User.Create().
		SetPhone(u.Phone).
		SetNillableEmail(func() *string {
			if u.Email != "" {
				return &u.Email
			}
			return nil
		}()).
		SetPasswordHash(u.PasswordHash).
		SetRole(user.Role(u.Role)).
		Save(ctx)
}

func (r *userRepoImpl) GetUserByID(ctx context.Context, id uuid.UUID) (*ent.User, error) {
	return r.client.User.Query().Where(user.IDEQ(id)).First(ctx)
}

func (r *userRepoImpl) GetUserByPhone(ctx context.Context, phone string) (*ent.User, error) {
	return r.client.User.Query().Where(user.PhoneEQ(phone)).First(ctx)
}

func (r *userRepoImpl) CreateDriverProfile(ctx context.Context, dp *ent.DriverProfile) error {
	_, err := r.client.DriverProfile.Create().
		SetUserID(dp.Edges.User.ID).
		SetLicenseNumber(dp.LicenseNumber).
		SetIDCard(dp.IDCard).
		Save(ctx)
	return err
}

func (r *userRepoImpl) CreateShipperProfile(ctx context.Context, sp *ent.ShipperProfile) error {
	_, err := r.client.ShipperProfile.Create().
		SetUserID(sp.Edges.User.ID).
		SetNillableCompanyName(func() *string {
			if sp.CompanyName != "" {
				return &sp.CompanyName
			}
			return nil
		}()).
		SetNillableTaxCode(func() *string {
			if sp.TaxCode != "" {
				return &sp.TaxCode
			}
			return nil
		}()).
		Save(ctx)
	return err
}

func (r *userRepoImpl) GetDriverProfile(ctx context.Context, userID uuid.UUID) (*ent.DriverProfile, error) {
	return r.client.DriverProfile.Query().
		Where(driverprofile.HasUserWith(user.IDEQ(userID))).
		First(ctx)
}

func (r *userRepoImpl) GetShipperProfile(ctx context.Context, userID uuid.UUID) (*ent.ShipperProfile, error) {
	return r.client.ShipperProfile.Query().
		Where(shipperprofile.HasUserWith(user.IDEQ(userID))).
		First(ctx)
}

func (r *userRepoImpl) UpdateDriverKYC(ctx context.Context, userID uuid.UUID, status string) error {
	return r.client.DriverProfile.Update().
		Where(driverprofile.HasUserWith(user.IDEQ(userID))).
		SetKycStatus(driverprofile.KycStatus(status)).
		Exec(ctx)
}
