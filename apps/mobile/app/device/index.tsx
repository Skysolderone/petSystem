import * as Haptics from "expo-haptics";
import { router } from "expo-router";
import { useEffect, useState } from "react";
import { Pressable, Text, View } from "react-native";

import { DeviceCard } from "@/components/device/device-card";
import { PrimaryButton } from "@/components/ui/primary-button";
import { Screen } from "@/components/ui/screen";
import { SectionCard } from "@/components/ui/section-card";
import { TextField } from "@/components/ui/text-field";
import { theme, useAppPalette } from "@/constants/theme";
import { useCreateDevice, useDevicesList } from "@/services/queries/use-devices";
import { usePetsList } from "@/services/queries/use-pets";
import { usePetStore } from "@/stores/pet-store";
import { ApiError } from "@/types/api";
import { useI18n } from "@/utils/i18n";

const deviceTypes = ["feeder", "water_fountain", "camera", "gps_collar"] as const;

export default function DeviceListScreen() {
  const palette = useAppPalette();
  const { t } = useI18n();
  const petsQuery = usePetsList();
  const devicesQuery = useDevicesList();
  const createDeviceMutation = useCreateDevice();
  const selectedPetId = usePetStore((state) => state.selectedPetId);
  const setSelectedPetId = usePetStore((state) => state.setSelectedPetId);
  const pets = petsQuery.data?.pages.flatMap((page) => page.items) ?? [];
  const devices = devicesQuery.data ?? [];

  const [nickname, setNickname] = useState("");
  const [serialNumber, setSerialNumber] = useState("");
  const [deviceType, setDeviceType] = useState<(typeof deviceTypes)[number]>("feeder");

  useEffect(() => {
    if (!selectedPetId && pets[0]?.id) {
      setSelectedPetId(pets[0].id);
    }
  }, [pets, selectedPetId, setSelectedPetId]);

  async function handleCreateDevice() {
    await createDeviceMutation.mutateAsync({
      pet_id: selectedPetId ?? undefined,
      nickname,
      device_type: deviceType,
      brand: "PetVerse",
      model: deviceType.toUpperCase(),
      serial_number: serialNumber,
    });
    await Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success).catch(() => null);
    setNickname("");
    setSerialNumber("");
  }

  const errorMessage =
    createDeviceMutation.error instanceof ApiError ? createDeviceMutation.error.message : undefined;

  return (
    <Screen>
      <SectionCard title={t("device.list.bindTitle")} subtitle={t("device.list.bindSubtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          <View style={{ flexDirection: "row", flexWrap: "wrap", gap: theme.spacing.sm }}>
            {deviceTypes.map((type) => (
              <Pressable
                key={type}
                onPress={() => setDeviceType(type)}
                style={{
                  borderRadius: theme.radius.pill,
                  borderCurve: "continuous",
                  paddingHorizontal: 14,
                  paddingVertical: 10,
                  backgroundColor: deviceType === type ? palette.primary : palette.surfaceMuted,
                }}
              >
                <Text selectable style={{ color: deviceType === type ? "#FFFFFF" : palette.textSecondary, fontWeight: "600" }}>
                  {type}
                </Text>
              </Pressable>
            ))}
          </View>
          <TextField
            label={t("device.list.nickname")}
            value={nickname}
            onChangeText={setNickname}
            placeholder={t("device.list.nicknamePlaceholder")}
          />
          <TextField
            label={t("device.list.serialNumber")}
            value={serialNumber}
            onChangeText={setSerialNumber}
            placeholder={t("device.list.serialNumberPlaceholder")}
          />
          <Text selectable style={{ color: palette.textSecondary }}>
            {t("device.list.selectedPet", {
              name: pets.find((pet) => pet.id === selectedPetId)?.name ?? t("device.list.noPetSelected"),
            })}
          </Text>
          {pets.length > 0 ? (
            <View style={{ flexDirection: "row", flexWrap: "wrap", gap: theme.spacing.sm }}>
              {pets.map((pet) => (
                <Pressable
                  key={pet.id}
                  onPress={() => setSelectedPetId(pet.id)}
                  style={{
                    borderRadius: theme.radius.pill,
                    borderCurve: "continuous",
                    paddingHorizontal: 14,
                    paddingVertical: 10,
                    borderWidth: 1,
                    borderColor: selectedPetId === pet.id ? palette.primary : palette.border,
                    backgroundColor: selectedPetId === pet.id ? palette.primaryLight : palette.surface,
                  }}
                >
                  <Text selectable style={{ color: palette.text, fontWeight: "600" }}>
                    {pet.name}
                  </Text>
                </Pressable>
              ))}
            </View>
          ) : null}
          {errorMessage ? (
            <Text selectable style={{ color: palette.error }}>
              {errorMessage}
            </Text>
          ) : null}
          <PrimaryButton
            label={t("device.list.bindSubmit")}
            onPress={handleCreateDevice}
            loading={createDeviceMutation.isPending}
            disabled={!serialNumber}
          />
        </View>
      </SectionCard>

      <SectionCard title={t("device.list.sectionTitle")} subtitle={t("device.list.sectionSubtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          {devices.length > 0 ? (
            devices.map((device) => (
              <Pressable key={device.id} onPress={() => router.push(`/device/${device.id}`)}>
                <DeviceCard device={device} />
              </Pressable>
            ))
          ) : (
            <Text selectable style={{ color: palette.textSecondary }}>
              {t("device.list.empty")}
            </Text>
          )}
        </View>
      </SectionCard>
    </Screen>
  );
}
