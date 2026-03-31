import { Redirect } from "expo-router";
import { ActivityIndicator, View } from "react-native";

import { useAppPalette } from "@/constants/theme";
import { useAuthStore } from "@/stores/auth-store";

export default function IndexRoute() {
  const palette = useAppPalette();
  const isHydrated = useAuthStore((state) => state.isHydrated);
  const session = useAuthStore((state) => state.session);

  if (!isHydrated) {
    return (
      <View
        style={{
          flex: 1,
          alignItems: "center",
          justifyContent: "center",
          backgroundColor: palette.background,
        }}
      >
        <ActivityIndicator color={palette.primary} />
      </View>
    );
  }

  return <Redirect href={session ? "/(tabs)" : "/auth/login"} />;
}
