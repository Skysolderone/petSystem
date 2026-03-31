import * as Haptics from "expo-haptics";
import { Stack, useLocalSearchParams } from "expo-router";
import { Text, View } from "react-native";

import { PrimaryButton } from "@/components/ui/primary-button";
import { Screen } from "@/components/ui/screen";
import { SectionCard } from "@/components/ui/section-card";
import { theme, useAppPalette } from "@/constants/theme";
import { useTrainingPlanDetail, useUpdateTrainingPlan } from "@/services/queries/use-training";
import { useI18n } from "@/utils/i18n";

const progressPresets = [0, 25, 50, 75, 100];

export default function TrainingPlanDetailScreen() {
  const palette = useAppPalette();
  const { t } = useI18n();
  const { planId } = useLocalSearchParams<{ planId: string }>();
  const trainingPlanQuery = useTrainingPlanDetail(planId);
  const updateTrainingPlan = useUpdateTrainingPlan(planId);
  const plan = trainingPlanQuery.data;

  async function handleUpdateProgress(progress: number) {
    await updateTrainingPlan.mutateAsync({ progress });
    await Haptics.selectionAsync().catch(() => null);
  }

  function getDifficultyLabel(value?: string) {
    switch (value) {
      case "beginner":
        return t("training.difficulty.beginner");
      case "intermediate":
        return t("training.difficulty.intermediate");
      case "advanced":
        return t("training.difficulty.advanced");
      default:
        return value ?? "--";
    }
  }

  return (
    <Screen>
      <Stack.Screen options={{ title: plan?.title ?? t("training.navTitle") }} />

      <SectionCard title={plan?.title ?? t("training.detail.loading")} subtitle={plan?.description ?? t("training.detail.loadingSubtitle")}>
        <View style={{ gap: theme.spacing.xs }}>
          <Text selectable style={{ color: palette.textSecondary }}>
            {t("training.detail.summaryMeta", {
              category: plan?.category ?? "--",
              difficulty: getDifficultyLabel(plan?.difficulty),
            })}
          </Text>
          <Text selectable style={{ color: palette.primary, fontWeight: "700" }}>
            {t("training.detail.progress", { progress: plan?.progress ?? 0 })}
          </Text>
        </View>
      </SectionCard>

      <SectionCard title={t("training.detail.updateTitle")} subtitle={t("training.detail.updateSubtitle")}>
        <View style={{ gap: theme.spacing.sm }}>
          {progressPresets.map((progress) => (
            <PrimaryButton
              key={progress}
              label={t("training.detail.progressPreset", { progress })}
              onPress={() => handleUpdateProgress(progress)}
              variant={plan?.progress === progress ? "primary" : "secondary"}
              loading={updateTrainingPlan.isPending && plan?.progress !== progress}
            />
          ))}
        </View>
      </SectionCard>

      <SectionCard title={t("training.detail.stepsTitle")} subtitle={t("training.detail.stepsSubtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          {(plan?.steps ?? []).map((step, index) => (
            <View
              key={`${step.title ?? "step"}-${index}`}
              style={{
                borderRadius: theme.radius.md,
                borderCurve: "continuous",
                padding: theme.spacing.md,
                backgroundColor: palette.surfaceMuted,
                gap: theme.spacing.xs,
              }}
            >
              <Text selectable style={{ color: palette.text, fontWeight: "700" }}>
                {String(step.title ?? t("training.detail.stepFallback", { index: index + 1 }))}
              </Text>
              <Text selectable style={{ color: palette.textSecondary }}>
                {String(step.instruction ?? t("training.detail.stepInstructionFallback"))}
              </Text>
              <Text selectable style={{ color: palette.textSecondary }}>
                {t("training.detail.duration", { duration: String(step.duration_minutes ?? "--") })}
              </Text>
            </View>
          ))}
        </View>
      </SectionCard>
    </Screen>
  );
}
