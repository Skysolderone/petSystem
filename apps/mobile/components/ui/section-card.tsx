import { ReactNode } from "react";
import { Text, View } from "react-native";

import { theme, useAppPalette } from "@/constants/theme";

interface SectionCardProps {
  title: string;
  subtitle?: string;
  children?: ReactNode;
}

export function SectionCard({ title, subtitle, children }: SectionCardProps) {
  const palette = useAppPalette();

  return (
    <View
      style={{
        borderRadius: theme.radius.lg,
        borderCurve: "continuous",
        padding: theme.spacing.lg,
        gap: theme.spacing.md,
        backgroundColor: palette.surface,
        borderWidth: 1,
        borderColor: palette.border,
        boxShadow: "0 10px 30px rgba(15, 23, 42, 0.06)",
      }}
    >
      <View style={{ gap: theme.spacing.xs }}>
        <Text
          selectable
          style={{
            color: palette.text,
            fontSize: theme.fontSize.lg,
            fontWeight: "700",
          }}
        >
          {title}
        </Text>
        {subtitle ? (
          <Text
            selectable
            style={{
              color: palette.textSecondary,
              lineHeight: 22,
            }}
          >
            {subtitle}
          </Text>
        ) : null}
      </View>
      {children}
    </View>
  );
}
