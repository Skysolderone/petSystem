package dto

import (
	"time"

	"petverse/server/internal/model"
)

type CreateDeviceRequest struct {
	PetID        *string        `json:"pet_id"`
	DeviceType   string         `json:"device_type" binding:"required,min=1,max=30"`
	Brand        string         `json:"brand"`
	Model        string         `json:"model"`
	Nickname     string         `json:"nickname"`
	SerialNumber string         `json:"serial_number" binding:"required,min=3,max=100"`
	Status       string         `json:"status"`
	Config       map[string]any `json:"config"`
	FirmwareVer  string         `json:"firmware_ver"`
	BatteryLevel *int           `json:"battery_level"`
}

type UpdateDeviceRequest struct {
	PetID        *string        `json:"pet_id"`
	DeviceType   *string        `json:"device_type"`
	Brand        *string        `json:"brand"`
	Model        *string        `json:"model"`
	Nickname     *string        `json:"nickname"`
	Status       *string        `json:"status"`
	Config       map[string]any `json:"config"`
	FirmwareVer  *string        `json:"firmware_ver"`
	BatteryLevel *int           `json:"battery_level"`
}

type DeviceCommandRequest struct {
	Command string         `json:"command" binding:"required,min=1,max=50"`
	Params  map[string]any `json:"params"`
}

type DeviceResponse struct {
	ID           string         `json:"id"`
	OwnerID      string         `json:"owner_id"`
	PetID        *string        `json:"pet_id,omitempty"`
	DeviceType   string         `json:"device_type"`
	Brand        string         `json:"brand"`
	Model        string         `json:"model"`
	Nickname     string         `json:"nickname"`
	SerialNumber string         `json:"serial_number"`
	Status       string         `json:"status"`
	Config       map[string]any `json:"config"`
	LastSeen     *time.Time     `json:"last_seen,omitempty"`
	FirmwareVer  string         `json:"firmware_ver"`
	BatteryLevel *int           `json:"battery_level,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type DeviceDataPointResponse struct {
	ID       string         `json:"id"`
	Time     time.Time      `json:"time"`
	DeviceID string         `json:"device_id"`
	Metric   string         `json:"metric"`
	Value    float64        `json:"value"`
	Unit     string         `json:"unit"`
	Meta     map[string]any `json:"meta"`
}

type DeviceStatusResponse struct {
	Device           DeviceResponse            `json:"device"`
	LatestDataPoints []DeviceDataPointResponse `json:"latest_data_points"`
}

type DeviceCommandResponse struct {
	Accepted   bool                    `json:"accepted"`
	Command    string                  `json:"command"`
	ExecutedAt time.Time               `json:"executed_at"`
	DataPoint  DeviceDataPointResponse `json:"data_point"`
}

type DeviceStreamMessage struct {
	Type      string                `json:"type"`
	Timestamp time.Time             `json:"timestamp"`
	Status    *DeviceStatusResponse `json:"status,omitempty"`
}

func ToDeviceResponse(device *model.Device) DeviceResponse {
	var petID *string
	if device.PetID != nil {
		value := device.PetID.String()
		petID = &value
	}

	return DeviceResponse{
		ID:           device.ID.String(),
		OwnerID:      device.OwnerID.String(),
		PetID:        petID,
		DeviceType:   device.DeviceType,
		Brand:        device.Brand,
		Model:        device.Model,
		Nickname:     device.Nickname,
		SerialNumber: device.SerialNumber,
		Status:       device.Status,
		Config:       decodeMap(device.Config),
		LastSeen:     device.LastSeen,
		FirmwareVer:  device.FirmwareVer,
		BatteryLevel: device.BatteryLevel,
		CreatedAt:    device.CreatedAt,
		UpdatedAt:    device.UpdatedAt,
	}
}

func ToDeviceDataPointResponse(point *model.DeviceDataPoint) DeviceDataPointResponse {
	return DeviceDataPointResponse{
		ID:       point.ID.String(),
		Time:     point.Time,
		DeviceID: point.DeviceID.String(),
		Metric:   point.Metric,
		Value:    point.Value,
		Unit:     point.Unit,
		Meta:     decodeMap(point.Meta),
	}
}
