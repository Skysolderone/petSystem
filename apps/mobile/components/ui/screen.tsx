import { ReactNode } from "react";
import { ScrollView, View } from "react-native";

import { theme, useAppPalette } from "@/constants/theme";

interface ScreenProps {
  children: ReactNode;
}

export function Screen({ children }: ScreenProps) {
  const palette = useAppPalette();

  return (
    <ScrollView
      contentInsetAdjustmentBehavior="automatic"
      style={{ flex: 1, backgroundColor: palette.background }}
      contentContainerStyle={{
        padding: theme.spacing.lg,
        gap: theme.spacing.md,
      }}
    >
      <View style={{ gap: theme.spacing.md }}>{children}</View>
    </ScrollView>
  );
}
