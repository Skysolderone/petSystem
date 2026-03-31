import { Text, View } from "react-native";

import { theme, useAppPalette } from "@/constants/theme";
import type { ServiceProvider } from "@/types/service";
import { useI18n } from "@/utils/i18n";

interface ProviderCardProps {
  provider: ServiceProvider;
}

export function ProviderCard({ provider }: ProviderCardProps) {
  const palette = useAppPalette();
  const { t } = useI18n();
  const localizedType =
    provider.type === "vet_clinic"
      ? t("service.type.vet_clinic")
      : provider.type === "grooming"
        ? t("service.type.grooming")
        : provider.type === "boarding"
          ? t("service.type.boarding")
          : provider.type;

  return (
    <View
      style={{
        borderRadius: theme.radius.lg,
        borderCurve: "continuous",
        padding: theme.spacing.md,
        gap: theme.spacing.sm,
        borderWidth: 1,
        borderColor: palette.border,
        backgroundColor: palette.surface,
      }}
    >
      <Text selectable style={{ color: palette.text, fontWeight: "700", fontSize: theme.fontSize.md }}>
        {provider.name}
      </Text>
      <Text selectable style={{ color: palette.textSecondary }}>
        {t("service.card.meta", {
          type: localizedType,
          rating: provider.rating.toFixed(1),
          count: provider.review_count,
        })}
      </Text>
      <Text selectable style={{ color: palette.textSecondary, lineHeight: 22 }}>
        {provider.description}
      </Text>
      <Text selectable style={{ color: palette.primary }}>
        {typeof provider.distance_km === "number"
          ? t("service.card.distance", {
              address: provider.address,
              distance: provider.distance_km.toFixed(1),
            })
          : provider.address}
      </Text>
    </View>
  );
}
