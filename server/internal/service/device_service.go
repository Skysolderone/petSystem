package service

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"petverse/server/internal/dto"
	"petverse/server/internal/model"
	"petverse/server/internal/pkg/apperror"
	"petverse/server/internal/pkg/events"
	"petverse/server/internal/pkg/pagination"
	"petverse/server/internal/ws"
)

type DeviceService struct {
	devices deviceRepository
	pets    devicePetRepository
	events  events.Publisher
	hub     *ws.Hub
}

type deviceRepository interface {
	CreateDevice(ctx context.Context, device *model.Device) error
	ListDevicesByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) ([]model.Device, int64, error)
	GetDeviceByIDAndOwner(ctx context.Context, id, ownerID uuid.UUID) (*model.Device, error)
	UpdateDevice(ctx context.Context, device *model.Device) error
	DeleteDevice(ctx context.Context, device *model.Device) error
	CreateDataPoints(ctx context.Context, points []model.DeviceDataPoint) error
	ListDataPoints(ctx context.Context, deviceID uuid.UUID, metric string, since *time.Time, limit int) ([]model.DeviceDataPoint, error)
	LatestDataPoints(ctx context.Context, deviceID uuid.UUID, limit int) ([]model.DeviceDataPoint, error)
}

type devicePetRepository interface {
	GetByIDAndOwner(ctx context.Context, id, ownerID uuid.UUID) (*model.Pet, error)
}

type DeviceServiceOption func(*DeviceService)

func WithDeviceEvents(publisher events.Publisher) DeviceServiceOption {
	return func(service *DeviceService) {
		service.events = publisher
	}
}

func NewDeviceService(devices deviceRepository, pets devicePetRepository, hub *ws.Hub, options ...DeviceServiceOption) *DeviceService {
	service := &DeviceService{
		devices: devices,
		pets:    pets,
		hub:     hub,
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func (s *DeviceService) List(ctx context.Context, ownerID uuid.UUID, page, pageSize int) ([]model.Device, int64, int, int, error) {
	page, pageSize = pagination.Normalize(page, pageSize)
	devices, total, err := s.devices.ListDevicesByOwner(ctx, ownerID, page, pageSize)
	if err != nil {
		return nil, 0, 0, 0, apperror.Wrap(http.StatusInternalServerError, "list_devices_failed", "failed to load devices", err)
	}
	return devices, total, page, pageSize, nil
}

func (s *DeviceService) Create(ctx context.Context, ownerID uuid.UUID, req dto.CreateDeviceRequest) (*model.Device, error) {
	var petID *uuid.UUID
	if req.PetID != nil && *req.PetID != "" {
		parsedPetID, err := uuid.Parse(*req.PetID)
		if err != nil {
			return nil, apperror.New(http.StatusBadRequest, "invalid_pet_id", "pet id is invalid")
		}
		if _, err := s.pets.GetByIDAndOwner(ctx, parsedPetID, ownerID); err != nil {
			return nil, apperror.Wrap(http.StatusInternalServerError, "pet_lookup_failed", "failed to load pet", err)
		}
		petID = &parsedPetID
	}

	status := req.Status
	if status == "" {
		status = "online"
	}

	now := time.Now()
	device := &model.Device{
		OwnerID:      ownerID,
		PetID:        petID,
		DeviceType:   req.DeviceType,
		Brand:        req.Brand,
		Model:        req.Model,
		Nickname:     req.Nickname,
		SerialNumber: req.SerialNumber,
		Status:       status,
		Config:       datatypes.JSON(dto.EncodeMap(req.Config)),
		LastSeen:     &now,
		FirmwareVer:  req.FirmwareVer,
		BatteryLevel: defaultBattery(req.BatteryLevel),
	}

	if err := s.devices.CreateDevice(ctx, device); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "create_device_failed", "failed to create device", err)
	}

	seedPoints := seedDataPoints(device)
	if err := s.devices.CreateDataPoints(ctx, seedPoints); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "seed_device_data_failed", "failed to seed device data", err)
	}

	s.publish(ctx, "device.created", map[string]any{
		"device_id":      device.ID.String(),
		"owner_id":       device.OwnerID.String(),
		"pet_id":         nullableUUID(device.PetID),
		"device_type":    device.DeviceType,
		"serial_number":  device.SerialNumber,
		"battery_level":  device.BatteryLevel,
		"seed_data_size": len(seedPoints),
	})

	return device, nil
}

