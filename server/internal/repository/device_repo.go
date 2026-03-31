package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"petverse/server/internal/model"
	"petverse/server/internal/pkg/pagination"
)

type DeviceRepository struct {
	db     *gorm.DB
	dataDB *gorm.DB
}

func NewDeviceRepository(db *gorm.DB) *DeviceRepository {
	return &DeviceRepository{db: db, dataDB: db}
}

func NewDeviceRepositoryWithTimeseries(db, dataDB *gorm.DB) *DeviceRepository {
	if dataDB == nil {
		dataDB = db
	}
	return &DeviceRepository{db: db, dataDB: dataDB}
}

func (r *DeviceRepository) CreateDevice(ctx context.Context, device *model.Device) error {
	return r.db.WithContext(ctx).Create(device).Error
}

func (r *DeviceRepository) ListDevicesByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) ([]model.Device, int64, error) {
	var (
		devices []model.Device
		total   int64
	)

	query := r.db.WithContext(ctx).Model(&model.Device{}).Where("owner_id = ?", ownerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("created_at DESC").
		Limit(pageSize).
		Offset(pagination.Offset(page, pageSize)).
		Find(&devices).Error
	if err != nil {
		return nil, 0, err
	}

	return devices, total, nil
}

func (r *DeviceRepository) GetDeviceByIDAndOwner(ctx context.Context, id, ownerID uuid.UUID) (*model.Device, error) {
	var device model.Device
	err := r.db.WithContext(ctx).Where("id = ? AND owner_id = ?", id, ownerID).First(&device).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *DeviceRepository) UpdateDevice(ctx context.Context, device *model.Device) error {
	return r.db.WithContext(ctx).Save(device).Error
}

func (r *DeviceRepository) DeleteDevice(ctx context.Context, device *model.Device) error {
	return r.db.WithContext(ctx).Delete(device).Error
}

func (r *DeviceRepository) CreateDataPoints(ctx context.Context, points []model.DeviceDataPoint) error {
	if len(points) == 0 {
		return nil
	}
	return r.dataStore(ctx).Create(&points).Error
}

func (r *DeviceRepository) ListDataPoints(ctx context.Context, deviceID uuid.UUID, metric string, since *time.Time, limit int) ([]model.DeviceDataPoint, error) {
	var points []model.DeviceDataPoint
	query := r.dataStore(ctx).Where("device_id = ?", deviceID)
	if metric != "" {
		query = query.Where("metric = ?", metric)
	}
	if since != nil {
		query = query.Where("time >= ?", *since)
	}

	err := query.
		Order("time DESC").
		Limit(limit).
		Find(&points).Error
	if err != nil {
		return nil, err
	}
	return points, nil
}

func (r *DeviceRepository) LatestDataPoints(ctx context.Context, deviceID uuid.UUID, limit int) ([]model.DeviceDataPoint, error) {
	return r.ListDataPoints(ctx, deviceID, "", nil, limit)
}

func (r *DeviceRepository) ListDataPointsByPet(ctx context.Context, petID uuid.UUID, limit int) ([]model.DeviceDataPoint, error) {
	var points []model.DeviceDataPoint
	if r.dataDB == nil || r.dataDB == r.db {
		err := r.db.WithContext(ctx).
			Table("device_data_points").
			Select("device_data_points.*").
			Joins("JOIN devices ON devices.id = device_data_points.device_id").
			Where("devices.pet_id = ?", petID).
			Order("device_data_points.time DESC").
			Limit(limit).
			Scan(&points).Error
		if err != nil {
			return nil, err
		}
		return points, nil
	}

	var deviceIDs []uuid.UUID
	if err := r.db.WithContext(ctx).
		Model(&model.Device{}).
		Where("pet_id = ?", petID).
		Pluck("id", &deviceIDs).Error; err != nil {
		return nil, err
	}
	if len(deviceIDs) == 0 {
		return []model.DeviceDataPoint{}, nil
	}

	err := r.dataStore(ctx).
		Where("device_id IN ?", deviceIDs).
		Order("time DESC").
		Limit(limit).
		Find(&points).Error
	if err != nil {
		return nil, err
	}
	return points, nil
}

func (r *DeviceRepository) dataStore(ctx context.Context) *gorm.DB {
	if r.dataDB != nil {
		return r.dataDB.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}
