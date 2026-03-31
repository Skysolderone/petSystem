import { router } from "expo-router";
import { useState } from "react";
import { Pressable, Text, View } from "react-native";
import MapView, { Marker } from "react-native-maps";

import { ProviderCard } from "@/components/service/provider-card";
import { PrimaryButton } from "@/components/ui/primary-button";
import { Screen } from "@/components/ui/screen";
import { SectionCard } from "@/components/ui/section-card";
import { theme, useAppPalette } from "@/constants/theme";
import { useServicesList } from "@/services/queries/use-services";
import { useI18n } from "@/utils/i18n";

const filters = ["", "vet_clinic", "grooming", "boarding"] as const;

export default function ServicesScreen() {
  const palette = useAppPalette();
  const { t } = useI18n();
  const [selectedFilter, setSelectedFilter] = useState<(typeof filters)[number]>("");
  const servicesQuery = useServicesList(selectedFilter || undefined);
  const providers = servicesQuery.data ?? [];
  const firstProvider = providers[0];

  function getFilterLabel(filter: (typeof filters)[number]) {
    switch (filter) {
      case "":
        return t("services.filter.all");
      case "vet_clinic":
        return t("services.filter.vet");
      case "grooming":
        return t("services.filter.grooming");
      case "boarding":
        return t("services.filter.boarding");
    }
  }

  return (
    <Screen>
      <SectionCard title={t("services.nearbyTitle")} subtitle={t("services.nearbySubtitle")}>
        <View style={{ flexDirection: "row", flexWrap: "wrap", gap: theme.spacing.sm }}>
          {filters.map((filter) => (
            <Pressable
              key={filter}
              onPress={() => setSelectedFilter(filter)}
              style={{
                borderRadius: theme.radius.pill,
                borderCurve: "continuous",
                paddingHorizontal: 14,
                paddingVertical: 10,
                backgroundColor: selectedFilter === filter ? palette.primary : palette.surfaceMuted,
              }}
            >
              <Text selectable style={{ color: selectedFilter === filter ? "#FFFFFF" : palette.textSecondary, fontWeight: "600" }}>
                {getFilterLabel(filter)}
              </Text>
            </Pressable>
          ))}
        </View>
      </SectionCard>

      <SectionCard title={t("services.mapTitle")} subtitle={t("services.mapSubtitle")}>
        {firstProvider ? (
          <MapView
            style={{ height: 260, borderRadius: theme.radius.lg }}
            initialRegion={{
              latitude: firstProvider.latitude,
              longitude: firstProvider.longitude,
              latitudeDelta: 0.15,
              longitudeDelta: 0.15,
            }}
          >
            {providers.map((provider) => (
              <Marker
                key={provider.id}
                coordinate={{ latitude: provider.latitude, longitude: provider.longitude }}
                title={provider.name}
                description={provider.address}
              />
            ))}
          </MapView>
        ) : (
          <Text selectable style={{ color: palette.textSecondary }}>
            {t("services.empty")}
          </Text>
        )}
      </SectionCard>

      <SectionCard title={t("services.listTitle")} subtitle={t("services.listSubtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          {providers.map((provider) => (
            <View key={provider.id} style={{ gap: theme.spacing.sm }}>
              <ProviderCard provider={provider} />
              <PrimaryButton label={t("services.book")} onPress={() => router.push(`/booking/${provider.id}`)} />
            </View>
          ))}
        </View>
      </SectionCard>
    </Screen>
  );
}
