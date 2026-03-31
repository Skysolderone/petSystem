import "react-native-gesture-handler";
import "react-native-reanimated";

import { QueryClient, QueryClientProvider, useQueryClient } from "@tanstack/react-query";
import * as Notifications from "expo-notifications";
import { Stack } from "expo-router";
import { useEffect } from "react";
import { GestureHandlerRootView } from "react-native-gesture-handler";

import { useAuthStore } from "@/stores/auth-store";
import { useNotificationStore } from "@/stores/notification-store";
import { useSettingsStore } from "@/stores/settings-store";
import { registerPushToken, unregisterPushToken } from "@/services/queries/use-notifications";
import type { NotificationStreamMessage } from "@/types/notification";
import { useI18n } from "@/utils/i18n";
import { clearStoredPushToken, getStoredPushToken, setStoredPushToken } from "@/utils/push-token";
import { getWsUrl } from "@/utils/ws-url";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60,
      retry: 1,
    },
  },
});

Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowAlert: true,
    shouldPlaySound: false,
    shouldSetBadge: false,
  }),
});

function AuthBootstrap() {
  const hydrate = useAuthStore((state) => state.hydrate);

  useEffect(() => {
    void hydrate();
  }, [hydrate]);

  return null;
}

function SettingsBootstrap() {
  const hydrate = useSettingsStore((state) => state.hydrate);

  useEffect(() => {
    void hydrate();
  }, [hydrate]);

  return null;
}

function NotificationBootstrap() {
  const session = useAuthStore((state) => state.session);
  const queryClient = useQueryClient();
  const setPermissionStatus = useNotificationStore((state) => state.setPermissionStatus);
  const setUnreadCount = useNotificationStore((state) => state.setUnreadCount);
  const resetNotifications = useNotificationStore((state) => state.reset);
  const notificationsEnabled = useSettingsStore((state) => state.notificationsEnabled);

  useEffect(() => {
    if (!session?.accessToken) {
      resetNotifications();
      return;
    }

    let isUnmounted = false;

    void (async () => {
      const storedPushToken = await getStoredPushToken();

      if (!notificationsEnabled) {
        setPermissionStatus("disabled");
        if (storedPushToken) {
          await unregisterPushToken(storedPushToken).catch(() => null);
          await clearStoredPushToken().catch(() => null);
        }
        return;
      }

      const permission = await Notifications.requestPermissionsAsync().catch(() => null);
      const status = permission?.status ?? "undetermined";
      setPermissionStatus(status);

      if (status !== "granted" || isUnmounted) {
        if (storedPushToken) {
          await unregisterPushToken(storedPushToken).catch(() => null);
          await clearStoredPushToken().catch(() => null);
        }
        return;
      }

      const projectId = process.env.EXPO_PUBLIC_EXPO_PROJECT_ID;
      const tokenResponse = projectId
        ? await Notifications.getExpoPushTokenAsync({ projectId }).catch(() => null)
        : await Notifications.getExpoPushTokenAsync().catch(() => null);
      const expoPushToken = tokenResponse?.data;
      if (!expoPushToken) {
        return;
      }

      await registerPushToken({
        token: expoPushToken,
        provider: "expo",
        platform: process.env.EXPO_OS ?? "unknown",
      }).catch(() => null);
      await setStoredPushToken(expoPushToken).catch(() => null);
    })();

    const socket = new WebSocket(getWsUrl(`/api/v1/ws?access_token=${session.accessToken}`));
    socket.onmessage = (event) => {
      const payload = JSON.parse(event.data) as NotificationStreamMessage;
      if (typeof payload.unread_count === "number") {
        setUnreadCount(payload.unread_count);
      }
      if (notificationsEnabled && payload.type === "notification" && payload.notification) {
        void Notifications.scheduleNotificationAsync({
          content: {
            title: payload.notification.title,
            body: payload.notification.body,
          },
          trigger: null,
        }).catch(() => null);
      }
      void queryClient.invalidateQueries({ queryKey: ["notifications"] });
    };

    return () => {
      isUnmounted = true;
      socket.close();
    };
  }, [notificationsEnabled, queryClient, resetNotifications, session?.accessToken, setPermissionStatus, setUnreadCount]);

  return null;
}

export default function RootLayout() {
  const { t } = useI18n();

  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
      <QueryClientProvider client={queryClient}>
        <AuthBootstrap />
        <SettingsBootstrap />
        <NotificationBootstrap />
        <Stack
          screenOptions={{
            headerBackTitle: t("common.back"),
          }}
        >
          <Stack.Screen name="index" options={{ headerShown: false }} />
          <Stack.Screen name="auth/login" options={{ title: t("auth.login.submit") }} />
          <Stack.Screen name="auth/register" options={{ title: t("auth.register.submit") }} />
          <Stack.Screen name="(tabs)" options={{ headerShown: false }} />
          <Stack.Screen name="pet/index" options={{ title: t("pet.list.title") }} />
          <Stack.Screen name="pet/[id]" options={{ title: t("pet.detail.title") }} />
          <Stack.Screen name="pet/add" options={{ title: t("pet.add.title") }} />
          <Stack.Screen name="device/index" options={{ title: t("device.list.navTitle") }} />
          <Stack.Screen name="device/[id]" options={{ title: t("device.detail.navTitle") }} />
          <Stack.Screen name="booking/[serviceId]" options={{ title: t("booking.navTitle") }} />
          <Stack.Screen name="booking/confirm" options={{ title: t("booking.confirmNavTitle") }} />
          <Stack.Screen name="training/index" options={{ title: t("training.navTitle") }} />
          <Stack.Screen name="training/[planId]" options={{ title: t("training.navTitle") }} />
          <Stack.Screen name="shop/index" options={{ title: t("home.shop") }} />
          <Stack.Screen name="notifications/index" options={{ title: t("notifications.title") }} />
        </Stack>
      </QueryClientProvider>
    </GestureHandlerRootView>
  );
}
