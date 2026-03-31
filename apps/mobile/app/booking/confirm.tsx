import { Stack, useLocalSearchParams } from "expo-router";
import { Text, View } from "react-native";

import { Screen } from "@/components/ui/screen";
import { SectionCard } from "@/components/ui/section-card";
import { theme, useAppPalette } from "@/constants/theme";
import { useI18n } from "@/utils/i18n";

export default function BookingConfirmScreen() {
  const palette = useAppPalette();
  const { t } = useI18n();
  const { providerName, serviceName, startTime } = useLocalSearchParams<{
    providerName?: string;
    serviceName?: string;
    startTime?: string;
  }>();

  return (
    <Screen>
      <Stack.Screen options={{ title: t("booking.confirmNavTitle") }} />

      <SectionCard title={t("booking.confirmTitle")} subtitle={t("booking.confirmSubtitle")}>
        <View style={{ gap: theme.spacing.sm }}>
          <Text selectable style={{ color: palette.textSecondary }}>
            {t("booking.confirm.provider", { name: providerName ?? "--" })}
          </Text>
          <Text selectable style={{ color: palette.textSecondary }}>
            {t("booking.confirm.service", { name: serviceName ?? "--" })}
          </Text>
          <Text selectable style={{ color: palette.textSecondary }}>
            {t("booking.confirm.startTime", {
              time: startTime ? new Date(startTime).toLocaleString() : "--",
            })}
          </Text>
        </View>
      </SectionCard>
    </Screen>
  );
}
