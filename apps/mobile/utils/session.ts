import * as SecureStore from "expo-secure-store";

import type { AuthSession } from "@/types/auth";

const sessionKey = "petverse.auth.session";

export async function getStoredSession(): Promise<AuthSession | null> {
  const rawValue = await SecureStore.getItemAsync(sessionKey);
  if (!rawValue) {
    return null;
  }

  try {
    return JSON.parse(rawValue) as AuthSession;
  } catch {
    await SecureStore.deleteItemAsync(sessionKey);
    return null;
  }
}

export async function setStoredSession(session: AuthSession): Promise<void> {
  await SecureStore.setItemAsync(sessionKey, JSON.stringify(session));
}

export async function clearStoredSession(): Promise<void> {
  await SecureStore.deleteItemAsync(sessionKey);
}
