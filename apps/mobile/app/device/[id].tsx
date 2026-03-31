import * as Haptics from "expo-haptics";
import { Stack, useLocalSearchParams } from "expo-router";
import { useEffect, useState } from "react";
import { Pressable, Text, View } from "react-native";

import { PrimaryButton } from "@/components/ui/primary-button";
import { Screen } from "@/components/ui/screen";
import { SectionCard } from "@/components/ui/section-card";
import { theme, useAppPalette } from "@/constants/theme";
import { useDeviceData, useDeviceStatus, useSendDeviceCommand } from "@/services/queries/use-devices";
import { useDeviceStore } from "@/stores/device-store";
import type { DeviceStreamMessage } from "@/types/device";
import { useI18n } from "@/utils/i18n";
import { getStoredSession } from "@/utils/session";
import { getWsUrl } from "@/utils/ws-url";

export default function DeviceDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const palette = useAppPalette();
  const { t } = useI18n();
  const statusQuery = useDeviceStatus(id);
  const dataQuery = useDeviceData(id);
  const sendDeviceCommand = useSendDeviceCommand(id);
  const liveStatus = useDeviceStore((state) => (id ? state.liveStatusById[id] : undefined));
  const setLiveStatus = useDeviceStore((state) => state.setLiveStatus);
  const clearLiveStatus = useDeviceStore((state) => state.clearLiveStatus);
  const [socketState, setSocketState] = useState<"connecting" | "live" | "closed">("connecting");

  const resolvedStatus = liveStatus ?? statusQuery.data;
  const latestPoints = liveStatus?.latest_data_points ?? dataQuery.data ?? statusQuery.data?.latest_data_points ?? [];

  useEffect(() => {
    if (!id) {
      return;
    }

    let socket: WebSocket | null = null;
    let isUnmounted = false;

    void getStoredSession().then((session) => {
      if (isUnmounted || !session?.accessToken) {
        setSocketState("closed");
        return;
      }

      socket = new WebSocket(
        getWsUrl(`/api/v1/devices/${id}/stream?access_token=${encodeURIComponent(session.accessToken)}`)
      );

      socket.onopen = () => setSocketState("live");
      socket.onclose = () => setSocketState("closed");
      socket.onerror = () => setSocketState("closed");
      socket.onmessage = (event) => {
        const payload = JSON.parse(event.data) as DeviceStreamMessage;
        if (payload.status) {
          setLiveStatus(id, payload.status);
        }
      };
    });

    return () => {
      isUnmounted = true;
      clearLiveStatus(id);
      socket?.close();
    };
  }, [clearLiveStatus, id, setLiveStatus]);

  async function handleCommand(command: string, value?: number) {
    await sendDeviceCommand.mutateAsync({
      command,
      params: value ? { value } : undefined,
    });
    await Haptics.selectionAsync().catch(() => null);
  }

  const connectionStateLabel =
    socketState === "live"
      ? t("device.detail.connection.live")
      : socketState === "closed"
        ? t("device.detail.connection.closed")
        : t("device.detail.connection.connecting");

  return (
    <Screen>
      <Stack.Screen options={{ title: resolvedStatus?.device.nickname || t("device.detail.navTitle") }} />

      <SectionCard
        title={resolvedStatus?.device.nickname ?? t("device.detail.loading")}
        subtitle={t("device.detail.connection", { state: connectionStateLabel })}
      >
        <View style={{ gap: theme.spacing.sm }}>
          <Text selectable style={{ color: palette.textSecondary }}>
            {t("device.detail.type")}：{resolvedStatus?.device.device_type ?? "--"}
          </Text>
          <Text selectable style={{ color: palette.textSecondary }}>
            {t("device.detail.status")}：{resolvedStatus?.device.status ?? "--"}
          </Text>
          <Text selectable style={{ color: palette.textSecondary }}>
            {t("device.detail.battery")}：{resolvedStatus?.device.battery_level ?? "--"}%
          </Text>
          <Text selectable style={{ color: palette.textSecondary }}>
            {t("device.detail.lastSeen")}：
            {resolvedStatus?.device.last_seen ? new Date(resolvedStatus.device.last_seen).toLocaleString() : "--"}
          </Text>
        </View>
      </SectionCard>

      <SectionCard title={t("device.detail.quickActionsTitle")} subtitle={t("device.detail.quickActionsSubtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          <PrimaryButton
            label={t("device.detail.command.feedNow")}
            onPress={() => handleCommand("feed_now", 45)}
            loading={sendDeviceCommand.isPending}
          />
          <PrimaryButton
            label={t("device.detail.command.refreshWater")}
            onPress={() => handleCommand("refresh_water", 160)}
            variant="secondary"
          />
          <PrimaryButton label={t("device.detail.command.snapshot")} onPress={() => handleCommand("snapshot")} variant="ghost" />
        </View>
      </SectionCard>

      <SectionCard title={t("device.detail.latestDataTitle")} subtitle={t("device.detail.latestDataSubtitle")}>
        <View style={{ gap: theme.spacing.sm }}>
          {latestPoints.length > 0 ? (
            latestPoints.map((point) => (
              <View
                key={point.id}
                style={{
                  borderRadius: theme.radius.md,
                  borderCurve: "continuous",
                  backgroundColor: palette.surfaceMuted,
                  padding: theme.spacing.md,
                  gap: theme.spacing.xs,
                }}
              >
                <Text selectable style={{ color: palette.text, fontWeight: "700" }}>
                  {point.metric} · {point.value}
                  {point.unit}
                </Text>
                <Text selectable style={{ color: palette.textSecondary }}>
                  {new Date(point.time).toLocaleString()}
                </Text>
              </View>
            ))
          ) : (
            <Text selectable style={{ color: palette.textSecondary }}>
              {t("device.detail.empty")}
            </Text>
          )}
        </View>
      </SectionCard>
    </Screen>
  );
}
