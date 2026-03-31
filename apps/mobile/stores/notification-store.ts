import { create } from "zustand";

interface NotificationState {
  permissionStatus: string;
  unreadCount: number;
  setPermissionStatus: (status: string) => void;
  setUnreadCount: (count: number) => void;
  reset: () => void;
}

export const useNotificationStore = create<NotificationState>((set) => ({
  permissionStatus: "unknown",
  unreadCount: 0,
  setPermissionStatus: (permissionStatus) => set({ permissionStatus }),
  setUnreadCount: (unreadCount) => set({ unreadCount }),
  reset: () =>
    set({
      permissionStatus: "unknown",
      unreadCount: 0,
    }),
}));
