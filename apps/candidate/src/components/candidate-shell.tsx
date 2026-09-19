import { BrandGlow, BrandLogo, ButtonLink } from "@capa/ui";
import Link from "next/link";
import type { ReactNode } from "react";

export function CandidateShell({ children }: { children: ReactNode }) {
  return (
    <main className="relative isolate min-h-svh w-full min-w-0 overflow-x-hidden bg-capa-bg text-capa-ink selection:bg-capa-rose/40">
      <BrandGlow />
      <div className="mx-auto flex min-h-svh w-full min-w-0 max-w-[1280px] flex-col border-x border-dashed border-capa-line">
        <header className="flex min-h-24 items-center justify-between gap-4 border-b border-dashed border-capa-line px-5 sm:px-10">
          <Link href="/" aria-label="CAPA ana sayfa" className="rounded-lg focus-visible:outline-capa-violet"><BrandLogo /></Link>
          <nav aria-label="Ana menü" className="hidden items-center gap-8 text-sm font-semibold md:flex">
            <Link href="/#surec" className="transition-colors hover:text-capa-brand">Nasıl çalışır?</Link>
            <Link href="/#hakkinda" className="transition-colors hover:text-capa-brand">Proje hakkında</Link>
          </nav>
          <ButtonLink href="/apply" arrow="diagonal" className="h-10 px-4 text-xs sm:text-sm">Başvur</ButtonLink>
        </header>
        <div className="min-w-0 flex-1">{children}</div>
        <footer id="hakkinda" className="border-t border-dashed border-capa-line px-5 py-8 text-center text-xs leading-6 text-capa-muted sm:px-10">
          CAPA · Açık kaynak aday başvuru projesi.
        </footer>
      </div>
    </main>
  );
}
