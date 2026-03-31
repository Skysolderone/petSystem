import { Text, View } from "react-native";

import { theme, useAppPalette } from "@/constants/theme";
import type { Device } from "@/types/device";
import { useI18n } from "@/utils/i18n";

interface DeviceCardProps {
  device: Device;
}

export function DeviceCard({ device }: DeviceCardProps) {
  const palette = useAppPalette();
  const { t } = useI18n();
  const model = device.model || t("device.card.modelFallback");

  return (
    <View
      style={{
        borderRadius: theme.radius.lg,
        borderCurve: "continuous",
        padding: theme.spacing.md,
        gap: theme.spacing.sm,
        backgroundColor: palette.surface,
        borderWidth: 1,
        borderColor: palette.border,
      }}
    >
      <Text selectable style={{ color: palette.text, fontWeight: "700", fontSize: theme.fontSize.md }}>
        {device.nickname || device.device_type}
      </Text>
      <Text selectable style={{ color: palette.textSecondary }}>
        {t("device.card.summary", { brand: device.brand || "PetVerse", model })}
      </Text>
      <Text selectable style={{ color: palette.primary, fontWeight: "700" }}>
        {t("device.card.status", { status: device.status.toUpperCase(), battery: device.battery_level ?? "--" })}
      </Text>
    </View>
  );
}
