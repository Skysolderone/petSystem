import { create } from "zustand";

import type { AuthSession, User } from "@/types/auth";
import { clearStoredSession, getStoredSession, setStoredSession } from "@/utils/session";

interface AuthState {
  isHydrated: boolean;
  session: AuthSession | null;
  hydrate: () => Promise<void>;
  setSession: (session: AuthSession) => Promise<void>;
  updateUser: (user: User) => Promise<void>;
  logout: () => Promise<void>;
  clearSessionState: () => void;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  isHydrated: false,
  session: null,
  hydrate: async () => {
    const session = await getStoredSession();
    set({
      isHydrated: true,
      session,
    });
  },
  setSession: async (session) => {
    await setStoredSession(session);
    set({ session });
  },
  updateUser: async (user) => {
    const currentSession = get().session;
    if (!currentSession) {
      return;
    }

    const nextSession = {
      ...currentSession,
      user,
    };
    await setStoredSession(nextSession);
    set({ session: nextSession });
  },
  logout: async () => {
    await clearStoredSession();
    set({
      session: null,
      isHydrated: true,
    });
  },
  clearSessionState: () => {
    set({
      session: null,
      isHydrated: true,
    });
  },
}));
