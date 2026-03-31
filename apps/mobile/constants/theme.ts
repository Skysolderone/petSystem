import { useColorScheme } from "react-native";

import { useSettingsStore } from "@/stores/settings-store";

const lightColors = {
  primary: "#10B981",
  primaryLight: "#D1FAE5",
  secondary: "#0F766E",
  accent: "#F97316",
  social: "#EC4899",
  info: "#2563EB",
  background: "#F6FBF8",
  surface: "#FFFFFF",
  surfaceMuted: "#ECFDF5",
  border: "#D1FAE5",
  text: "#0F172A",
  textSecondary: "#475569",
  textMuted: "#94A3B8",
  error: "#DC2626",
};

const darkColors = {
  primary: "#34D399",
  primaryLight: "#064E3B",
  secondary: "#5EEAD4",
  accent: "#FB923C",
  social: "#F472B6",
  info: "#60A5FA",
  background: "#061310",
  surface: "#0F1F1B",
  surfaceMuted: "#123128",
  border: "#1E463B",
  text: "#F8FAFC",
  textSecondary: "#CBD5E1",
  textMuted: "#94A3B8",
  error: "#F87171",
};

export const theme = {
  spacing: { xs: 4, sm: 8, md: 16, lg: 24, xl: 32 },
  radius: { sm: 10, md: 16, lg: 24, pill: 999 },
  fontSize: { sm: 13, md: 16, lg: 20, xl: 28, hero: 36 },
  light: lightColors,
  dark: darkColors,
} as const;

export type AppPalette = typeof lightColors;

export function useAppPalette(): AppPalette {
  const colorScheme = useColorScheme();
  const themePreference = useSettingsStore((state) => state.themePreference);

  if (themePreference === "light") {
    return lightColors;
  }
  if (themePreference === "dark") {
    return darkColors;
  }
  return colorScheme === "dark" ? darkColors : lightColors;
}
