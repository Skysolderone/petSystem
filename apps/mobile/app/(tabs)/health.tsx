import * as Haptics from "expo-haptics";
import { useEffect, useState } from "react";
import { Pressable, Text, View } from "react-native";

import { PrimaryButton } from "@/components/ui/primary-button";
import { Screen } from "@/components/ui/screen";
import { SectionCard } from "@/components/ui/section-card";
import { TextField } from "@/components/ui/text-field";
import { theme, useAppPalette } from "@/constants/theme";
import { useAskHealthAI, useCreateHealthRecord, useHealthAlerts, useHealthRecords, useHealthSummary } from "@/services/queries/use-health";
import { usePetsList } from "@/services/queries/use-pets";
import { usePetStore } from "@/stores/pet-store";
import { ApiError } from "@/types/api";
import { useI18n } from "@/utils/i18n";

const recordTypes = ["weight", "medication", "vaccination", "symptom"] as const;

export default function HealthScreen() {
  const palette = useAppPalette();
  const { t } = useI18n();
  const petsQuery = usePetsList();
  const pets = petsQuery.data?.pages.flatMap((page) => page.items) ?? [];
  const selectedPetId = usePetStore((state) => state.selectedPetId);
  const setSelectedPetId = usePetStore((state) => state.setSelectedPetId);
  const selectedPet = pets.find((pet) => pet.id === selectedPetId) ?? pets[0];
  const petId = selectedPet?.id;

  const recordsQuery = useHealthRecords(petId);
  const summaryQuery = useHealthSummary(petId);
  const alertsQuery = useHealthAlerts(petId);
  const createHealthRecord = useCreateHealthRecord(petId);
  const askHealthAI = useAskHealthAI(petId);

  const [recordType, setRecordType] = useState<(typeof recordTypes)[number]>("weight");
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [weightValue, setWeightValue] = useState("");
  const [question, setQuestion] = useState("");

  useEffect(() => {
    if (!selectedPetId && pets[0]?.id) {
      setSelectedPetId(pets[0].id);
    }
  }, [pets, selectedPetId, setSelectedPetId]);

  async function handleCreateRecord() {
    await createHealthRecord.mutateAsync({
      type: recordType,
      title,
      description,
      data: recordType === "weight" && weightValue ? { weight: Number(weightValue) } : undefined,
      recorded_at: new Date().toISOString(),
    });
    await Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success).catch(() => null);
    setTitle("");
    setDescription("");
    setWeightValue("");
  }

  async function handleAskAI() {
    await askHealthAI.mutateAsync(question);
    await Haptics.selectionAsync().catch(() => null);
  }

  function getRecordTypeLabel(type: (typeof recordTypes)[number]) {
    switch (type) {
      case "weight":
        return t("health.recordType.weight");
      case "medication":
        return t("health.recordType.medication");
      case "vaccination":
        return t("health.recordType.vaccination");
      case "symptom":
        return t("health.recordType.symptom");
    }
  }

  const errorMessage = createHealthRecord.error instanceof ApiError ? createHealthRecord.error.message : undefined;
  const aiErrorMessage = askHealthAI.error instanceof ApiError ? askHealthAI.error.message : undefined;

  return (
    <Screen>
      <SectionCard title={t("health.selectTitle")} subtitle={t("health.selectSubtitle")}>
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
                  borderColor: selectedPet?.id === pet.id ? palette.primary : palette.border,
                  backgroundColor: selectedPet?.id === pet.id ? palette.primaryLight : palette.surface,
                }}
              >
                <Text selectable style={{ color: palette.text, fontWeight: "700" }}>
                  {pet.name}
                </Text>
              </Pressable>
            ))}
          </View>
        ) : (
          <Text selectable style={{ color: palette.textSecondary }}>
            {t("health.noPets")}
          </Text>
        )}
      </SectionCard>

      <SectionCard
        title={t("health.summaryTitle", { score: summaryQuery.data?.score ?? "--" })}
        subtitle={t("health.summarySubtitle", {
          status: summaryQuery.data?.status ?? t("health.summaryPending"),
          count: summaryQuery.data?.data_points_analyzed ?? 0,
        })}
      >
        <View style={{ gap: theme.spacing.sm }}>
          {(summaryQuery.data?.insights ?? []).slice(0, 3).map((insight) => (
            <Text key={insight} selectable style={{ color: palette.textSecondary, lineHeight: 22 }}>
              {insight}
            </Text>
          ))}
          {(summaryQuery.data?.recommended_actions ?? []).slice(0, 2).map((action) => (
            <Text key={action} selectable style={{ color: palette.primary, lineHeight: 22 }}>
              {t("health.summaryActionPrefix", { action })}
            </Text>
          ))}
        </View>
      </SectionCard>

      <SectionCard title={t("health.quickRecordTitle")} subtitle={t("health.quickRecordSubtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          <View style={{ flexDirection: "row", flexWrap: "wrap", gap: theme.spacing.sm }}>
            {recordTypes.map((type) => (
              <Pressable
                key={type}
                onPress={() => setRecordType(type)}
                style={{
                  borderRadius: theme.radius.pill,
                  borderCurve: "continuous",
                  paddingHorizontal: 14,
                  paddingVertical: 10,
                  backgroundColor: recordType === type ? palette.primary : palette.surfaceMuted,
                }}
              >
                <Text selectable style={{ color: recordType === type ? "#FFFFFF" : palette.textSecondary, fontWeight: "600" }}>
                  {getRecordTypeLabel(type)}
                </Text>
              </Pressable>
            ))}
          </View>
          <TextField label={t("health.form.title")} value={title} onChangeText={setTitle} placeholder={t("health.form.titlePlaceholder")} />
          <TextField
            label={t("health.form.description")}
            value={description}
            onChangeText={setDescription}
            placeholder={t("health.form.descriptionPlaceholder")}
          />
          {recordType === "weight" ? (
            <TextField
              label={t("health.form.weight")}
              value={weightValue}
              onChangeText={setWeightValue}
              placeholder={t("health.form.weightPlaceholder")}
              keyboardType="numeric"
            />
          ) : null}
          {errorMessage ? (
            <Text selectable style={{ color: palette.error }}>
              {errorMessage}
            </Text>
          ) : null}
          <PrimaryButton
            label={t("health.form.submit")}
            onPress={handleCreateRecord}
            loading={createHealthRecord.isPending}
            disabled={!petId || !title}
          />
        </View>
      </SectionCard>

      <SectionCard title={t("health.recordsTitle")} subtitle={t("health.recordsSubtitle")}>
        <View style={{ gap: theme.spacing.sm }}>
          {(recordsQuery.data ?? []).length > 0 ? (
            recordsQuery.data?.map((record) => (
              <View
                key={record.id}
                style={{
                  borderRadius: theme.radius.md,
                  borderCurve: "continuous",
                  backgroundColor: palette.surfaceMuted,
                  padding: theme.spacing.md,
                  gap: theme.spacing.xs,
                }}
              >
                <Text selectable style={{ color: palette.text, fontWeight: "700" }}>
                  {record.title}
                </Text>
                <Text selectable style={{ color: palette.textSecondary }}>
                  {getRecordTypeLabel(record.type as (typeof recordTypes)[number])} · {new Date(record.recorded_at).toLocaleDateString()}
                </Text>
                {record.description ? (
                  <Text selectable style={{ color: palette.textSecondary }}>
                    {record.description}
                  </Text>
                ) : null}
              </View>
            ))
          ) : (
            <Text selectable style={{ color: palette.textSecondary }}>
              {t("health.recordsEmpty")}
            </Text>
          )}
        </View>
      </SectionCard>

      <SectionCard title={t("health.alertsTitle")} subtitle={t("health.alertsSubtitle")}>
        <View style={{ gap: theme.spacing.sm }}>
          {(alertsQuery.data ?? []).length > 0 ? (
            alertsQuery.data?.map((alert) => (
              <View
                key={alert.id}
                style={{
                  borderRadius: theme.radius.md,
                  borderCurve: "continuous",
                  padding: theme.spacing.md,
                  backgroundColor: alert.severity === "high" ? "#FEF2F2" : palette.surfaceMuted,
                  borderWidth: 1,
                  borderColor: alert.severity === "high" ? "#FCA5A5" : palette.border,
                  gap: theme.spacing.xs,
                }}
              >
                <Text selectable style={{ color: palette.text, fontWeight: "700" }}>
                  {alert.title}
                </Text>
                <Text selectable style={{ color: palette.textSecondary }}>
                  {alert.message}
                </Text>
              </View>
            ))
          ) : (
            <Text selectable style={{ color: palette.textSecondary }}>
              {t("health.alertsEmpty")}
            </Text>
          )}
        </View>
      </SectionCard>

      <SectionCard title={t("health.aiTitle")} subtitle={t("health.aiSubtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          <TextField
            label={t("health.aiQuestion")}
            value={question}
            onChangeText={setQuestion}
            placeholder={t("health.aiQuestionPlaceholder")}
          />
          {aiErrorMessage ? (
            <Text selectable style={{ color: palette.error }}>
              {aiErrorMessage}
            </Text>
          ) : null}
          <PrimaryButton label={t("health.aiSubmit")} onPress={handleAskAI} loading={askHealthAI.isPending} disabled={!petId || !question} />
          {askHealthAI.data?.answer ? (
            <Text selectable style={{ color: palette.textSecondary, lineHeight: 22 }}>
              {askHealthAI.data.answer}
            </Text>
          ) : null}
        </View>
      </SectionCard>
    </Screen>
  );
}
