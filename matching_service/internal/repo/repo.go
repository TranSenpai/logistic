package repo

import (
	"context"

	"matching_service/ent"
	"matching_service/ent/asks"
	"matching_service/ent/bids"
	"matching_service/internal/biz"
	"matching_service/internal/entity"
	"matching_service/internal/mapper"
	"matching_service/internal/mapper/generated"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
)

type matchingRepoImpl struct {
	client *ent.Client
	mapper mapper.AppMapper
}

func NewMatchingRepo(client *ent.Client) biz.MatchingRepo {
	return &matchingRepoImpl{
		client: client,
		mapper: &generated.AppMapperImpl{},
	}
}

func (r *matchingRepoImpl) CreateBid(ctx context.Context, bid *entity.Bid) error {
	dao, err := r.client.Bids.Create().
		SetUserID(bid.UserID).
		SetOriginLat(bid.Origin.Latitude).
		SetOriginLng(bid.Origin.Longitude).
		SetDestinationLat(bid.Destination.Latitude).
		SetDestinationLng(bid.Destination.Longitude).
		SetVolumeM3(bid.VolumeM3).
		SetWeightKg(bid.WeightKg).
		SetMaxPrice(bid.MaxPrice).
		SetStatus(bid.Status).
		Save(ctx)

	if err != nil {
		return wrapError(err)
	}
	entity := r.mapper.EntBidToEntityBid(dao)
	*bid = entity

	return nil
}

func (r *matchingRepoImpl) CreateAsk(ctx context.Context, ask *entity.Ask) error {
	dao, err := r.client.Asks.Create().
		SetID(ask.ID).
		SetDriverID(ask.DriverID).
		SetVehicleID(ask.VehicleID).
		SetOriginLat(ask.CurrentLocation.Latitude).
		SetOriginLng(ask.CurrentLocation.Longitude).
		SetDestinationLat(ask.Destination.Latitude).
		SetDestinationLng(ask.Destination.Longitude).
		SetZoneID(ask.CurrentLocation.ZoneID).
		SetAvailableVolumeM3(ask.AvailableVolumeM3).
		SetAvailableWeightKg(ask.AvailableWeightKg).
		SetMinPrice(ask.MinPrice).
		SetStatus(ask.Status).
		Save(ctx)

	if err != nil {
		return wrapError(err)
	}
	entity := r.mapper.EntAskToEntityAsk(dao)
	*ask = entity

	return nil
}

func (r *matchingRepoImpl) GetPendingBids(ctx context.Context, zone string) ([]entity.Bid, error) {
	daoList, err := r.client.Bids.Query().
		Where(bids.ZoneID(zone)).
		Where(bids.Status(entity.BidStatusPending)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return r.mapper.EntBidListToEntityBidList(daoList), nil
}

func (r *matchingRepoImpl) GetPendingAsks(ctx context.Context, zone string) ([]entity.Ask, error) {
	daoList, err := r.client.Asks.Query().
		Where(asks.ZoneID(zone)).
		Where(asks.Status(entity.AskStatusPending)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return r.mapper.EntAskListToEntityAskList(daoList), nil
}

func (r *matchingRepoImpl) FindAskForBid(ctx context.Context, bid *entity.Bid) ([]entity.Ask, error) {
	daoList, err := r.client.Asks.Query().
		Where(asks.ZoneID(bid.Origin.ZoneID)).
		Where(asks.Status(entity.AskStatusPending)).
		Where(asks.AvailableVolumeM3GT(bid.VolumeM3)).
		Where(asks.AvailableWeightKgGT(bid.WeightKg)).
		Where(asks.MinPriceLTE(bid.MaxPrice)).
		Where(func(s *sql.Selector) {
			s.Where(sql.ExprP(
				"ST_DWithin(current_coordinates, ST_SetSRID(ST_MakePoint($1, $2), 4326), $3)",
				bid.Origin.Longitude,
				bid.Origin.Latitude,
				5000,
			))
		}).
		Order(asks.ByMinPrice()).
		All(ctx)

	if err != nil {
		return nil, err
	}

	return r.mapper.EntAskListToEntityAskList(daoList), nil
}

func (r *matchingRepoImpl) FindBidForAsk(ctx context.Context, ask *entity.Ask) ([]entity.Bid, error) {
	daoList, err := r.client.Bids.Query().
		Where(bids.ZoneID(ask.CurrentLocation.ZoneID)).
		Where(bids.Status(entity.BidStatusPending)).
		Where(func(s *sql.Selector) {
			s.Where(sql.ExprP(
				"ST_DWithin(pickup_coordinates, ST_SetSRID(ST_MakePoint($1, $2), 4326), $3)",
				ask.CurrentLocation.Longitude,
				ask.CurrentLocation.Latitude,
				5000,
			))
			s.Where(sql.ExprP(
				"ST_DWithin(delivery_coordinates, ST_SetSRID(ST_MakePoint($1, $2), 4326), $3)",
				ask.Destination.Longitude,
				ask.Destination.Latitude,
				5000,
			))
		}).
		Order(bids.ByMaxPrice()).All(ctx)

	if err != nil {
		return nil, err
	}

	return r.mapper.EntBidListToEntityBidList(daoList), nil
}

func (r *matchingRepoImpl) UpdateAsk(ctx context.Context, ask *entity.Ask) error {
	dao, err := r.client.Asks.
		UpdateOneID(ask.ID).
		SetOriginLat(ask.CurrentLocation.Latitude).
		SetOriginLng(ask.CurrentLocation.Longitude).
		SetAvailableVolumeM3(ask.AvailableVolumeM3).
		SetAvailableWeightKg(ask.AvailableWeightKg).
		SetMinPrice(ask.MinPrice).
		SetStatus(ask.Status).
		Save(ctx)

	if err != nil {
		return wrapError(err)
	}
	entity := r.mapper.EntAskToEntityAsk(dao)
	*ask = entity

	return nil
}

func (r *matchingRepoImpl) UpdateBid(ctx context.Context, bid *entity.Bid) error {
	dao, err := r.client.Bids.
		UpdateOneID(bid.ID).
		SetOriginLat(bid.Origin.Latitude).
		SetOriginLng(bid.Origin.Longitude).
		SetDestinationLat(bid.Destination.Latitude).
		SetDestinationLng(bid.Destination.Longitude).
		SetVolumeM3(bid.VolumeM3).
		SetWeightKg(bid.WeightKg).
		SetMaxPrice(bid.MaxPrice).
		SetStatus(bid.Status).
		Save(ctx)

	if err != nil {
		return wrapError(err)
	}
	entity := r.mapper.EntBidToEntityBid(dao)
	*bid = entity

	return nil
}

func (r *matchingRepoImpl) DeleteBid(ctx context.Context, id uuid.UUID) error {
	return r.client.Bids.DeleteOneID(id).Exec(ctx)
}

func (r *matchingRepoImpl) DeleteAsk(ctx context.Context, id uuid.UUID) error {
	return r.client.Asks.DeleteOneID(id).Exec(ctx)
}
