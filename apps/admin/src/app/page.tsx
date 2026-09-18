import { BrandLogo } from "@capa/ui";
import { headers } from "next/headers";
import { redirect } from "next/navigation";

import { getApiHealth } from "@/lib/api-health";
import { ADMIN_AUTH_API_PATHS, type Admin, type AdminResponse } from "@/lib/admin-auth";
import { LogoutButton } from "./logout-button";

type CurrentAdminResult = { admin?: Admin; unauthorized: boolean; unavailable: boolean };

async function getCurrentAdmin(cookie: string): Promise<CurrentAdminResult> {
  try {
    const apiURL = process.env.CAPA_API_URL ?? "http://localhost:8080";
    const response = await fetch(new URL(ADMIN_AUTH_API_PATHS.me, apiURL), { headers: { cookie }, cache: "no-store" });
    if (response.status === 401) return { unauthorized: true, unavailable: false };
    if (!response.ok) return { unauthorized: false, unavailable: true };
    const data = (await response.json()) as AdminResponse;
    return { admin: data.admin, unauthorized: false, unavailable: false };
  } catch {
    return { unauthorized: false, unavailable: true };
  }
}

export default async function Home() {
  const cookie = (await headers()).get("cookie");
  if (!cookie) redirect("/login");

  const result = await getCurrentAdmin(cookie);
  if (result.unauthorized) redirect("/login?notice=session-expired");
  if (result.unavailable) redirect("/login?notice=network-error");
  if (!result.admin) redirect("/login");

  const apiHealth = await getApiHealth();

  return (
    <main className="min-h-svh bg-capa-bg px-5 py-8 text-capa-ink sm:px-10">
      <section className="mx-auto flex w-full max-w-[960px] flex-col gap-6 rounded-lg border border-capa-line bg-capa-panel p-6 shadow-sm sm:p-8">
        <div className="flex flex-col gap-3 border-b border-capa-line pb-5 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-4"><BrandLogo /><div><p className="text-xs font-semibold uppercase tracking-[0.16em] text-capa-muted">CAPA / Yönetici</p><h1 className="mt-2 text-2xl font-semibold">Yönetici çalışma alanı</h1></div></div>
          <p className="text-sm text-capa-muted">{result.admin.email}</p>
        </div>
        <p className="text-sm leading-6 text-capa-muted">Oturum doğrulandı. Başvuru listesi ve raporlar Faz 3 ve Faz 5 kapsamında eklenecek.</p>
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"><LogoutButton /><span className="text-xs text-capa-muted">{apiHealth.label}</span></div>
      </section>
    </main>
  );
}
