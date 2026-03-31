import { router } from "expo-router";
import * as Haptics from "expo-haptics";
import * as ImagePicker from "expo-image-picker";
import { useEffect, useState } from "react";
import { Text, View } from "react-native";

import { PrimaryButton } from "@/components/ui/primary-button";
import { Screen } from "@/components/ui/screen";
import { SectionCard } from "@/components/ui/section-card";
import { TextField } from "@/components/ui/text-field";
import { theme, useAppPalette } from "@/constants/theme";
import { useUserMe, useUpdateUserLocation, useUpdateUserProfile, useUploadUserAvatar } from "@/services/queries/use-user";
import { useAuthStore } from "@/stores/auth-store";
import { useNotificationStore } from "@/stores/notification-store";
import { useSettingsStore } from "@/stores/settings-store";
import { useI18n } from "@/utils/i18n";
import { getStoredPushToken, clearStoredPushToken } from "@/utils/push-token";
import { unregisterPushToken } from "@/services/queries/use-notifications";

export default function ProfileScreen() {
  const palette = useAppPalette();
  const { t } = useI18n();
  const session = useAuthStore((state) => state.session);
  const logout = useAuthStore((state) => state.logout);
  const unreadCount = useNotificationStore((state) => state.unreadCount);
  const meQuery = useUserMe();
  const updateProfile = useUpdateUserProfile();
  const updateLocation = useUpdateUserLocation();
  const uploadUserAvatar = useUploadUserAvatar();
  const themePreference = useSettingsStore((state) => state.themePreference);
  const language = useSettingsStore((state) => state.language);
  const notificationsEnabled = useSettingsStore((state) => state.notificationsEnabled);
  const setThemePreference = useSettingsStore((state) => state.setThemePreference);
  const setLanguage = useSettingsStore((state) => state.setLanguage);
  const setNotificationsEnabled = useSettingsStore((state) => state.setNotificationsEnabled);
  const [nickname, setNickname] = useState("");
  const [email, setEmail] = useState("");
  const [latitude, setLatitude] = useState("");
  const [longitude, setLongitude] = useState("");

  useEffect(() => {
    if (meQuery.data) {
      setNickname(meQuery.data.nickname ?? "");
      setEmail(meQuery.data.email ?? "");
      setLatitude(meQuery.data.latitude !== undefined ? String(meQuery.data.latitude) : "");
      setLongitude(meQuery.data.longitude !== undefined ? String(meQuery.data.longitude) : "");
    }
  }, [meQuery.data]);

  async function handleLogout() {
    const pushToken = await getStoredPushToken().catch(() => null);
    if (pushToken) {
      await unregisterPushToken(pushToken).catch(() => null);
      await clearStoredPushToken().catch(() => null);
    }
    await logout();
    router.replace("/auth/login");
  }

  async function handleUpdateProfile() {
    await updateProfile.mutateAsync({
      nickname,
      email,
    });
    await Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success).catch(() => null);
  }

  async function handleUpdateLocation() {
    await updateLocation.mutateAsync({
      latitude: latitude ? Number(latitude) : undefined,
      longitude: longitude ? Number(longitude) : undefined,
    });
    await Haptics.selectionAsync().catch(() => null);
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
    await uploadUserAvatar.mutateAsync({
      uri: asset.uri,
      name: asset.fileName ?? "avatar.jpg",
      type: asset.mimeType ?? "image/jpeg",
    });
    await Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success).catch(() => null);
  }

  async function handleThemeChange(nextTheme: "system" | "light" | "dark") {
    await setThemePreference(nextTheme);
    await Haptics.selectionAsync().catch(() => null);
  }

  async function handleLanguageChange(nextLanguage: "zh" | "en" | "ja") {
    await setLanguage(nextLanguage);
    await Haptics.selectionAsync().catch(() => null);
  }

  async function handleNotificationsChange(enabled: boolean) {
    await setNotificationsEnabled(enabled);
    await Haptics.selectionAsync().catch(() => null);
  }

  const themeOptions = [
    { label: t("profile.themeSystem"), value: "system" as const },
    { label: t("profile.themeLight"), value: "light" as const },
    { label: t("profile.themeDark"), value: "dark" as const },
  ];

  const languageOptions = [
    { label: t("profile.languageZh"), value: "zh" as const },
    { label: t("profile.languageEn"), value: "en" as const },
    { label: t("profile.languageJa"), value: "ja" as const },
  ];

  return (
    <Screen>
      <SectionCard title={session?.user.nickname ?? t("profile.notLoggedIn")} subtitle={session?.user.phone ?? t("profile.noPhone")}>
        <View style={{ gap: theme.spacing.sm }}>
          <Text selectable style={{ color: palette.textSecondary }}>
            {t("profile.role", { value: session?.user.role ?? "guest" })}
          </Text>
          <Text selectable style={{ color: palette.textSecondary }}>
            {t("profile.plan", { value: session?.user.plan_type ?? "free" })}
          </Text>
        </View>
      </SectionCard>

      <SectionCard title={t("profile.editTitle")} subtitle={t("profile.editSubtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          <TextField label={t("profile.nickname")} value={nickname} onChangeText={setNickname} placeholder={t("profile.nicknamePlaceholder")} />
          <TextField label={t("profile.email")} value={email} onChangeText={setEmail} placeholder={t("profile.emailPlaceholder")} />
          <PrimaryButton
            label={t("profile.saveProfile")}
            onPress={handleUpdateProfile}
            loading={updateProfile.isPending}
            disabled={!nickname}
          />
          <PrimaryButton
            label={t("profile.uploadAvatar")}
            onPress={handleUploadAvatar}
            loading={uploadUserAvatar.isPending}
            variant="secondary"
          />
          <TextField label={t("profile.latitude")} value={latitude} onChangeText={setLatitude} placeholder={t("profile.latitudePlaceholder")} keyboardType="numeric" />
          <TextField label={t("profile.longitude")} value={longitude} onChangeText={setLongitude} placeholder={t("profile.longitudePlaceholder")} keyboardType="numeric" />
          <PrimaryButton
            label={t("profile.updateLocation")}
            onPress={handleUpdateLocation}
            loading={updateLocation.isPending}
            disabled={!latitude || !longitude}
            variant="ghost"
          />
        </View>
      </SectionCard>

      <SectionCard title={t("profile.settingsTitle")} subtitle={t("profile.settingsSubtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          <View style={{ gap: theme.spacing.sm }}>
            <Text selectable style={{ color: palette.textSecondary, fontWeight: "600" }}>
              {t("profile.theme")}
            </Text>
            <View style={{ flexDirection: "row", flexWrap: "wrap", gap: theme.spacing.sm }}>
              {themeOptions.map((option) => (
                <PrimaryButton
                  key={option.value}
                  label={option.label}
                  onPress={() => handleThemeChange(option.value)}
                  variant={themePreference === option.value ? "primary" : "secondary"}
                />
              ))}
            </View>
          </View>
          <View style={{ gap: theme.spacing.sm }}>
            <Text selectable style={{ color: palette.textSecondary, fontWeight: "600" }}>
              {t("profile.language")}
            </Text>
            <View style={{ flexDirection: "row", flexWrap: "wrap", gap: theme.spacing.sm }}>
              {languageOptions.map((option) => (
                <PrimaryButton
                  key={option.value}
                  label={option.label}
                  onPress={() => handleLanguageChange(option.value)}
                  variant={language === option.value ? "primary" : "secondary"}
                />
              ))}
            </View>
          </View>
          <View style={{ gap: theme.spacing.sm }}>
            <Text selectable style={{ color: palette.textSecondary, fontWeight: "600" }}>
              {t("profile.notifications")}
            </Text>
            <View style={{ flexDirection: "row", gap: theme.spacing.sm }}>
              <PrimaryButton
                label={t("profile.enable")}
                onPress={() => handleNotificationsChange(true)}
                variant={notificationsEnabled ? "primary" : "secondary"}
              />
              <PrimaryButton
                label={t("profile.disable")}
                onPress={() => handleNotificationsChange(false)}
                variant={!notificationsEnabled ? "primary" : "secondary"}
              />
            </View>
          </View>
        </View>
      </SectionCard>

      <SectionCard title={t("profile.manageTitle")} subtitle={t("profile.manageSubtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          <PrimaryButton label={t("profile.myPets")} onPress={() => router.push("/pet")} />
          <PrimaryButton label={t("profile.myDevices")} onPress={() => router.push("/device")} variant="secondary" />
          <PrimaryButton label={t("profile.training")} onPress={() => router.push("/training")} variant="secondary" />
          <PrimaryButton label={t("profile.shop")} onPress={() => router.push("/shop")} variant="secondary" />
          <PrimaryButton label={t("profile.notificationsCenter", { count: unreadCount })} onPress={() => router.push("/notifications")} variant="ghost" />
          <PrimaryButton label={t("profile.logout")} onPress={handleLogout} variant="secondary" />
        </View>
      </SectionCard>
    </Screen>
  );
}
