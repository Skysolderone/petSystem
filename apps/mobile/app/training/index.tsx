import * as Haptics from "expo-haptics";
import { router, Stack } from "expo-router";
import { useEffect, useState } from "react";
import { Pressable, Text, View } from "react-native";

import { PrimaryButton } from "@/components/ui/primary-button";
import { Screen } from "@/components/ui/screen";
import { SectionCard } from "@/components/ui/section-card";
import { TextField } from "@/components/ui/text-field";
import { theme, useAppPalette } from "@/constants/theme";
import { useNotificationsList } from "@/services/queries/use-notifications";
import { usePetsList } from "@/services/queries/use-pets";
import { useCreateTrainingPlan, useGenerateTrainingPlan, useTrainingPlans } from "@/services/queries/use-training";
import { useNotificationStore } from "@/stores/notification-store";
import { usePetStore } from "@/stores/pet-store";
import { useI18n } from "@/utils/i18n";

const difficulties = ["beginner", "intermediate", "advanced"] as const;

export default function TrainingScreen() {
  const palette = useAppPalette();
  const { t } = useI18n();
  const petsQuery = usePetsList();
  const notificationsQuery = useNotificationsList();
  const selectedPetId = usePetStore((state) => state.selectedPetId);
  const setSelectedPetId = usePetStore((state) => state.setSelectedPetId);
  const unreadCount = useNotificationStore((state) => state.unreadCount);
  const pets = petsQuery.data?.pages.flatMap((page) => page.items) ?? [];
  const trainingPlansQuery = useTrainingPlans(selectedPetId ?? pets[0]?.id);
  const createTrainingPlan = useCreateTrainingPlan(selectedPetId ?? pets[0]?.id);
  const generateTrainingPlan = useGenerateTrainingPlan(selectedPetId ?? pets[0]?.id);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [goal, setGoal] = useState("");
  const [difficulty, setDifficulty] = useState<(typeof difficulties)[number]>("beginner");

  useEffect(() => {
    if (!selectedPetId && pets[0]?.id) {
      setSelectedPetId(pets[0].id);
    }
  }, [pets, selectedPetId, setSelectedPetId]);

  async function handleCreatePlan() {
    const plan = await createTrainingPlan.mutateAsync({
      title,
      description,
      difficulty,
      category: "obedience",
      steps: [
        {
          day: 1,
          title: t("training.demoStepTitle"),
          instruction: t("training.demoStepInstruction"),
          duration_minutes: 10,
        },
      ],
    });
    await Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success).catch(() => null);
    setTitle("");
    setDescription("");
    router.push(`/training/${plan.id}`);
  }

  async function handleGeneratePlan() {
    const plan = await generateTrainingPlan.mutateAsync({
      goal,
      difficulty,
      category: "obedience",
    });
    await Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success).catch(() => null);
    setGoal("");
    router.push(`/training/${plan.id}`);
  }

  function getDifficultyLabel(value: (typeof difficulties)[number]) {
    switch (value) {
      case "beginner":
        return t("training.difficulty.beginner");
      case "intermediate":
        return t("training.difficulty.intermediate");
      case "advanced":
        return t("training.difficulty.advanced");
    }
  }

  return (
    <Screen>
      <Stack.Screen options={{ title: t("training.navTitle") }} />

      <SectionCard
        title={t("training.hubTitle")}
        subtitle={t("training.hubSubtitle", {
          count: unreadCount || notificationsQuery.data?.filter((item) => !item.is_read).length || 0,
        })}
      >
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
                backgroundColor: selectedPetId === pet.id ? palette.primary : palette.surfaceMuted,
              }}
            >
              <Text selectable style={{ color: selectedPetId === pet.id ? "#FFFFFF" : palette.textSecondary, fontWeight: "700" }}>
                {pet.name}
              </Text>
            </Pressable>
          ))}
        </View>
      </SectionCard>

      <SectionCard title={t("training.manualTitle")} subtitle={t("training.manualSubtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          <TextField label={t("training.form.title")} value={title} onChangeText={setTitle} placeholder={t("training.form.titlePlaceholder")} />
          <TextField
            label={t("training.form.description")}
            value={description}
            onChangeText={setDescription}
            placeholder={t("training.form.descriptionPlaceholder")}
          />
          <PrimaryButton
            label={t("training.form.submit")}
            onPress={handleCreatePlan}
            loading={createTrainingPlan.isPending}
            disabled={!selectedPetId || !title}
          />
        </View>
      </SectionCard>

      <SectionCard title={t("training.aiTitle")} subtitle={t("training.aiSubtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          <TextField label={t("training.goal")} value={goal} onChangeText={setGoal} placeholder={t("training.goalPlaceholder")} />
          <View style={{ flexDirection: "row", flexWrap: "wrap", gap: theme.spacing.sm }}>
            {difficulties.map((item) => (
              <Pressable
                key={item}
                onPress={() => setDifficulty(item)}
                style={{
                  borderRadius: theme.radius.pill,
                  borderCurve: "continuous",
                  paddingHorizontal: 14,
                  paddingVertical: 10,
                  backgroundColor: difficulty === item ? palette.secondary : palette.surfaceMuted,
                }}
              >
                <Text selectable style={{ color: difficulty === item ? "#FFFFFF" : palette.textSecondary, fontWeight: "700" }}>
                  {getDifficultyLabel(item)}
                </Text>
              </Pressable>
            ))}
          </View>
          <PrimaryButton
            label={t("training.generate")}
            onPress={handleGeneratePlan}
            loading={generateTrainingPlan.isPending}
            disabled={!selectedPetId || !goal}
          />
        </View>
      </SectionCard>

      <SectionCard title={t("training.listTitle")} subtitle={t("training.listSubtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          {(trainingPlansQuery.data ?? []).length > 0 ? (
            trainingPlansQuery.data?.map((plan) => (
              <Pressable
                key={plan.id}
                onPress={() => router.push(`/training/${plan.id}`)}
                style={{
                  borderRadius: theme.radius.lg,
                  borderCurve: "continuous",
                  padding: theme.spacing.md,
                  backgroundColor: palette.surfaceMuted,
                  gap: theme.spacing.xs,
                }}
              >
                <Text selectable style={{ color: palette.text, fontWeight: "700" }}>
                  {plan.title}
                </Text>
                <Text selectable style={{ color: palette.textSecondary }}>
                  {t("training.list.meta", {
                    source: plan.ai_generated ? t("training.list.metaAi") : t("training.list.metaManual"),
                    progress: plan.progress,
                  })}
                </Text>
              </Pressable>
            ))
          ) : (
            <Text selectable style={{ color: palette.textSecondary }}>
              {t("training.list.empty")}
            </Text>
          )}
        </View>
      </SectionCard>
    </Screen>
  );
}