func (s *DeviceService) Get(ctx context.Context, ownerID, deviceID uuid.UUID) (*model.Device, error) {
	device, err := s.devices.GetDeviceByIDAndOwner(ctx, deviceID, ownerID)
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "get_device_failed", "failed to load device", err)
	}
	if device == nil {
		return nil, apperror.New(http.StatusNotFound, "device_not_found", "device not found")
	}
	return device, nil
}

func (s *DeviceService) Update(ctx context.Context, ownerID, deviceID uuid.UUID, req dto.UpdateDeviceRequest) (*model.Device, error) {
	device, err := s.Get(ctx, ownerID, deviceID)
	if err != nil {
		return nil, err
	}

	if req.PetID != nil {
		if *req.PetID == "" {
			device.PetID = nil
		} else {
			parsedPetID, err := uuid.Parse(*req.PetID)
			if err != nil {
				return nil, apperror.New(http.StatusBadRequest, "invalid_pet_id", "pet id is invalid")
			}
			if _, err := s.pets.GetByIDAndOwner(ctx, parsedPetID, ownerID); err != nil {
				return nil, apperror.Wrap(http.StatusInternalServerError, "pet_lookup_failed", "failed to load pet", err)
			}
			device.PetID = &parsedPetID
		}
	}
	if req.DeviceType != nil {
		device.DeviceType = *req.DeviceType
	}
	if req.Brand != nil {
		device.Brand = *req.Brand
	}
	if req.Model != nil {
		device.Model = *req.Model
	}
	if req.Nickname != nil {
		device.Nickname = *req.Nickname
	}
	if req.Status != nil {
		device.Status = *req.Status
	}
	if req.Config != nil {
		device.Config = datatypes.JSON(dto.EncodeMap(req.Config))
	}
	if req.FirmwareVer != nil {
		device.FirmwareVer = *req.FirmwareVer
	}
	if req.BatteryLevel != nil {
		device.BatteryLevel = req.BatteryLevel
	}
	now := time.Now()
	device.LastSeen = &now

	if err := s.devices.UpdateDevice(ctx, device); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "update_device_failed", "failed to update device", err)
	}

	return device, nil
}

func (s *DeviceService) Delete(ctx context.Context, ownerID, deviceID uuid.UUID) error {
	device, err := s.Get(ctx, ownerID, deviceID)
	if err != nil {
		return err
	}
	if err := s.devices.DeleteDevice(ctx, device); err != nil {
		return apperror.Wrap(http.StatusInternalServerError, "delete_device_failed", "failed to delete device", err)
	}
	return nil
}

func (s *DeviceService) Command(ctx context.Context, ownerID, deviceID uuid.UUID, req dto.DeviceCommandRequest) (*model.DeviceDataPoint, error) {
	device, err := s.Get(ctx, ownerID, deviceID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	device.Status = "online"
	device.LastSeen = &now
	if device.BatteryLevel != nil {
		level := max(*device.BatteryLevel-1, 5)
		device.BatteryLevel = &level
	}

	point := model.DeviceDataPoint{
		DeviceID: device.ID,
		Time:     now,
		Metric:   metricForCommand(device.DeviceType, req.Command),
		Value:    valueForCommand(req.Command, req.Params),
		Unit:     unitForMetric(metricForCommand(device.DeviceType, req.Command)),
		Meta:     datatypes.JSON(dto.EncodeMap(req.Params)),
	}

	if err := s.devices.UpdateDevice(ctx, device); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "update_device_failed", "failed to update device", err)
	}
	if err := s.devices.CreateDataPoints(ctx, []model.DeviceDataPoint{point}); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "create_device_data_failed", "failed to create device data", err)
	}

	s.publish(ctx, "device.data_point.created", map[string]any{
		"device_id": device.ID.String(),
		"owner_id":  ownerID.String(),
		"pet_id":    nullableUUID(device.PetID),
		"metric":    point.Metric,
		"value":     point.Value,
		"unit":      point.Unit,
		"time":      point.Time,
		"command":   req.Command,
	})

	s.broadcastStatus(ctx, ownerID, device.ID)
	return &point, nil
}

