import { Text, TextInput, View } from "react-native";

import { theme, useAppPalette } from "@/constants/theme";

interface TextFieldProps {
  label: string;
  value: string;
  onChangeText: (value: string) => void;
  placeholder?: string;
  secureTextEntry?: boolean;
  keyboardType?: "default" | "phone-pad" | "numeric";
}

export function TextField({
  label,
  value,
  onChangeText,
  placeholder,
  secureTextEntry = false,
  keyboardType = "default",
}: TextFieldProps) {
  const palette = useAppPalette();

  return (
    <View style={{ gap: theme.spacing.sm }}>
      <Text
        selectable
        style={{
          color: palette.textSecondary,
          fontWeight: "600",
          fontSize: theme.fontSize.sm,
        }}
      >
        {label}
      </Text>
      <TextInput
        value={value}
        onChangeText={onChangeText}
        placeholder={placeholder}
        placeholderTextColor={palette.textMuted}
        secureTextEntry={secureTextEntry}
        keyboardType={keyboardType}
        style={{
          borderRadius: theme.radius.md,
          borderCurve: "continuous",
          borderWidth: 1,
          borderColor: palette.border,
          backgroundColor: palette.surface,
          color: palette.text,
          paddingHorizontal: 16,
          paddingVertical: 14,
          fontSize: theme.fontSize.md,
        }}
      />
    </View>
  );
}
