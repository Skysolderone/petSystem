import * as Haptics from "expo-haptics";
import { Stack, router, useLocalSearchParams } from "expo-router";
import { useEffect, useState } from "react";
import { Pressable, Text, View } from "react-native";

import { PrimaryButton } from "@/components/ui/primary-button";
import { Screen } from "@/components/ui/screen";
import { SectionCard } from "@/components/ui/section-card";
import { TextField } from "@/components/ui/text-field";
import { theme, useAppPalette } from "@/constants/theme";
import { useCreateBooking } from "@/services/queries/use-bookings";
import { usePetsList } from "@/services/queries/use-pets";
import { useServiceAvailability, useServiceDetail, useServiceReviews } from "@/services/queries/use-services";
import { usePetStore } from "@/stores/pet-store";
import { useI18n } from "@/utils/i18n";

export default function BookingScreen() {
  const { serviceId } = useLocalSearchParams<{ serviceId: string }>();
  const palette = useAppPalette();
  const { t } = useI18n();
  const providerQuery = useServiceDetail(serviceId);
  const availabilityQuery = useServiceAvailability(serviceId);
  const reviewsQuery = useServiceReviews(serviceId);
  const petsQuery = usePetsList();
  const createBooking = useCreateBooking();
  const selectedPetId = usePetStore((state) => state.selectedPetId);
  const setSelectedPetId = usePetStore((state) => state.setSelectedPetId);
  const pets = petsQuery.data?.pages.flatMap((page) => page.items) ?? [];
  const slots = availabilityQuery.data?.slots ?? [];
  const [selectedSlot, setSelectedSlot] = useState<string | null>(null);
  const [notes, setNotes] = useState("");

  useEffect(() => {
    if (!selectedPetId && pets[0]?.id) {
      setSelectedPetId(pets[0].id);
    }
  }, [pets, selectedPetId, setSelectedPetId]);

  async function handleBooking() {
    if (!serviceId || !selectedPetId || !selectedSlot || !providerQuery.data) {
      return;
    }

    const booking = await createBooking.mutateAsync({
      provider_id: serviceId,
      pet_id: selectedPetId,
      service_name: String(providerQuery.data.services[0]?.name ?? providerQuery.data.name),
      start_time: selectedSlot,
      price: Number(providerQuery.data.services[0]?.price ?? 0),
      notes,
    });

    await Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success).catch(() => null);
    router.replace({
      pathname: "/booking/confirm",
      params: {
        providerName: providerQuery.data.name,
        serviceName: booking.service_name,
        startTime: booking.start_time,
      },
    });
  }

  return (
    <Screen>
      <Stack.Screen options={{ title: providerQuery.data?.name ?? t("booking.navTitle") }} />

      <SectionCard title={providerQuery.data?.name ?? t("booking.loading")} subtitle={providerQuery.data?.address ?? t("booking.loadingSubtitle")}>
        <Text selectable style={{ color: palette.textSecondary, lineHeight: 22 }}>
          {providerQuery.data?.description ?? t("booking.descriptionLoading")}
        </Text>
      </SectionCard>

      <SectionCard title={t("booking.petTitle")} subtitle={t("booking.petSubtitle")}>
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
              <Text selectable style={{ color: palette.text, fontWeight: "700" }}>
                {pet.name}
              </Text>
            </Pressable>
          ))}
        </View>
      </SectionCard>

      <SectionCard title={t("booking.slotTitle")} subtitle={t("booking.slotSubtitle")}>
        <View style={{ flexDirection: "row", flexWrap: "wrap", gap: theme.spacing.sm }}>
          {slots.slice(0, 8).map((slot) => (
            <Pressable
              key={slot}
              onPress={() => setSelectedSlot(slot)}
              style={{
                borderRadius: theme.radius.md,
                borderCurve: "continuous",
                paddingHorizontal: 14,
                paddingVertical: 10,
                backgroundColor: selectedSlot === slot ? palette.primary : palette.surfaceMuted,
              }}
            >
              <Text selectable style={{ color: selectedSlot === slot ? "#FFFFFF" : palette.textSecondary, fontWeight: "600" }}>
                {new Date(slot).toLocaleString()}
              </Text>
            </Pressable>
          ))}
        </View>
      </SectionCard>

      <SectionCard title={t("booking.notesTitle")} subtitle={t("booking.notesSubtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          <TextField label={t("booking.notesLabel")} value={notes} onChangeText={setNotes} placeholder={t("booking.notesPlaceholder")} />
          <PrimaryButton
            label={t("booking.confirm")}
            onPress={handleBooking}
            loading={createBooking.isPending}
            disabled={!selectedPetId || !selectedSlot}
          />
        </View>
      </SectionCard>

      <SectionCard title={t("booking.reviewsTitle")} subtitle={t("booking.reviewsSubtitle")}>
        <View style={{ gap: theme.spacing.sm }}>
          {(reviewsQuery.data ?? []).slice(0, 3).map((review) => (
            <View
              key={review.id}
              style={{
                borderRadius: theme.radius.md,
                borderCurve: "continuous",
                backgroundColor: palette.surfaceMuted,
                padding: theme.spacing.md,
                gap: theme.spacing.xs,
              }}
            >
              <Text selectable style={{ color: palette.text, fontWeight: "700" }}>
                {review.rating ?? "--"} / 5
              </Text>
              <Text selectable style={{ color: palette.textSecondary }}>
                {review.review || t("booking.reviewEmpty")}
              </Text>
            </View>
          ))}
        </View>
      </SectionCard>
    </Screen>
  );
}
