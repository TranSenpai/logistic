package repo

import (
	"context"
	"fmt"
	"strconv"

	"matching_service/ent"
	"matching_service/ent/ask"
	"matching_service/ent/bid"
	"matching_service/internal/biz"
	"matching_service/internal/entity"
	"matching_service/internal/mapper"
	"matching_service/internal/mapper/generated"

	"entgo.io/ent/dialect/sql"
)

type matchingRepoImpl struct {
	client *ent.Client
	mapper mapper.Converter
}

func NewMatchingRepo(client *ent.Client) biz.MatchingRepo {
	return &matchingRepoImpl{
		client: client,
		mapper: &generated.ConverterImpl{},
	}
}

func (r *matchingRepoImpl) CreateBid(ctx context.Context, bid *entity.Bid) error {
	userID, _ := strconv.Atoi(bid.UserID)
	pickupPoint := fmt.Sprintf("POINT(%f %f)", bid.Origin.Longitude, bid.Origin.Latitude)
	deliveryPoint := fmt.Sprintf("POINT(%f %f)", bid.Destination.Longitude, bid.Destination.Latitude)

	statusInt := mapper.BidStatusToInt(bid.Status)

	_, err := r.client.Bid.Create().
		SetUserID(userID).
		SetPickupCoordinates(pickupPoint).
		SetDeliveryCoordinates(deliveryPoint).
		SetVolumeM3(bid.VolumeM3).
		SetWeightKg(bid.WeightKg).
		SetMaxPrice(bid.MaxPrice).
		SetStatus(statusInt).
		Save(ctx)

	return err
}

func (r *matchingRepoImpl) CreateAsk(ctx context.Context, ask *entity.Ask) error {
	driverID, _ := strconv.Atoi(ask.DriverID)
	currentPoint := fmt.Sprintf("POINT(%f %f)", ask.CurrentLocation.Longitude, ask.CurrentLocation.Latitude)

	_, err := r.client.Ask.Create().
		SetDriverID(driverID).
		SetCurrentCoordinates(currentPoint).
		SetAvailableVolumeM3(ask.AvailableVolumeM3).
		SetAvailableWeightKg(ask.AvailableWeightKg).
		SetMinPrice(ask.MinPrice).
		SetStatus(mapper.AskStatusToInt(ask.Status)).
		Save(ctx)

	return err
}

func (r *matchingRepoImpl) GetPendingBids(ctx context.Context, zone string) ([]entity.Bid, error) {
	daoList, err := r.client.Bid.Query().
		Where(bid.ZoneID(zone)).
		Where(bid.Status(mapper.BidStatusToInt(entity.BidStatusPending))).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return r.mapper.EntBidListToEntityBidList(daoList), nil
}

func (r *matchingRepoImpl) GetPendingAsks(ctx context.Context, zone string) ([]entity.Ask, error) {
	daoList, err := r.client.Ask.Query().
		Where(ask.ZoneID(zone)).
		Where(ask.Status(mapper.AskStatusToInt(entity.AskStatusPending))).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return r.mapper.EntAskListToEntityAskList(daoList), nil
}

func (r *matchingRepoImpl) FindAskForBid(ctx context.Context, bid *entity.Bid) ([]entity.Ask, error) {
	query := r.client.Ask.Query()

	daoList, err := query.Where(ask.ZoneID(bid.Origin.ZoneID)).
		Where(ask.Status(mapper.AskStatusToInt(entity.AskStatusPending))).
		Where(ask.AvailableVolumeM3GT(bid.VolumeM3)).
		Where(ask.AvailableWeightKgGT(bid.WeightKg)).
		Where(ask.MinPriceLTE(bid.MaxPrice)).
		Where(func(s *sql.Selector) {
			s.Where(sql.ExprP(
				"ST_DWithin(current_coordinates, ST_SetSRID(ST_MakePoint($1, $2), 4326), $3)",
				bid.Origin.Longitude,
				bid.Origin.Latitude,
				5000,
			))
		}).
		Order(ask.ByMinPrice()).
		All(ctx)

	if err != nil {
		return nil, err
	}

	return r.mapper.EntAskListToEntityAskList(daoList), nil
}

func (r *matchingRepoImpl) FindBidForAsk(ctx context.Context, ask *entity.Ask) ([]entity.Bid, error) {
	query := r.client.Bid.Query()

	daoList, err := query.Where(bid.ZoneID(ask.CurrentLocation.ZoneID)).
		Where(bid.Status(mapper.BidStatusToInt(entity.BidStatusPending))).
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
		}).Order(bid.ByMaxPrice()).All(ctx)

	if err != nil {
		return nil, err
	}

	return r.mapper.EntBidListToEntityBidList(daoList), nil
}

func (r *matchingRepoImpl) UpdateAsk(ctx context.Context, ask *entity.Ask) error {
	askID, err := strconv.Atoi(ask.ID)
	if err != nil {
		return fmt.Errorf("invalid ask ID: %w", err)
	}

	currentPoint := fmt.Sprintf("POINT(%f %f)", ask.CurrentLocation.Longitude, ask.CurrentLocation.Latitude)
	_, err = r.client.Ask.UpdateOneID(askID).
		SetCurrentCoordinates(currentPoint).
		SetAvailableVolumeM3(ask.AvailableVolumeM3).
		SetAvailableWeightKg(ask.AvailableWeightKg).
		SetMinPrice(ask.MinPrice).
		SetStatus(mapper.AskStatusToInt(ask.Status)).
		Save(ctx)
	return err
}

func (r *matchingRepoImpl) UpdateBid(ctx context.Context, bid *entity.Bid) error {
	bidID, err := strconv.Atoi(bid.ID)
	if err != nil {
		return fmt.Errorf("invalid bid ID: %w", err)
	}

	pickupPoint := fmt.Sprintf("POINT(%f %f)", bid.Origin.Longitude, bid.Origin.Latitude)
	deliveryPoint := fmt.Sprintf("POINT(%f %f)", bid.Destination.Longitude, bid.Destination.Latitude)

	_, err = r.client.Bid.UpdateOneID(bidID).
		SetPickupCoordinates(pickupPoint).
		SetDeliveryCoordinates(deliveryPoint).
		SetVolumeM3(bid.VolumeM3).
		SetWeightKg(bid.WeightKg).
		SetMaxPrice(bid.MaxPrice).
		SetStatus(mapper.BidStatusToInt(bid.Status)).
		Save(ctx)
	return err
}

func (r *matchingRepoImpl) DeleteBid(ctx context.Context, bidID string) error {
	id, err := strconv.Atoi(bidID)
	if err != nil {
		return fmt.Errorf("invalid bid ID: %w", err)
	}
	return r.client.Bid.DeleteOneID(id).Exec(ctx)
}

func (r *matchingRepoImpl) DeleteAsk(ctx context.Context, askID string) error {
	id, err := strconv.Atoi(askID)
	if err != nil {
		return fmt.Errorf("invalid ask ID: %w", err)
	}
	return r.client.Ask.DeleteOneID(id).Exec(ctx)
}
