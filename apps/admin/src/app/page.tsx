import { BrandLogo, ButtonLink } from "@capa/ui";
import { headers } from "next/headers";
import { redirect } from "next/navigation";

import { getApiHealth } from "@/lib/api-health";
import { ADMIN_AUTH_API_PATHS, type Admin, type AdminResponse } from "@/lib/admin-auth";
import { formatSubmittedAt, getApplications, positionLabels } from "@/lib/applications";
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

type PageProps = { searchParams: Promise<{ offset?: string | string[] }> };

export default async function Home({ searchParams }: PageProps) {
  const cookie = (await headers()).get("cookie");
  if (!cookie) redirect("/login");

  const result = await getCurrentAdmin(cookie);
  if (result.unauthorized) redirect("/login?notice=session-expired");
  if (result.unavailable) redirect("/login?notice=network-error");
  if (!result.admin) redirect("/login");

  const apiHealth = await getApiHealth();
  const query = await searchParams;
  const rawOffset = Array.isArray(query.offset) ? query.offset[0] : query.offset;
  const parsedOffset = rawOffset ? Number.parseInt(rawOffset, 10) : 0;
  const offset = Number.isInteger(parsedOffset) && parsedOffset >= 0 ? parsedOffset : 0;
  const applications = await getApplications(cookie, 20, offset);
  const applicationData = applications.status === 200 ? applications.data : null;

  return (
    <main className="min-h-svh overflow-x-hidden bg-capa-bg px-5 py-8 text-capa-ink sm:px-10">
      <section className="mx-auto flex w-full max-w-[1120px] flex-col gap-6 rounded-lg border border-capa-line bg-capa-panel p-6 shadow-sm sm:p-8">
        <div className="flex flex-col gap-3 border-b border-capa-line pb-5 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-4"><BrandLogo /><div><p className="text-xs font-semibold uppercase tracking-[0.16em] text-capa-muted">CAPA / Yönetici</p><h1 className="mt-2 text-2xl font-semibold">Yönetici çalışma alanı</h1></div></div>
          <p className="text-sm text-capa-muted">{result.admin.email}</p>
        </div>
        <div className="flex flex-col gap-3 border-b border-capa-line pb-6 sm:flex-row sm:items-center sm:justify-between"><div><h2 className="text-xl font-semibold">Başvurular</h2><p className="mt-1 text-sm text-capa-muted">Gönderilmiş aday başvuruları.</p></div><span className="text-xs text-capa-muted">{apiHealth.label}</span></div>
        {applications.status === "network" || applications.status === 500 ? (
          <div className="rounded-lg border border-red-200 bg-red-50 p-5 text-sm text-red-900" role="alert">Başvurular yüklenemedi. API bağlantısını kontrol edip tekrar deneyin.</div>
        ) : applications.status === 401 ? (
          <div className="rounded-lg border border-red-200 bg-red-50 p-5 text-sm text-red-900" role="alert">Oturumunuz geçersiz. Lütfen tekrar giriş yapın.</div>
        ) : applicationData === null ? (
          <div className="rounded-lg border border-red-200 bg-red-50 p-5 text-sm text-red-900" role="alert">Başvurular yüklenemedi. Lütfen tekrar deneyin.</div>
        ) : applicationData.applications.length === 0 ? (
          <div className="rounded-lg border border-dashed border-capa-line bg-capa-bg/60 p-8 text-center"><h3 className="font-semibold">Henüz başvuru yok</h3><p className="mt-2 text-sm text-capa-muted">Yeni gönderilmiş başvurular burada görünecek.</p></div>
        ) : (
          <>
            <div className="hidden overflow-hidden rounded-lg border border-capa-line md:block">
              <table className="w-full text-left text-sm"><thead className="bg-capa-bg/70 text-xs uppercase tracking-[0.12em] text-capa-muted"><tr><th className="px-4 py-3 font-semibold">Aday</th><th className="px-4 py-3 font-semibold">Pozisyon</th><th className="px-4 py-3 font-semibold">Gönderildi</th><th className="px-4 py-3" /></tr></thead><tbody className="divide-y divide-capa-line">{applicationData.applications.map((application) => <tr key={application.id} className="transition-colors hover:bg-capa-bg/50"><td className="px-4 py-4"><p className="font-semibold">{application.full_name}</p><p className="mt-1 text-xs text-capa-muted">{application.email}</p></td><td className="px-4 py-4 text-capa-muted">{positionLabels[application.position_code] ?? application.position_code}</td><td className="px-4 py-4 text-capa-muted">{formatSubmittedAt(application.submitted_at)}</td><td className="px-4 py-4 text-right"><ButtonLink href={`/applications/${application.id}`} variant="secondary" arrow="diagonal" className="h-9 px-3 text-xs">Detay</ButtonLink></td></tr>)}</tbody></table>
            </div>
            <div className="grid gap-3 md:hidden">{applicationData.applications.map((application) => <article key={application.id} className="rounded-lg border border-capa-line p-4"><div className="flex items-start justify-between gap-4"><div><h3 className="font-semibold">{application.full_name}</h3><p className="mt-1 break-all text-xs text-capa-muted">{application.email}</p></div><span className="shrink-0 rounded-full bg-capa-violet/10 px-2.5 py-1 text-[11px] font-semibold text-capa-violet">{positionLabels[application.position_code] ?? application.position_code}</span></div><div className="mt-4 flex items-center justify-between gap-3"><time className="text-xs text-capa-muted">{formatSubmittedAt(application.submitted_at)}</time><ButtonLink href={`/applications/${application.id}`} variant="secondary" arrow="diagonal" className="h-9 px-3 text-xs">Detay</ButtonLink></div></article>)}</div>
            <div className="flex items-center justify-between text-xs text-capa-muted"><span>{applicationData.offset + 1}–{applicationData.offset + applicationData.applications.length} başvuru</span><div className="flex gap-2">{applicationData.offset > 0 && <ButtonLink href={`/?offset=${Math.max(0, applicationData.offset - applicationData.limit)}`} variant="secondary" arrow={false} className="h-9 px-3 text-xs">Önceki</ButtonLink>}{applicationData.applications.length === applicationData.limit && <ButtonLink href={`/?offset=${applicationData.offset + applicationData.limit}`} variant="secondary" arrow={false} className="h-9 px-3 text-xs">Sonraki</ButtonLink>}</div></div>
          </>
        )}
        <div className="flex flex-col gap-3 border-t border-capa-line pt-6 sm:flex-row sm:items-center sm:justify-between"><LogoutButton /><span className="text-xs text-capa-muted">Oturum doğrulandı</span></div>
      </section>
    </main>
  );
}
