export interface Device {
  id: string;
  owner_id: string;
  pet_id?: string;
  device_type: string;
  brand: string;
  model: string;
  nickname: string;
  serial_number: string;
  status: string;
  config: Record<string, unknown>;
  last_seen?: string;
  firmware_ver: string;
  battery_level?: number;
  created_at: string;
  updated_at: string;
}

export interface DeviceDataPoint {
  id: string;
  time: string;
  device_id: string;
  metric: string;
  value: number;
  unit: string;
  meta: Record<string, unknown>;
}

export interface DeviceStatus {
  device: Device;
  latest_data_points: DeviceDataPoint[];
}

export interface DeviceCommandResponse {
  accepted: boolean;
  command: string;
  executed_at: string;
  data_point: DeviceDataPoint;
}

export interface DeviceStreamMessage {
  type: string;
  timestamp: string;
  status?: DeviceStatus;
}

export interface CreateDevicePayload {
  pet_id?: string;
  device_type: string;
  brand?: string;
  model?: string;
  nickname?: string;
  serial_number: string;
}
