import type { SystemStatusState } from "@capa/ui";

type ApiHealth = {
  label: string;
  detail: string;
  state: SystemStatusState;
};

export async function getApiHealth(): Promise<ApiHealth> {
  const apiURL = process.env.CAPA_API_URL ?? "http://localhost:8080";
  const readyURL = new URL("/health/ready", apiURL);

  try {
    const response = await fetch(readyURL, {
      cache: "no-store",
      signal: AbortSignal.timeout(1500),
    });

    if (!response.ok) {
      return {
        label: "API erişilebilir",
        detail: "Veritabanı hazır değil",
        state: "offline",
      };
    }

    return {
      label: "API ve veritabanı hazır",
      detail: apiURL,
      state: "ready",
    };
  } catch {
    return {
      label: "API bağlantısı yok",
      detail: apiURL,
      state: "offline",
    };
  }
}
