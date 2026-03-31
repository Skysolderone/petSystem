import { Pressable, Text, View } from "react-native";
import { Image } from "expo-image";

import { theme, useAppPalette } from "@/constants/theme";
import type { Pet } from "@/types/pet";
import { useI18n } from "@/utils/i18n";

interface PetCardProps {
  pet: Pet;
  onPress?: () => void;
}

export function PetCard({ pet, onPress }: PetCardProps) {
  const palette = useAppPalette();
  const { t } = useI18n();

  return (
    <Pressable
      onPress={onPress}
      disabled={!onPress}
      style={({ pressed }) => ({
        width: 240,
        borderRadius: theme.radius.lg,
        borderCurve: "continuous",
        padding: theme.spacing.md,
        gap: theme.spacing.md,
        backgroundColor: palette.surface,
        borderWidth: 1,
        borderColor: palette.border,
        opacity: pressed ? 0.9 : 1,
      })}
    >
      {pet.avatar_url ? (
        <Image
          source={pet.avatar_url}
          style={{
            width: "100%",
            height: 140,
            borderRadius: theme.radius.md,
          }}
          contentFit="cover"
        />
      ) : (
        <View
          style={{
            height: 140,
            borderRadius: theme.radius.md,
            backgroundColor: palette.primaryLight,
            alignItems: "center",
            justifyContent: "center",
          }}
        >
          <Text selectable style={{ fontSize: 44 }}>
            🐾
          </Text>
        </View>
      )}
      <View style={{ gap: theme.spacing.xs }}>
        <Text
          selectable
          style={{
            color: palette.text,
            fontWeight: "700",
            fontSize: theme.fontSize.lg,
          }}
        >
          {pet.name}
        </Text>
        <Text selectable style={{ color: palette.textSecondary }}>
          {pet.species} · {pet.breed || t("pet.card.noBreed")}
        </Text>
        <Text selectable style={{ color: palette.primary, fontWeight: "700" }}>
          {t("pet.card.healthScore", { score: pet.health_score ?? "--" })}
        </Text>
      </View>
    </Pressable>
  );
}
