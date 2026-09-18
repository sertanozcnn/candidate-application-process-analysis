export const ADMIN_AUTH_ROUTES = {
  csrf: "/api/admin/csrf",
  login: "/api/admin/auth/login",
  logout: "/api/admin/auth/logout",
  me: "/api/admin/auth/me",
} as const;

export const ADMIN_AUTH_API_PATHS = {
  login: "/v1/admin/auth/login",
  logout: "/v1/admin/auth/logout",
  me: "/v1/admin/auth/me",
} as const;

export const ADMIN_AUTH_UPSTREAM_PATHS = {
  login: "auth/login",
  logout: "auth/logout",
  me: "auth/me",
} as const;

export const ADMIN_AUTH_UPSTREAM_ALLOWLIST: ReadonlySet<string> = new Set(Object.values(ADMIN_AUTH_UPSTREAM_PATHS));

export type Admin = {
  id: string;
  email: string;
};

export type AdminResponse = {
  admin: Admin;
};

export type AuthNotice = {
  message: string;
  tone: "error" | "success" | "info";
};

export const ADMIN_AUTH_NOTICES = {
  invalidCredentials: { message: "E-posta veya parola hatalı.", tone: "error" },
  loginFailed: { message: "Giriş yapılamadı. Lütfen tekrar deneyin.", tone: "error" },
  logoutFailed: { message: "Çıkış yapılamadı. Lütfen tekrar deneyin.", tone: "error" },
  networkError: { message: "Ağ bağlantısı kurulamadı. Lütfen tekrar deneyin.", tone: "error" },
  logoutSuccess: { message: "Oturum başarıyla kapatıldı.", tone: "success" },
  sessionExpired: { message: "Oturumunuzun süresi doldu. Lütfen tekrar giriş yapın.", tone: "error" },
} satisfies Record<string, AuthNotice>;

export function authNoticeFromQuery(value?: string): AuthNotice | undefined {
  switch (value) {
    case "logout-success":
      return ADMIN_AUTH_NOTICES.logoutSuccess;
    case "session-expired":
      return ADMIN_AUTH_NOTICES.sessionExpired;
    case "network-error":
      return ADMIN_AUTH_NOTICES.networkError;
    default:
      return undefined;
  }
}
export async function getAdminCsrfToken(): Promise<string> {
  const response = await fetch(ADMIN_AUTH_ROUTES.csrf, { cache: "no-store" });
  if (!response.ok) throw new Error("csrf request failed");

  const data = (await response.json()) as { csrfToken?: unknown };
  if (typeof data.csrfToken !== "string" || data.csrfToken === "") {
    throw new Error("csrf token missing");
  }

  return data.csrfToken;
}
