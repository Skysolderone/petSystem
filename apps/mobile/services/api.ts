import { useAuthStore } from "@/stores/auth-store";
import type { ApiEnvelope } from "@/types/api";
import { ApiError } from "@/types/api";
import type { AuthResponse, AuthSession } from "@/types/auth";
import { getApiBaseUrl } from "@/utils/api-base-url";
import { clearStoredSession, getStoredSession, setStoredSession } from "@/utils/session";

type RequestOptions = RequestInit & {
  auth?: boolean;
  retryOnAuthError?: boolean;
};

let refreshPromise: Promise<AuthSession | null> | null = null;

function mapAuthResponse(payload: AuthResponse): AuthSession {
  return {
    user: payload.user,
    accessToken: payload.access_token,
    refreshToken: payload.refresh_token,
    expiresIn: payload.expires_in,
  };
}

async function parseResponse<T>(response: Response): Promise<ApiEnvelope<T>> {
  const rawText = await response.text();
  const payload = rawText ? (JSON.parse(rawText) as ApiEnvelope<T>) : null;

  if (!response.ok) {
    throw new ApiError(
      payload?.message ?? "Request failed",
      response.status,
      (payload?.data as { error_code?: string } | undefined)?.error_code
    );
  }

  if (!payload) {
    throw new ApiError("Empty response body", response.status);
  }

  return payload;
}

async function refreshSession(): Promise<AuthSession | null> {
  if (refreshPromise) {
    return refreshPromise;
  }

  refreshPromise = (async () => {
    const session = await getStoredSession();
    if (!session?.refreshToken) {
      return null;
    }

    const response = await fetch(`${getApiBaseUrl()}/auth/refresh`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        refresh_token: session.refreshToken,
      }),
    });

    if (!response.ok) {
      await clearStoredSession();
      useAuthStore.getState().clearSessionState();
      return null;
    }

    const payload = await parseResponse<AuthResponse>(response);
    const nextSession = mapAuthResponse(payload.data);
    await setStoredSession(nextSession);
    await useAuthStore.getState().setSession(nextSession);
    return nextSession;
  })();

  try {
    return await refreshPromise;
  } finally {
    refreshPromise = null;
  }
}

export async function apiRequest<T>(path: string, options: RequestOptions = {}): Promise<ApiEnvelope<T>> {
  const { auth = true, retryOnAuthError = true, headers, ...restOptions } = options;
  const session = await getStoredSession();
  const isFormData = typeof FormData !== "undefined" && restOptions.body instanceof FormData;

  const requestHeaders = new Headers(headers);
  if (!requestHeaders.has("Content-Type") && restOptions.body && !isFormData) {
    requestHeaders.set("Content-Type", "application/json");
  }
  if (auth && session?.accessToken) {
    requestHeaders.set("Authorization", `Bearer ${session.accessToken}`);
  }

  const response = await fetch(`${getApiBaseUrl()}${path}`, {
    ...restOptions,
    headers: requestHeaders,
  });

  if (response.status === 401 && auth && retryOnAuthError) {
    const refreshedSession = await refreshSession();
    if (refreshedSession) {
      return apiRequest<T>(path, {
        ...options,
        retryOnAuthError: false,
      });
    }
  }

  return parseResponse<T>(response);
}

export async function publicRequest<T>(path: string, options: RequestOptions = {}): Promise<ApiEnvelope<T>> {
  return apiRequest<T>(path, {
    ...options,
    auth: false,
    retryOnAuthError: false,
  });
}

export { mapAuthResponse };
