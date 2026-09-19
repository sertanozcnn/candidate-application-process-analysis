export type Application = {
  id: string;
  position_code: string;
  full_name: string;
  email: string;
  experience: string;
  submitted_at: string;
};

type ApplicationListResponse = {
  applications: Application[];
  limit: number;
  offset: number;
};

type ApplicationDetailResponse = {
  application: Application;
};

export type ApplicationFetchResult<T> =
  | { data: T; status: 200 }
  | { data: null; status: 401 | 404 | 500 | "network" };

const apiURL = process.env.CAPA_API_URL ?? "http://localhost:8080";

export async function getApplications(cookie: string, limit = 20, offset = 0): Promise<ApplicationFetchResult<ApplicationListResponse>> {
  try {
    const url = new URL("/v1/admin/applications", apiURL);
    url.searchParams.set("limit", String(limit));
    url.searchParams.set("offset", String(offset));
    const response = await fetch(url, { headers: { cookie }, cache: "no-store" });
    if (response.status === 401) return { data: null, status: 401 };
    if (!response.ok) return { data: null, status: 500 };
    return { data: (await response.json()) as ApplicationListResponse, status: 200 };
  } catch {
    return { data: null, status: "network" };
  }
}

export async function getApplication(cookie: string, id: string): Promise<ApplicationFetchResult<Application>> {
  try {
    const url = new URL(`/v1/admin/applications/${encodeURIComponent(id)}`, apiURL);
    const response = await fetch(url, { headers: { cookie }, cache: "no-store" });
    if (response.status === 401) return { data: null, status: 401 };
    if (response.status === 404) return { data: null, status: 404 };
    if (!response.ok) return { data: null, status: 500 };
    const body = (await response.json()) as ApplicationDetailResponse;
    return { data: body.application, status: 200 };
  } catch {
    return { data: null, status: "network" };
  }
}

export const positionLabels: Record<string, string> = {
  frontend: "Frontend geliştirici",
  backend: "Backend geliştirici",
  fullstack: "Fullstack geliştirici",
};

export function formatSubmittedAt(value: string) {
  return new Intl.DateTimeFormat("tr-TR", { dateStyle: "medium", timeStyle: "short" }).format(new Date(value));
}
