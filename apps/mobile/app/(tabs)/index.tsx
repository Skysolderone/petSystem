import { FlashList } from "@shopify/flash-list";
import { router } from "expo-router";
import { Text, View } from "react-native";

import { DeviceCard } from "@/components/device/device-card";
import { PetCard } from "@/components/pet/pet-card";
import { PrimaryButton } from "@/components/ui/primary-button";
import { Screen } from "@/components/ui/screen";
import { SectionCard } from "@/components/ui/section-card";
import { theme, useAppPalette } from "@/constants/theme";
import { useDevicesList } from "@/services/queries/use-devices";
import { useHealthSummary } from "@/services/queries/use-health";
import { usePetsList } from "@/services/queries/use-pets";
import { useAuthStore } from "@/stores/auth-store";
import { useNotificationStore } from "@/stores/notification-store";
import { useI18n } from "@/utils/i18n";

export default function HomeScreen() {
  const palette = useAppPalette();
  const { t } = useI18n();
  const session = useAuthStore((state) => state.session);
  const unreadCount = useNotificationStore((state) => state.unreadCount);
  const petsQuery = usePetsList();
  const devicesQuery = useDevicesList();
  const pets = petsQuery.data?.pages.flatMap((page) => page.items) ?? [];
  const devices = devicesQuery.data ?? [];
  const summaryQuery = useHealthSummary(pets[0]?.id);

  return (
    <Screen>
      <SectionCard
        title={t("home.greeting", { name: session?.user.nickname ?? t("home.defaultUser") })}
        subtitle={t("home.subtitle")}
      />

      <SectionCard title={t("home.myPetsTitle")} subtitle={t("home.myPetsSubtitle")}>
        {pets.length > 0 ? (
          <View style={{ height: 250 }}>
            <FlashList
              data={pets}
              horizontal
              estimatedItemSize={240}
              showsHorizontalScrollIndicator={false}
              contentContainerStyle={{ paddingRight: theme.spacing.md }}
              ItemSeparatorComponent={() => <View style={{ width: theme.spacing.md }} />}
              renderItem={({ item }) => <PetCard pet={item} onPress={() => router.push(`/pet/${item.id}`)} />}
            />
          </View>
        ) : (
          <View style={{ gap: theme.spacing.md }}>
            <Text selectable style={{ color: palette.textSecondary }}>
              {t("home.noPets")}
            </Text>
            <PrimaryButton label={t("home.addFirstPet")} onPress={() => router.push("/pet/add")} />
          </View>
        )}
      </SectionCard>

      <SectionCard
        title={t("home.aiInsightsTitle")}
        subtitle={pets[0]?.name ? `${pets[0].name} AI` : t("home.aiInsightsTitle")}
      >
        {summaryQuery.data ? (
          <View style={{ gap: theme.spacing.sm }}>
            <Text selectable style={{ color: palette.primary, fontWeight: "700" }}>
              {t("home.healthScore", { score: summaryQuery.data.score, status: summaryQuery.data.status })}
            </Text>
            {(summaryQuery.data.insights ?? []).slice(0, 2).map((insight) => (
              <Text key={insight} selectable style={{ color: palette.textSecondary, lineHeight: 22 }}>
                {insight}
              </Text>
            ))}
          </View>
        ) : (
          <Text selectable style={{ color: palette.textSecondary, lineHeight: 22 }}>
            {t("home.aiInsightsEmpty")}
          </Text>
        )}
      </SectionCard>

      <SectionCard title={t("home.todoTitle")} subtitle={t("home.todoSubtitle")}>
        <Text selectable style={{ color: palette.textSecondary }}>
          {t("home.todoItems")}
        </Text>
      </SectionCard>

      <SectionCard title={t("home.phase4Title")} subtitle={t("home.phase4Subtitle", { count: unreadCount })}>
        <View style={{ gap: theme.spacing.md }}>
          <PrimaryButton label={t("home.training")} onPress={() => router.push("/training")} />
          <PrimaryButton label={t("home.shop")} onPress={() => router.push("/shop")} variant="secondary" />
          <PrimaryButton label={t("home.notifications")} onPress={() => router.push("/notifications")} variant="ghost" />
        </View>
      </SectionCard>

      <SectionCard title={t("home.deviceOverviewTitle")} subtitle={t("home.deviceOverviewSubtitle")}>
        {devices.length > 0 ? (
          <View style={{ gap: theme.spacing.md }}>
            {devices.slice(0, 3).map((device) => (
              <DeviceCard key={device.id} device={device} />
            ))}
            <PrimaryButton label={t("home.viewAllDevices")} onPress={() => router.push("/device")} variant="secondary" />
          </View>
        ) : (
          <View style={{ gap: theme.spacing.md }}>
            <Text selectable style={{ color: palette.textSecondary }}>
              {t("home.noDevices")}
            </Text>
            <PrimaryButton label={t("home.bindDevice")} onPress={() => router.push("/device")} variant="secondary" />
          </View>
        )}
      </SectionCard>
    </Screen>
  );
}
