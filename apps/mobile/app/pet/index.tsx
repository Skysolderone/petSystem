import { FlashList } from "@shopify/flash-list";
import { router } from "expo-router";
import { Text, View } from "react-native";

import { PetCard } from "@/components/pet/pet-card";
import { PrimaryButton } from "@/components/ui/primary-button";
import { Screen } from "@/components/ui/screen";
import { SectionCard } from "@/components/ui/section-card";
import { theme, useAppPalette } from "@/constants/theme";
import { usePetsList } from "@/services/queries/use-pets";
import { useI18n } from "@/utils/i18n";

export default function PetListScreen() {
  const palette = useAppPalette();
  const { t } = useI18n();
  const petsQuery = usePetsList();
  const pets = petsQuery.data?.pages.flatMap((page) => page.items) ?? [];

  return (
    <Screen>
      <SectionCard title={t("pet.list.title")} subtitle={t("pet.list.subtitle")}>
        <PrimaryButton label={t("pet.list.add")} onPress={() => router.push("/pet/add")} />
      </SectionCard>

      <SectionCard title={t("pet.list.archive")}>
        <View style={{ minHeight: 320 }}>
          {pets.length > 0 ? (
            <FlashList
              data={pets}
              estimatedItemSize={260}
              ItemSeparatorComponent={() => <View style={{ height: theme.spacing.md }} />}
              renderItem={({ item }) => <PetCard pet={item} onPress={() => router.push(`/pet/${item.id}`)} />}
            />
          ) : (
            <Text selectable style={{ color: palette.textSecondary }}>
              {t("pet.list.empty")}
            </Text>
          )}
        </View>
      </SectionCard>
    </Screen>
  );
}
