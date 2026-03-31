import * as SecureStore from "expo-secure-store";

type SocialProvider = "wechat" | "apple" | "google";

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
