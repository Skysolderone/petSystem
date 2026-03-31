import * as Haptics from "expo-haptics";
import { router } from "expo-router";
import { useState } from "react";
import { Text, View } from "react-native";

import { PrimaryButton } from "@/components/ui/primary-button";
import { Screen } from "@/components/ui/screen";
import { SectionCard } from "@/components/ui/section-card";
import { TextField } from "@/components/ui/text-field";
import { theme, useAppPalette } from "@/constants/theme";
import { useRegister } from "@/services/queries/use-auth";
import { ApiError } from "@/types/api";
import { useI18n } from "@/utils/i18n";

export default function RegisterScreen() {
  const palette = useAppPalette();
  const { t } = useI18n();
  const [nickname, setNickname] = useState("");
  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const registerMutation = useRegister();

  const errorMessage =
    registerMutation.error instanceof ApiError ? registerMutation.error.message : undefined;

  async function handleRegister() {
    await registerMutation.mutateAsync({ nickname, phone, password });
    await Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success).catch(() => null);
    router.replace("/(tabs)");
  }

  return (
    <Screen>
      <SectionCard title={t("auth.register.title")} subtitle={t("auth.register.subtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          <TextField
            label={t("auth.register.nickname")}
            value={nickname}
            onChangeText={setNickname}
            placeholder={t("auth.register.nicknamePlaceholder")}
          />
          <TextField
            label={t("auth.register.phone")}
            value={phone}
            onChangeText={setPhone}
            placeholder="13800000000"
            keyboardType="phone-pad"
          />
          <TextField
            label={t("auth.register.password")}
            value={password}
            onChangeText={setPassword}
            placeholder={t("auth.register.passwordPlaceholder")}
            secureTextEntry
          />
          {errorMessage ? (
            <Text selectable style={{ color: palette.error }}>
              {errorMessage}
            </Text>
          ) : null}
          <PrimaryButton
            label={t("auth.register.submit")}
            onPress={handleRegister}
            loading={registerMutation.isPending}
            disabled={!nickname || !phone || !password}
          />
        </View>
      </SectionCard>
    </Screen>
  );
}
