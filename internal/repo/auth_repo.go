package repo

import (
	"context"
	"fmt"
	"goBackend/ent"
	"goBackend/ent/users"
	"goBackend/internal/biz"
	"goBackend/internal/entity"
)

type authRepoImpl struct {
	client *ent.Client
}

func NewAuthRepo(client *ent.Client) biz.AuthRepo {
	return &authRepoImpl{client: client}
}

func (r *authRepoImpl) FindByEmail(ctx context.Context, email string) (*entity.UserProfile, string, error) {
	u, err := r.client.Users.
		Query().
		Where(users.EmailEQ(email)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {

			return nil, "", fmt.Errorf("repo findByEmail: %w", biz.ErrUserNotFound)
		}
		return nil, "", fmt.Errorf("repo findByEmail: unexpected db error: %w", err)
	}

	var hashedPassword string
	if u.Password != nil {
		hashedPassword = *u.Password
	}

	profile := mapEntUserToProfile(u)
	return profile, hashedPassword, nil
}

func (r *authRepoImpl) Save(ctx context.Context, user entity.UserRegister, hashedPassword string) (*entity.UserProfile, error) {
	u, err := r.client.Users.
		Create().
		SetEmail(user.Email).
		SetFullName(user.FullName).
		SetPassword(hashedPassword).
		Save(ctx)

	if err != nil {

		if ent.IsConstraintError(err) {
			return nil, fmt.Errorf("repo save: %w", biz.ErrDuplicateKey)
		}
		return nil, fmt.Errorf("repo save: unexpected db error: %w", err)
	}

	return mapEntUserToProfile(u), nil
}

func (r *authRepoImpl) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	exists, err := r.client.Users.
		Query().
		Where(users.EmailEQ(email)).
		Exist(ctx)

	if err != nil {
		return false, fmt.Errorf("repo existsByEmail: %w", err)
	}

	return exists, nil
}

func mapEntUserToProfile(u *ent.Users) *entity.UserProfile {
	if u == nil {
		return nil
	}
	return &entity.UserProfile{
		Id:        int64(u.ID),
		Email:     u.Email,
		FullName:  u.FullName,
		Avatar:    u.Avatar,
		CreatedAt: &u.CreatedAt,
		UpdatedAt: &u.UpdatedAt,
	}
}
