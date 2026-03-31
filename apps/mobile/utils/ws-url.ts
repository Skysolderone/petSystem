import { getApiBaseUrl } from "@/utils/api-base-url";

export function getWsUrl(path: string): string {
  const baseUrl = getApiBaseUrl();
  const origin = baseUrl.replace(/\/api\/v1$/, "").replace(/^http/, "ws");
  return `${origin}${path}`;
}
