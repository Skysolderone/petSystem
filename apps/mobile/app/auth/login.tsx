import * as Haptics from "expo-haptics";
import { Link, router } from "expo-router";
import { useState } from "react";
import { Text, View } from "react-native";

import { PrimaryButton } from "@/components/ui/primary-button";
import { Screen } from "@/components/ui/screen";
import { SectionCard } from "@/components/ui/section-card";
import { TextField } from "@/components/ui/text-field";
import { theme, useAppPalette } from "@/constants/theme";
import { useAppleLogin, useGoogleLogin, useLogin, useWechatLogin } from "@/services/queries/use-auth";
import { ApiError } from "@/types/api";
import { useI18n } from "@/utils/i18n";
import { getSocialLoginSeed } from "@/utils/social-login";

export default function LoginScreen() {
  const palette = useAppPalette();
  const { t } = useI18n();
  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const loginMutation = useLogin();
  const wechatLoginMutation = useWechatLogin();
  const appleLoginMutation = useAppleLogin();
  const googleLoginMutation = useGoogleLogin();

  const errorMessage =
    loginMutation.error instanceof ApiError
      ? loginMutation.error.message
      : wechatLoginMutation.error instanceof ApiError
        ? wechatLoginMutation.error.message
        : appleLoginMutation.error instanceof ApiError
          ? appleLoginMutation.error.message
          : googleLoginMutation.error instanceof ApiError
            ? googleLoginMutation.error.message
          : undefined;

  async function handleLogin() {
    await loginMutation.mutateAsync({ phone, password });
    await Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success).catch(() => null);
    router.replace("/(tabs)");
  }

  async function handleWechatLogin() {
    const openId = await getSocialLoginSeed("wechat");
    await wechatLoginMutation.mutateAsync({
      open_id: openId,
      nickname: t("auth.login.wechatDemoName"),
    });
    await Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success).catch(() => null);
    router.replace("/(tabs)");
  }

  async function handleAppleLogin() {
    const appleId = await getSocialLoginSeed("apple");
    await appleLoginMutation.mutateAsync({
      apple_id: appleId,
      nickname: t("auth.login.appleDemoName"),
      email: `${appleId}@petverse.local`,
    });
    await Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success).catch(() => null);
    router.replace("/(tabs)");
  }

  async function handleGoogleLogin() {
    const googleId = await getSocialLoginSeed("google");
    await googleLoginMutation.mutateAsync({
      google_id: googleId,
      nickname: t("auth.login.googleDemoName"),
      email: `${googleId}@petverse.local`,
    });
    await Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success).catch(() => null);
    router.replace("/(tabs)");
  }

  return (
    <Screen>
      <SectionCard
        title={t("auth.login.title")}
        subtitle={t("auth.login.subtitle")}
      >
        <View style={{ gap: theme.spacing.md }}>
          <TextField
            label={t("auth.login.phone")}
            value={phone}
            onChangeText={setPhone}
            placeholder="13800000000"
            keyboardType="phone-pad"
          />
          <TextField
            label={t("auth.login.password")}
            value={password}
            onChangeText={setPassword}
            placeholder={t("auth.login.passwordPlaceholder")}
            secureTextEntry
          />
          {errorMessage ? (
            <Text selectable style={{ color: palette.error }}>
              {errorMessage}
            </Text>
          ) : null}
          <PrimaryButton
            label={t("auth.login.submit")}
            onPress={handleLogin}
            loading={loginMutation.isPending}
            disabled={!phone || !password}
          />
          <PrimaryButton
            label={t("auth.login.wechat")}
            onPress={handleWechatLogin}
            loading={wechatLoginMutation.isPending}
            variant="secondary"
          />
          <PrimaryButton
            label={t("auth.login.apple")}
            onPress={handleAppleLogin}
            loading={appleLoginMutation.isPending}
            variant="secondary"
          />
          <PrimaryButton
            label={t("auth.login.google")}
            onPress={handleGoogleLogin}
            loading={googleLoginMutation.isPending}
            variant="ghost"
          />
          <Link href="/auth/register" asChild>
            <Text
              selectable
              style={{
                color: palette.primary,
                textAlign: "center",
                fontWeight: "600",
              }}
            >
              {t("auth.login.registerLink")}
            </Text>
          </Link>
        </View>
      </SectionCard>
    </Screen>
  );
}
