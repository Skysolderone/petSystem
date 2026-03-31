import { useMutation } from "@tanstack/react-query";

import { mapAuthResponse, publicRequest } from "@/services/api";
import { useAuthStore } from "@/stores/auth-store";
import type { AuthResponse } from "@/types/auth";

interface LoginPayload {
  phone: string;
  password: string;
}

interface RegisterPayload extends LoginPayload {
  nickname: string;
}

interface WechatLoginPayload {
  open_id: string;
  nickname?: string;
  avatar_url?: string;
}

interface AppleLoginPayload {
  apple_id?: string;
  identity_token?: string;
  email?: string;
  nickname?: string;
  avatar_url?: string;
}

interface GoogleLoginPayload {
  google_id?: string;
  identity_token?: string;
  email?: string;
  nickname?: string;
  avatar_url?: string;
}

export function useLogin() {
  const setSession = useAuthStore((state) => state.setSession);

  return useMutation({
    mutationFn: async (payload: LoginPayload) => {
      const response = await publicRequest<AuthResponse>("/auth/login", {
        method: "POST",
        body: JSON.stringify(payload),
      });
      return response.data;
    },
    onSuccess: async (payload) => {
      await setSession(mapAuthResponse(payload));
    },
  });
}

export function useRegister() {
  const setSession = useAuthStore((state) => state.setSession);

  return useMutation({
    mutationFn: async (payload: RegisterPayload) => {
      const response = await publicRequest<AuthResponse>("/auth/register", {
        method: "POST",
        body: JSON.stringify(payload),
      });
      return response.data;
    },
    onSuccess: async (payload) => {
      await setSession(mapAuthResponse(payload));
    },
  });
}

export function useWechatLogin() {
  const setSession = useAuthStore((state) => state.setSession);

  return useMutation({
    mutationFn: async (payload: WechatLoginPayload) => {
      const response = await publicRequest<AuthResponse>("/auth/login/wechat", {
        method: "POST",
        body: JSON.stringify(payload),
      });
      return response.data;
    },
    onSuccess: async (payload) => {
      await setSession(mapAuthResponse(payload));
    },
  });
}

export function useAppleLogin() {
  const setSession = useAuthStore((state) => state.setSession);

  return useMutation({
    mutationFn: async (payload: AppleLoginPayload) => {
      const response = await publicRequest<AuthResponse>("/auth/login/apple", {
        method: "POST",
        body: JSON.stringify(payload),
      });
      return response.data;
    },
    onSuccess: async (payload) => {
      await setSession(mapAuthResponse(payload));
    },
  });
}

export function useGoogleLogin() {
  const setSession = useAuthStore((state) => state.setSession);

  return useMutation({
    mutationFn: async (payload: GoogleLoginPayload) => {
      const response = await publicRequest<AuthResponse>("/auth/login/google", {
        method: "POST",
        body: JSON.stringify(payload),
      });
      return response.data;
    },
    onSuccess: async (payload) => {
      await setSession(mapAuthResponse(payload));
    },
  });
}