func (s *DeviceService) Data(ctx context.Context, ownerID, deviceID uuid.UUID, metric string, hours, limit int) ([]model.DeviceDataPoint, error) {
	if _, err := s.Get(ctx, ownerID, deviceID); err != nil {
		return nil, err
	}

	var since *time.Time
	if hours > 0 {
		value := time.Now().Add(-time.Duration(hours) * time.Hour)
		since = &value
	}
	if limit <= 0 {
		limit = 50
	}

	points, err := s.devices.ListDataPoints(ctx, deviceID, metric, since, limit)
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "list_device_data_failed", "failed to load device data", err)
	}
	return points, nil
}

func (s *DeviceService) Status(ctx context.Context, ownerID, deviceID uuid.UUID) (*model.Device, []model.DeviceDataPoint, error) {
	device, err := s.Get(ctx, ownerID, deviceID)
	if err != nil {
		return nil, nil, err
	}

	points, err := s.devices.LatestDataPoints(ctx, deviceID, 6)
	if err != nil {
		return nil, nil, apperror.Wrap(http.StatusInternalServerError, "list_device_data_failed", "failed to load device data", err)
	}
	return device, points, nil
}

func (s *DeviceService) broadcastStatus(ctx context.Context, ownerID, deviceID uuid.UUID) {
	if s.hub == nil {
		return
	}

	device, points, err := s.Status(ctx, ownerID, deviceID)
	if err != nil {
		return
	}

	resolvedStatus := dto.DeviceStatusResponse{
		Device:           dto.ToDeviceResponse(device),
		LatestDataPoints: make([]dto.DeviceDataPointResponse, 0, len(points)),
	}
	for _, point := range points {
		pointCopy := point
		resolvedStatus.LatestDataPoints = append(resolvedStatus.LatestDataPoints, dto.ToDeviceDataPointResponse(&pointCopy))
	}

	s.hub.Broadcast(deviceID, dto.DeviceStreamMessage{
		Type:      "device_status",
		Timestamp: time.Now(),
		Status:    &resolvedStatus,
	})
}

func defaultBattery(raw *int) *int {
	if raw != nil {
		return raw
	}
	level := 84
	return &level
}

func seedDataPoints(device *model.Device) []model.DeviceDataPoint {
	metric := metricForDeviceType(device.DeviceType)
	unit := unitForMetric(metric)
	now := time.Now()
	values := []float64{36, 42, 38, 44, 41, 39}

	switch device.DeviceType {
	case "water_fountain":
		values = []float64{120, 100, 95, 105, 115, 98}
	case "camera":
		values = []float64{1, 0, 2, 1, 3, 1}
	case "gps_collar":
		values = []float64{78, 62, 85, 90, 74, 68}
	}

	points := make([]model.DeviceDataPoint, 0, len(values))
	for index, value := range values {
		points = append(points, model.DeviceDataPoint{
			DeviceID: device.ID,
			Time:     now.Add(-time.Duration(index+1) * 30 * time.Minute),
			Metric:   metric,
			Value:    value,
			Unit:     unit,
			Meta:     datatypes.JSON(dto.EncodeMap(map[string]any{"seeded": true})),
		})
	}
	return points
}

func metricForDeviceType(deviceType string) string {
	switch deviceType {
	case "water_fountain":
		return "water_intake"
	case "camera":
		return "motion"
	case "gps_collar":
		return "motion"
	default:
		return "feeding_amount"
	}
}

func metricForCommand(deviceType, command string) string {
	switch command {
	case "feed_now":
		return "feeding_amount"
	case "refresh_water":
		return "water_intake"
	case "snapshot":
		return "motion"
	default:
		return metricForDeviceType(deviceType)
	}
}

func unitForMetric(metric string) string {
	switch metric {
	case "water_intake":
		return "ml"
	case "feeding_amount":
		return "g"
	case "motion":
		return "score"
	default:
		return "count"
	}
}

func valueForCommand(command string, params map[string]any) float64 {
	if value, ok := params["value"]; ok {
		switch typed := value.(type) {
		case float64:
			return typed
		case int:
			return float64(typed)
		}
	}

	switch command {
	case "feed_now":
		return 45
	case "refresh_water":
		return 160
	case "snapshot":
		return 1
	default:
		return 1
	}
}

func max(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func (s *DeviceService) publish(ctx context.Context, subject string, payload any) {
	if s.events == nil {
		return
	}
	_ = s.events.PublishJSON(ctx, subject, payload)
}

func nullableUUID(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}
