import { ActivityIndicator, Pressable, Text } from "react-native";

import { theme, useAppPalette } from "@/constants/theme";

interface PrimaryButtonProps {
  label: string;
  onPress: () => void | Promise<void>;
  disabled?: boolean;
  loading?: boolean;
  variant?: "primary" | "secondary" | "ghost";
}

export function PrimaryButton({
  label,
  onPress,
  disabled = false,
  loading = false,
  variant = "primary",
}: PrimaryButtonProps) {
  const palette = useAppPalette();

  const backgroundColor =
    variant === "secondary" ? palette.surfaceMuted : variant === "ghost" ? "transparent" : palette.primary;
  const borderWidth = variant === "ghost" ? 0 : 1;
  const textColor = variant === "ghost" ? palette.primary : variant === "secondary" ? palette.text : "#FFFFFF";

  return (
    <Pressable
      onPress={() => {
        void onPress();
      }}
      disabled={disabled || loading}
      style={({ pressed }) => ({
        borderRadius: theme.radius.md,
        borderCurve: "continuous",
        backgroundColor,
        borderWidth,
        borderColor: variant === "ghost" ? "transparent" : palette.border,
        paddingVertical: 15,
        paddingHorizontal: 18,
        alignItems: "center",
        justifyContent: "center",
        opacity: disabled ? 0.55 : pressed ? 0.85 : 1,
      })}
    >
      {loading ? (
        <ActivityIndicator color={textColor} />
      ) : (
        <Text
          selectable
          style={{
            color: textColor,
            fontWeight: "700",
            fontSize: theme.fontSize.md,
          }}
        >
          {label}
        </Text>
      )}
    </Pressable>
  );
}
