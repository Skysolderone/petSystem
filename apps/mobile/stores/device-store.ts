import { create } from "zustand";

import type { DeviceStatus } from "@/types/device";

interface DeviceState {
  liveStatusById: Record<string, DeviceStatus>;
  setLiveStatus: (deviceId: string, status: DeviceStatus) => void;
  clearLiveStatus: (deviceId: string) => void;
}

export const useDeviceStore = create<DeviceState>((set) => ({
  liveStatusById: {},
  setLiveStatus: (deviceId, status) =>
    set((state) => ({
      liveStatusById: {
        ...state.liveStatusById,
        [deviceId]: status,
      },
    })),
  clearLiveStatus: (deviceId) =>
    set((state) => {
      const nextState = { ...state.liveStatusById };
      delete nextState[deviceId];
      return {
        liveStatusById: nextState,
      };
    }),
}));
