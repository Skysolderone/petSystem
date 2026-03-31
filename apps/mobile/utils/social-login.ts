import * as SecureStore from "expo-secure-store";
import * as WebBrowser from "expo-web-browser";

WebBrowser.maybeCompleteAuthSession();

type SocialProvider = "wechat" | "apple" | "google";

export interface GoogleNativeConfig {
  webClientId?: string;
  iosClientId?: string;
  androidClientId?: string;
}

function buildSeed(provider: SocialProvider): string {
  return `${provider}-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;
}

export async function getSocialLoginSeed(provider: SocialProvider): Promise<string> {
  const key = `petverse.social.${provider}`;
  const existing = await SecureStore.getItemAsync(key);
  if (existing) {
    return existing;
  }

  const seed = buildSeed(provider);
  await SecureStore.setItemAsync(key, seed);
  return seed;
}

export function getGoogleNativeConfig(): GoogleNativeConfig {
  return {
    webClientId: trimEnv(process.env.EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID),
    iosClientId: trimEnv(process.env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID),
    androidClientId: trimEnv(process.env.EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID),
  };
}

export function hasGoogleNativeConfig(): boolean {
  const config = getGoogleNativeConfig();
  return Boolean(config.webClientId || config.iosClientId || config.androidClientId);
}

export function formatAppleDisplayName(
  fullName:
    | {
        givenName?: string | null;
        middleName?: string | null;
        familyName?: string | null;
        nickname?: string | null;
      }
    | null
    | undefined,
): string | undefined {
  if (!fullName) {
    return undefined;
  }

  const parts = [
    fullName.familyName?.trim(),
    fullName.middleName?.trim(),
    fullName.givenName?.trim(),
  ].filter(Boolean);

  if (parts.length > 0) {
    return parts.join("");
  }

  return fullName.nickname?.trim() || undefined;
}

function trimEnv(value: string | undefined): string | undefined {
  const resolved = value?.trim();
  return resolved ? resolved : undefined;
}
