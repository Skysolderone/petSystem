import { useEffect, useState } from "react";
import { Pressable, Text, View } from "react-native";

import { PrimaryButton } from "@/components/ui/primary-button";
import { Screen } from "@/components/ui/screen";
import { SectionCard } from "@/components/ui/section-card";
import { TextField } from "@/components/ui/text-field";
import { theme, useAppPalette } from "@/constants/theme";
import { usePetsList } from "@/services/queries/use-pets";
import { useShopProducts, useShopRecommendations } from "@/services/queries/use-shop";
import { usePetStore } from "@/stores/pet-store";
import { useI18n } from "@/utils/i18n";

const categories = [
  { value: "" },
  { value: "food" },
  { value: "health" },
  { value: "toys" },
  { value: "grooming" },
] as const;

export default function ShopScreen() {
  const palette = useAppPalette();
  const { t } = useI18n();
  const petsQuery = usePetsList();
  const pets = petsQuery.data?.pages.flatMap((page) => page.items) ?? [];
  const selectedPetId = usePetStore((state) => state.selectedPetId);
  const setSelectedPetId = usePetStore((state) => state.setSelectedPetId);
  const [selectedCategory, setSelectedCategory] = useState<(typeof categories)[number]["value"]>("");
  const [search, setSearch] = useState("");
  const productsQuery = useShopProducts(selectedCategory || undefined, search || undefined);
  const recommendationsQuery = useShopRecommendations(selectedPetId ?? pets[0]?.id);

  useEffect(() => {
    if (!selectedPetId && pets[0]?.id) {
      setSelectedPetId(pets[0].id);
    }
  }, [pets, selectedPetId, setSelectedPetId]);

  function getCategoryLabel(category: string) {
    switch (category) {
      case "":
        return t("shop.category.all");
      case "food":
        return t("shop.category.food");
      case "health":
        return t("shop.category.health");
      case "toys":
        return t("shop.category.toys");
      case "grooming":
        return t("shop.category.grooming");
      default:
        return category;
    }
  }

  return (
    <Screen>
      <SectionCard title={t("shop.title")} subtitle={t("shop.subtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          <TextField label={t("shop.search")} value={search} onChangeText={setSearch} placeholder={t("shop.searchPlaceholder")} />
          <View style={{ flexDirection: "row", flexWrap: "wrap", gap: theme.spacing.sm }}>
            {categories.map((category) => (
              <Pressable
                key={category.value}
                onPress={() => setSelectedCategory(category.value)}
                style={{
                  borderRadius: theme.radius.pill,
                  borderCurve: "continuous",
                  paddingHorizontal: 14,
                  paddingVertical: 10,
                  backgroundColor: selectedCategory === category.value ? palette.primary : palette.surfaceMuted,
                }}
              >
                <Text selectable style={{ color: selectedCategory === category.value ? "#FFFFFF" : palette.textSecondary, fontWeight: "700" }}>
                  {getCategoryLabel(category.value)}
                </Text>
              </Pressable>
            ))}
          </View>
        </View>
      </SectionCard>

      <SectionCard title={t("shop.recommendationsTitle")} subtitle={t("shop.recommendationsSubtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          <View style={{ flexDirection: "row", flexWrap: "wrap", gap: theme.spacing.sm }}>
            {pets.map((pet) => (
              <PrimaryButton
                key={pet.id}
                label={pet.name}
                onPress={() => setSelectedPetId(pet.id)}
                variant={selectedPetId === pet.id ? "primary" : "secondary"}
              />
            ))}
          </View>
          {(recommendationsQuery.data ?? []).length > 0 ? (
            recommendationsQuery.data?.map((product) => (
              <View
                key={product.id}
                style={{
                  borderRadius: theme.radius.lg,
                  borderCurve: "continuous",
                  padding: theme.spacing.md,
                  backgroundColor: palette.surfaceMuted,
                  gap: theme.spacing.xs,
                }}
              >
                <Text selectable style={{ color: palette.text, fontWeight: "700" }}>
                  {product.name}
                </Text>
                <Text selectable style={{ color: palette.textSecondary }}>
                  {product.recommended_reason ?? t("shop.recommendationFallback")}
                </Text>
                <Text selectable style={{ color: palette.primary }}>
                  {product.currency} {(product.price / 100).toFixed(2)}
                </Text>
              </View>
            ))
          ) : (
            <Text selectable style={{ color: palette.textSecondary }}>
              {t("shop.recommendationsEmpty")}
            </Text>
          )}
        </View>
      </SectionCard>

      <SectionCard title={t("shop.listTitle")} subtitle={t("shop.listSubtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          {(productsQuery.data ?? []).length > 0 ? (
            productsQuery.data?.map((product) => (
              <View
                key={product.id}
                style={{
                  borderRadius: theme.radius.lg,
                  borderCurve: "continuous",
                  padding: theme.spacing.md,
                  backgroundColor: palette.surface,
                  borderWidth: 1,
                  borderColor: palette.border,
                  gap: theme.spacing.xs,
                }}
              >
                <Text selectable style={{ color: palette.text, fontWeight: "700" }}>
                  {product.name}
                </Text>
                <Text selectable style={{ color: palette.textSecondary }}>
                  {t("shop.productMeta", {
                    category: getCategoryLabel(product.category),
                    rating: product.rating.toFixed(1),
                  })}
                </Text>
                <Text selectable style={{ color: palette.textSecondary, lineHeight: 22 }}>
                  {product.description}
                </Text>
                <Text selectable style={{ color: palette.primary }}>
                  {product.currency} {(product.price / 100).toFixed(2)}
                </Text>
              </View>
            ))
          ) : (
            <Text selectable style={{ color: palette.textSecondary }}>
              {t("shop.listEmpty")}
            </Text>
          )}
        </View>
      </SectionCard>
    </Screen>
  );
}
