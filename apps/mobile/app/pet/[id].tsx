import * as Haptics from "expo-haptics";
import * as ImagePicker from "expo-image-picker";
import { Image } from "expo-image";
import { Stack, useLocalSearchParams } from "expo-router";
import { useEffect, useState } from "react";
import { Text, View } from "react-native";

import { PrimaryButton } from "@/components/ui/primary-button";
import { Screen } from "@/components/ui/screen";
import { SectionCard } from "@/components/ui/section-card";
import { TextField } from "@/components/ui/text-field";
import { theme, useAppPalette } from "@/constants/theme";
import { usePetDetail, useUpdatePet, useUploadPetAvatar } from "@/services/queries/use-pets";
import { useI18n } from "@/utils/i18n";

export default function PetDetailScreen() {
  const palette = useAppPalette();
  const { t } = useI18n();
  const { id } = useLocalSearchParams<{ id: string }>();
  const petQuery = usePetDetail(id);
  const updatePet = useUpdatePet(id);
  const uploadPetAvatar = useUploadPetAvatar(id);
  const [name, setName] = useState("");
  const [breed, setBreed] = useState("");
  const [weight, setWeight] = useState("");
  const [notes, setNotes] = useState("");

  useEffect(() => {
    if (petQuery.data) {
      setName(petQuery.data.name ?? "");
      setBreed(petQuery.data.breed ?? "");
      setWeight(petQuery.data.weight !== undefined ? String(petQuery.data.weight) : "");
      setNotes(petQuery.data.notes ?? "");
    }
  }, [petQuery.data]);

  async function handleUpdatePet() {
    await updatePet.mutateAsync({
      name,
      breed,
      weight: weight ? Number(weight) : undefined,
      notes,
    });
    await Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success).catch(() => null);
  }

  async function handleUploadAvatar() {
    const permission = await ImagePicker.requestMediaLibraryPermissionsAsync();
    if (!permission.granted) {
      return;
    }

    const result = await ImagePicker.launchImageLibraryAsync({
      mediaTypes: ImagePicker.MediaTypeOptions.Images,
      allowsEditing: true,
      quality: 0.8,
    });
    if (result.canceled || !result.assets[0]) {
      return;
    }

    const asset = result.assets[0];
    await uploadPetAvatar.mutateAsync({
      uri: asset.uri,
      name: asset.fileName ?? "pet-avatar.jpg",
      type: asset.mimeType ?? "image/jpeg",
    });
    await Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success).catch(() => null);
  }

  return (
    <Screen>
      <Stack.Screen options={{ title: petQuery.data?.name ?? t("pet.detail.title") }} />

      <SectionCard title={petQuery.data?.name ?? t("pet.detail.loading")} subtitle={`${petQuery.data?.species ?? "--"} · ${petQuery.data?.breed ?? t("pet.detail.noBreed")}`}>
        <View style={{ gap: theme.spacing.md }}>
          {petQuery.data?.avatar_url ? (
            <Image
              source={petQuery.data.avatar_url}
              style={{
                width: "100%",
                height: 220,
                borderRadius: theme.radius.lg,
              }}
              contentFit="cover"
            />
          ) : (
            <View
              style={{
                height: 220,
                borderRadius: theme.radius.lg,
                backgroundColor: palette.primaryLight,
                alignItems: "center",
                justifyContent: "center",
              }}
            >
              <Text selectable style={{ fontSize: 56 }}>
                🐶
              </Text>
            </View>
          )}
          <PrimaryButton label={t("pet.detail.uploadAvatar")} onPress={handleUploadAvatar} loading={uploadPetAvatar.isPending} variant="secondary" />
        </View>
      </SectionCard>

      <SectionCard title={t("pet.detail.editTitle")} subtitle={t("pet.detail.editSubtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          <TextField label={t("pet.detail.name")} value={name} onChangeText={setName} placeholder={t("pet.detail.namePlaceholder")} />
          <TextField label={t("pet.detail.breed")} value={breed} onChangeText={setBreed} placeholder={t("pet.detail.breedPlaceholder")} />
          <TextField label={t("pet.detail.weight")} value={weight} onChangeText={setWeight} placeholder={t("pet.detail.weightPlaceholder")} keyboardType="numeric" />
          <TextField label={t("pet.detail.notes")} value={notes} onChangeText={setNotes} placeholder={t("pet.detail.notesPlaceholder")} />
          <PrimaryButton label={t("pet.detail.save")} onPress={handleUpdatePet} loading={updatePet.isPending} disabled={!name} />
        </View>
      </SectionCard>
    </Screen>
  );
}
