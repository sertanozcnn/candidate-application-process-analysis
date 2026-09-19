import { BrandLogo, ButtonLink } from "@capa/ui";
import { headers } from "next/headers";
import { redirect } from "next/navigation";

import { LogoutButton } from "@/app/logout-button";
import { formatSubmittedAt, getApplication, positionLabels } from "@/lib/applications";
import { ADMIN_AUTH_API_PATHS, type Admin, type AdminResponse } from "@/lib/admin-auth";

type PageProps = { params: Promise<{ id: string }> };

async function getCurrentAdmin(cookie: string): Promise<{ admin?: Admin; unauthorized: boolean; unavailable: boolean }> {
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

export default async function ApplicationDetailPage({ params }: PageProps) {
  const cookie = (await headers()).get("cookie");
  if (!cookie) redirect("/login");
  const session = await getCurrentAdmin(cookie);
  if (session.unauthorized) redirect("/login?notice=session-expired");
  if (session.unavailable || !session.admin) redirect("/login?notice=network-error");

  const { id } = await params;
  const result = await getApplication(cookie, id);
  const application = result.status === 200 ? result.data : null;

  return (
    <main className="min-h-svh overflow-x-hidden bg-capa-bg px-5 py-8 text-capa-ink sm:px-10">
      <section className="mx-auto flex w-full max-w-[960px] flex-col gap-6 rounded-lg border border-capa-line bg-capa-panel p-6 shadow-sm sm:p-8">
        <div className="flex min-w-0 flex-col gap-4 border-b border-capa-line pb-5 sm:flex-row sm:items-center sm:justify-between"><div className="flex min-w-0 items-center gap-4"><BrandLogo /><div className="min-w-0"><p className="text-xs font-semibold uppercase tracking-[0.16em] text-capa-muted">CAPA / Başvuru detayı</p><h1 className="mt-2 break-words text-2xl font-semibold">Başvuru detayı</h1></div></div><p className="max-w-full break-words text-sm text-capa-muted [overflow-wrap:anywhere]">{session.admin.email}</p></div>
        {result.status === 404 ? <div className="rounded-lg border border-dashed border-capa-line bg-capa-bg/60 p-8 text-center"><h2 className="font-semibold">Başvuru bulunamadı</h2><p className="mt-2 text-sm text-capa-muted">Bu başvuru silinmiş veya geçersiz bir id kullanılmış olabilir.</p></div> : result.status === 401 ? <div className="rounded-lg border border-red-200 bg-red-50 p-5 text-sm text-red-900" role="alert">Oturumunuz geçersiz. Lütfen tekrar giriş yapın.</div> : result.status === 500 || result.status === "network" ? <div className="rounded-lg border border-red-200 bg-red-50 p-5 text-sm text-red-900" role="alert">Başvuru detayı yüklenemedi. Lütfen tekrar deneyin.</div> : application ? <article className="grid min-w-0 gap-6"><div className="flex min-w-0 flex-col gap-3 border-b border-capa-line pb-5 sm:flex-row sm:items-start sm:justify-between"><div className="min-w-0"><h2 className="break-words text-2xl font-semibold [overflow-wrap:anywhere]">{application.full_name}</h2><p className="mt-2 break-words text-sm text-capa-muted [overflow-wrap:anywhere]">{application.email}</p></div><span className="w-fit max-w-full shrink-0 break-words rounded-full bg-capa-violet/10 px-3 py-1.5 text-xs font-semibold text-capa-violet [overflow-wrap:anywhere]">{positionLabels[application.position_code] ?? application.position_code}</span></div><dl className="grid min-w-0 gap-4 sm:grid-cols-2"><div className="min-w-0 rounded-lg bg-capa-bg/70 p-4"><dt className="text-xs font-semibold uppercase tracking-[0.12em] text-capa-muted">Gönderilme zamanı</dt><dd className="mt-2 text-sm">{formatSubmittedAt(application.submitted_at)}</dd></div><div className="min-w-0 rounded-lg bg-capa-bg/70 p-4"><dt className="text-xs font-semibold uppercase tracking-[0.12em] text-capa-muted">Başvuru id</dt><dd className="mt-2 break-all font-mono text-xs">{application.id}</dd></div></dl><div className="min-w-0"><h3 className="text-sm font-semibold">Deneyim</h3><p className="mt-3 max-w-full whitespace-pre-wrap break-words rounded-lg border border-capa-line bg-capa-bg/50 p-5 text-sm leading-7 text-capa-muted [overflow-wrap:anywhere]">{application.experience}</p></div></article> : null}
        <div className="flex flex-col gap-3 border-t border-capa-line pt-6 sm:flex-row sm:items-center sm:justify-between"><ButtonLink href="/" variant="secondary" arrow="right" className="w-fit">Başvuru listesine dön</ButtonLink><LogoutButton /></div>
      </section>
    </main>
  );
}
