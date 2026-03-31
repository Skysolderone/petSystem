import * as SecureStore from "expo-secure-store";

export type ThemePreference = "system" | "light" | "dark";
export type AppLanguage = "zh" | "en" | "ja";

export interface AppSettings {
  themePreference: ThemePreference;
  language: AppLanguage;
  notificationsEnabled: boolean;
}

const settingsKey = "petverse.app.settings";

const defaultSettings: AppSettings = {
  themePreference: "system",
  language: "zh",
  notificationsEnabled: true,
};

export async function getStoredSettings(): Promise<AppSettings> {
  const rawValue = await SecureStore.getItemAsync(settingsKey);
  if (!rawValue) {
    return defaultSettings;
  }

  try {
    return {
      ...defaultSettings,
      ...(JSON.parse(rawValue) as Partial<AppSettings>),
    };
  } catch {
    await SecureStore.deleteItemAsync(settingsKey);
    return defaultSettings;
  }
}

export async function setStoredSettings(settings: AppSettings): Promise<void> {
  await SecureStore.setItemAsync(settingsKey, JSON.stringify(settings));
}

export { defaultSettings };
