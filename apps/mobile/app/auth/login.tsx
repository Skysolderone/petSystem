import * as Haptics from "expo-haptics";
import * as AppleAuthentication from "expo-apple-authentication";
import * as Google from "expo-auth-session/providers/google";
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
import {
  formatAppleDisplayName,
  getGoogleNativeConfig,
  getSocialLoginSeed,
  hasGoogleNativeConfig,
} from "@/utils/social-login";

export default function LoginScreen() {
  const palette = useAppPalette();
  const { t } = useI18n();
  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [oauthError, setOauthError] = useState<string>();
  const loginMutation = useLogin();
  const wechatLoginMutation = useWechatLogin();
  const appleLoginMutation = useAppleLogin();
  const googleLoginMutation = useGoogleLogin();
  const googleNativeConfig = getGoogleNativeConfig();
  const [googleRequest, , promptGoogle] = Google.useIdTokenAuthRequest(
    {
      webClientId: googleNativeConfig.webClientId ?? "petverse-placeholder-web.apps.googleusercontent.com",
      iosClientId: googleNativeConfig.iosClientId ?? "petverse-placeholder-ios.apps.googleusercontent.com",
      androidClientId: googleNativeConfig.androidClientId ?? "petverse-placeholder-android.apps.googleusercontent.com",
    },
    {
      native: "petverse:/oauthredirect",
    },
  );

  const errorMessage =
    loginMutation.error instanceof ApiError
      ? loginMutation.error.message
      : wechatLoginMutation.error instanceof ApiError
        ? wechatLoginMutation.error.message
        : appleLoginMutation.error instanceof ApiError
          ? appleLoginMutation.error.message
          : googleLoginMutation.error instanceof ApiError
            ? googleLoginMutation.error.message
            : oauthError;

  async function handleLogin() {
    setOauthError(undefined);
    await loginMutation.mutateAsync({ phone, password });
    await finishLogin();
  }

  async function handleWechatLogin() {
    setOauthError(undefined);
    const openId = await getSocialLoginSeed("wechat");
    await wechatLoginMutation.mutateAsync({
      open_id: openId,
      nickname: t("auth.login.wechatDemoName"),
    });
    await finishLogin();
  }

  async function handleAppleLogin() {
    setOauthError(undefined);

    const available = await AppleAuthentication.isAvailableAsync().catch(() => false);
    if (!available) {
      await loginWithAppleDemo();
      return;
    }

    let credential: AppleAuthentication.AppleAuthenticationCredential;
    try {
      credential = await AppleAuthentication.signInAsync({
        requestedScopes: [
          AppleAuthentication.AppleAuthenticationScope.FULL_NAME,
          AppleAuthentication.AppleAuthenticationScope.EMAIL,
        ],
      });
    } catch (error) {
      if (isCanceledError(error)) {
        return;
      }
      await loginWithAppleDemo();
      return;
    }

    if (!credential.identityToken) {
      await loginWithAppleDemo();
      return;
    }

    await appleLoginMutation.mutateAsync({
      apple_id: credential.user,
      identity_token: credential.identityToken,
      nickname: formatAppleDisplayName(credential.fullName),
      email: credential.email ?? undefined,
    });
    await finishLogin();
  }

  async function handleGoogleLogin() {
    setOauthError(undefined);

    if (!hasGoogleNativeConfig()) {
      await loginWithGoogleDemo();
      return;
    }
    if (!googleRequest) {
      setOauthError(t("auth.login.socialError"));
      return;
    }

    try {
      const result = await promptGoogle();
      if (result.type !== "success") {
        return;
      }

      const identityToken = result.params.id_token || result.authentication?.idToken;
      if (!identityToken) {
        throw new Error("missing Google identity token");
      }

      await googleLoginMutation.mutateAsync({
        identity_token: identityToken,
      });
      await finishLogin();
    } catch (error) {
      if (isCanceledError(error)) {
        return;
      }
      setOauthError(error instanceof Error ? error.message : t("auth.login.socialError"));
    }
  }

  async function loginWithAppleDemo() {
    const appleId = await getSocialLoginSeed("apple");
    await appleLoginMutation.mutateAsync({
      apple_id: appleId,
      nickname: t("auth.login.appleDemoName"),
      email: `${appleId}@petverse.local`,
    });
    await finishLogin();
  }

  async function loginWithGoogleDemo() {
    const googleId = await getSocialLoginSeed("google");
    await googleLoginMutation.mutateAsync({
      google_id: googleId,
      nickname: t("auth.login.googleDemoName"),
      email: `${googleId}@petverse.local`,
    });
    await finishLogin();
  }

  async function finishLogin() {
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

function isCanceledError(error: unknown): boolean {
  return (
    error instanceof Error &&
    (error.message.includes("cancel") || error.message.includes("canceled"))
  );
}
