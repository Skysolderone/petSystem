import * as Haptics from "expo-haptics";
import { router } from "expo-router";
import { useState } from "react";
import { Text, View } from "react-native";

import { PrimaryButton } from "@/components/ui/primary-button";
import { Screen } from "@/components/ui/screen";
import { SectionCard } from "@/components/ui/section-card";
import { TextField } from "@/components/ui/text-field";
import { theme, useAppPalette } from "@/constants/theme";
import { useCreatePet } from "@/services/queries/use-pets";
import { ApiError } from "@/types/api";
import { useI18n } from "@/utils/i18n";

export default function AddPetScreen() {
  const palette = useAppPalette();
  const { t } = useI18n();
  const [name, setName] = useState("");
  const [species, setSpecies] = useState("dog");
  const [breed, setBreed] = useState("");
  const [gender, setGender] = useState("unknown");
  const [weight, setWeight] = useState("");
  const [notes, setNotes] = useState("");
  const createPetMutation = useCreatePet();

  const errorMessage =
    createPetMutation.error instanceof ApiError ? createPetMutation.error.message : undefined;

  async function handleCreate() {
    await createPetMutation.mutateAsync({
      name,
      species,
      breed,
      gender,
      weight: weight ? Number(weight) : undefined,
      notes,
    });
    await Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success).catch(() => null);
    router.back();
  }

  return (
    <Screen>
      <SectionCard title={t("pet.add.title")} subtitle={t("pet.add.subtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          <TextField label={t("pet.add.name")} value={name} onChangeText={setName} placeholder={t("pet.add.namePlaceholder")} />
          <TextField label={t("pet.add.species")} value={species} onChangeText={setSpecies} placeholder="dog / cat" />
          <TextField label={t("pet.add.breed")} value={breed} onChangeText={setBreed} placeholder={t("pet.add.breedPlaceholder")} />
          <TextField label={t("pet.add.gender")} value={gender} onChangeText={setGender} placeholder="male / female / unknown" />
          <TextField
            label={t("pet.add.weight")}
            value={weight}
            onChangeText={setWeight}
            placeholder={t("pet.add.weightPlaceholder")}
            keyboardType="numeric"
          />
          <TextField label={t("pet.add.notes")} value={notes} onChangeText={setNotes} placeholder={t("pet.add.notesPlaceholder")} />
          {errorMessage ? (
            <Text selectable style={{ color: palette.error }}>
              {errorMessage}
            </Text>
          ) : null}
          <PrimaryButton
            label={t("pet.add.submit")}
            onPress={handleCreate}
            loading={createPetMutation.isPending}
            disabled={!name || !species}
          />
        </View>
      </SectionCard>
    </Screen>
  );
}
