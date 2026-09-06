import { BrandGlow, BrandLogo, ButtonLink, SystemStatus } from "@capa/ui";
import Link from "next/link";

import { getApiHealth } from "@/lib/api-health";

export const dynamic = "force-dynamic";

export default async function Home() {
  const apiHealth = await getApiHealth();

  return (
    <main className="relative isolate min-h-svh overflow-x-clip bg-capa-bg text-capa-ink selection:bg-capa-rose/20">
      <BrandGlow />
      <div className="mx-auto min-h-svh w-full max-w-[1280px] border-x border-dashed border-capa-line">
        <header className="flex min-h-24 items-center justify-between gap-4 border-b border-dashed border-capa-line px-5 sm:px-10">
          <Link href="/" aria-label="CAPA yönetici ana sayfa" className="rounded-lg focus-visible:outline-capa-violet"><BrandLogo /></Link>
          <span className="hidden text-xs font-semibold uppercase tracking-[0.18em] text-capa-muted sm:block">Yönetici çalışma alanı</span>
          <ButtonLink href="#moduller" arrow="diagonal" className="h-10 px-4 text-xs sm:text-sm">Paneli keşfet</ButtonLink>
        </header>
        <div className="grid items-center gap-12 px-5 py-16 sm:px-10 lg:grid-cols-[1fr_1.05fr] lg:gap-16 lg:py-24">
          <section className="min-w-0">
            <p className="inline-flex items-center gap-2 rounded-full border border-capa-rose/20 bg-capa-panel/50 px-3 py-1.5 text-xs font-medium text-capa-brand">
              <span className="size-1.5 rounded-full bg-capa-brand" /> Sürecin bütününü gör
            </p>
            <h1 className="mt-6 text-[clamp(2.25rem,4.3vw,3.75rem)] font-semibold leading-[1.13] tracking-[-0.05em]">
              Her başvuru,<br /><span className="bg-linear-to-r from-capa-brand to-capa-violet bg-clip-text text-transparent">bir yeni hikâye.</span>
            </h1>
            <p className="mt-6 max-w-md text-sm leading-7 text-capa-muted sm:text-base">
              Başvuruları bir arada incele. Adayların deneyimlerini ve başvuru sürecindeki etkileşimlerini anlaşılır bir görünümde keşfet.
            </p>
            <div className="mt-8 flex flex-col gap-3 sm:flex-row">
              <ButtonLink href="#moduller">Modülleri incele</ButtonLink>
              <ButtonLink href="#akis" variant="secondary">İşleyişi gör</ButtonLink>
            </div>
            <div className="mt-5">
              <SystemStatus {...apiHealth} />
            </div>
            <p className="mt-5 text-xs leading-5 text-capa-muted">Panel önizlemesi · Yönetici girişi yakında.</p>
          </section>
          <section id="moduller" aria-label="Admin panel önizlemesi" className="relative min-w-0 scroll-mt-6 rounded-lg border border-capa-line bg-capa-panel/85 p-5 shadow-xl shadow-capa-violet/10 sm:p-7">
            <div aria-hidden="true" className="absolute inset-x-6 top-0 h-0.5 bg-linear-to-r from-capa-coral via-capa-rose to-capa-violet" />
            <div className="flex items-center justify-between gap-3 border-b border-capa-line pb-5">
              <div><p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-capa-muted">CAPA / Çalışma alanı</p><h2 className="mt-1 text-lg font-semibold">Genel bakış</h2></div>
              <span className="rounded-full bg-capa-surface px-3 py-1 text-[10px] font-medium text-capa-brand">Önizleme</span>
            </div>
            <div className="mt-5 grid gap-3">
              {[
                ["01", "Başvurular", "Tüm başvuruları tek bir listede incele."],
                ["02", "Aday profili", "Deneyim ve başvuru bilgilerini keşfet."],
                ["03", "Etkileşimler", "Başvuru sürecini zaman çizelgesinde gör."],
              ].map(([number, label, detail]) => (
                <div key={number} className="flex items-start gap-4 rounded-lg border border-capa-line/70 bg-linear-to-br from-capa-panel to-capa-surface/60 p-4">
                  <span className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-linear-to-br from-capa-rose/15 to-capa-violet/15 text-xs font-semibold text-capa-brand">{number}</span>
                  <div><h3 className="text-sm font-semibold">{label}</h3><p className="mt-1 text-xs leading-5 text-capa-muted">{detail}</p></div>
                </div>
              ))}
            </div>
            <p className="mt-5 flex items-center gap-2 text-xs text-capa-muted"><span className="size-1.5 shrink-0 rounded-full bg-capa-violet" />Henüz gerçek başvuru verisi bağlı değil.</p>
          </section>
        </div>
        <section id="akis" className="scroll-mt-6 border-t border-dashed border-capa-line px-5 py-8 sm:px-10">
          <p className="text-xs font-semibold uppercase tracking-[0.16em] text-capa-muted">Üç adımda inceleme</p>
          <ol className="mt-5 grid gap-5 text-sm sm:grid-cols-3">
            {["Başvuruyu seç", "Adayı tanı", "Süreci incele"].map((step, index) => (
              <li key={step} className="flex items-center gap-3"><span className="text-xs font-semibold text-capa-brand">0{index + 1}</span>{step}</li>
            ))}
          </ol>
        </section>
      </div>
    </main>
  );
}
