import { Button, ButtonLink, BrandLogo, BrandGlow, SystemStatus } from "@capa/ui";
import Link from "next/link";

import { getApiHealth } from "@/lib/api-health";

export const dynamic = "force-dynamic";

export default async function Home() {
  const apiHealth = await getApiHealth();

  return (
    <main className="relative isolate min-h-svh overflow-x-clip bg-capa-bg text-capa-ink selection:bg-capa-rose/40">
      <BrandGlow />
      <div className="mx-auto min-h-svh w-full max-w-[1280px] border-x border-dashed border-capa-line">
        <header className="flex min-h-24 items-center justify-between gap-4 border-b border-dashed border-capa-line px-5 sm:px-10">
          <Link href="/" aria-label="CAPA ana sayfa" className="rounded-lg focus-visible:outline-capa-violet"><BrandLogo /></Link>
          <nav aria-label="Ana menü" className="hidden items-center gap-8 text-sm font-semibold md:flex">
            <a href="#surec" className="transition-colors hover:text-capa-brand">Nasıl çalışır?</a>
            <a href="#hakkinda" className="transition-colors hover:text-capa-brand">Proje hakkında</a>
          </nav>
          <ButtonLink href="#surec" arrow="diagonal" className="h-10 px-4 text-xs sm:text-sm">Süreci keşfet</ButtonLink>
        </header>
        <section className="flex flex-col items-center border-b border-dashed border-capa-line px-5 pb-14 pt-16 text-center sm:px-10 sm:pb-16 sm:pt-24">
          <p className="inline-flex items-center gap-2 rounded-full border border-capa-rose/40 bg-capa-panel/35 p-1 pr-3 text-[11px] font-medium sm:text-xs">
            <span className="rounded-full bg-capa-brand/10 px-2.5 py-1 text-capa-brand">CAPA</span>
            Kariyerindeki bir sonraki adım
          </p>
          <h1 className="mt-5 max-w-[900px] text-[clamp(2.25rem,5vw,4.25rem)] font-semibold leading-[1.13] tracking-[-0.055em]">
            Yeni bir başlangıç,<br />
            <span className="rounded-xl bg-capa-brand/15 px-2 text-capa-ink">kendini anlatmakla</span><br />
            başlar.
          </h1>
          <p className="mt-6 max-w-[600px] text-sm leading-7 text-capa-muted sm:text-base">
            Deneyimlerini paylaş, başvurunu tek bir yerden tamamla.
            Bir sonraki fırsatına sade ve anlaşılır bir süreçle ulaş.
          </p>
          <div className="mt-7 flex w-full flex-col justify-center gap-3 sm:w-auto sm:flex-row">
            <Button aria-disabled="true" arrow="diagonal" title="Başvuru akışı yakında açılacak">Başvurular yakında</Button>
            <ButtonLink href="#surec" variant="secondary">Süreci keşfet</ButtonLink>
          </div>
          <div className="mt-5 flex justify-center">
            <SystemStatus {...apiHealth} />
          </div>
        </section>
        <section id="surec" aria-label="Başvuru süreci" className="scroll-mt-6 px-5 py-10 sm:px-10">
          <ol className="mx-auto grid max-w-[960px] gap-6 sm:grid-cols-3">
            {[
              ["01", "Hesabını oluştur", "Kendine ait başvuru alanına giriş yap."],
              ["02", "Kendini anlat", "Deneyimlerini ve yetkinliklerini paylaş."],
              ["03", "Başvurunu tamamla", "Bilgilerini gözden geçir ve gönder."],
            ].map(([number, title, description]) => (
              <li key={number} className="flex items-start gap-3">
                <span className="flex size-9 shrink-0 items-center justify-center rounded-full border border-capa-brand/15 bg-capa-panel/60 text-xs font-semibold text-capa-brand">{number}</span>
                <div><h2 className="text-sm font-semibold">{title}</h2><p className="mt-1 text-xs leading-5 text-capa-muted">{description}</p></div>
              </li>
            ))}
          </ol>
          <p id="hakkinda" className="mt-10 text-center text-xs leading-6 text-capa-muted">CAPA · Açık kaynak aday başvuru projesi. Başvuru akışı geliştirme aşamasında.</p>
        </section>
      </div>
    </main>
  );
}
