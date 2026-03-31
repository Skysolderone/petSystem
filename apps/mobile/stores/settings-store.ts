import { create } from "zustand";

import type { AppLanguage, ThemePreference } from "@/utils/settings";
import { defaultSettings, getStoredSettings, setStoredSettings } from "@/utils/settings";

interface SettingsState {
  isHydrated: boolean;
  themePreference: ThemePreference;
  language: AppLanguage;
  notificationsEnabled: boolean;
  hydrate: () => Promise<void>;
  setThemePreference: (themePreference: ThemePreference) => Promise<void>;
  setLanguage: (language: AppLanguage) => Promise<void>;
  setNotificationsEnabled: (notificationsEnabled: boolean) => Promise<void>;
}

async function persistSettings(nextState: Pick<SettingsState, "themePreference" | "language" | "notificationsEnabled">) {
  await setStoredSettings({
    themePreference: nextState.themePreference,
    language: nextState.language,
    notificationsEnabled: nextState.notificationsEnabled,
  });
}

export const useSettingsStore = create<SettingsState>((set, get) => ({
  isHydrated: false,
  themePreference: defaultSettings.themePreference,
  language: defaultSettings.language,
  notificationsEnabled: defaultSettings.notificationsEnabled,
  hydrate: async () => {
    const settings = await getStoredSettings();
    set({
      isHydrated: true,
      ...settings,
    });
  },
  setThemePreference: async (themePreference) => {
    const nextState = {
      themePreference,
      language: get().language,
      notificationsEnabled: get().notificationsEnabled,
    };
    await persistSettings(nextState);
    set({ themePreference });
  },
  setLanguage: async (language) => {
    const nextState = {
      themePreference: get().themePreference,
      language,
      notificationsEnabled: get().notificationsEnabled,
    };
    await persistSettings(nextState);
    set({ language });
  },
  setNotificationsEnabled: async (notificationsEnabled) => {
    const nextState = {
      themePreference: get().themePreference,
      language: get().language,
      notificationsEnabled,
    };
    await persistSettings(nextState);
    set({ notificationsEnabled });
  },
}));
