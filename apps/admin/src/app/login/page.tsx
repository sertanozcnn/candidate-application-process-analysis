import { BrandLogo } from "@capa/ui";

import { authNoticeFromQuery } from "@/lib/admin-auth";
import { LoginForm } from "./login-form";

export default async function LoginPage({ searchParams }: { searchParams: Promise<{ notice?: string }> }) {
  const { notice } = await searchParams;

  return (
    <main className="flex min-h-svh items-center justify-center bg-capa-bg px-5 py-10 text-capa-ink">
      <section className="w-full max-w-md rounded-lg border border-capa-line bg-capa-panel p-6 shadow-sm sm:p-8">
        <BrandLogo />
        <p className="mt-8 text-xs font-semibold uppercase tracking-[0.16em] text-capa-muted">Yönetici girişi</p>
        <h1 className="mt-3 text-2xl font-semibold">Çalışma alanına giriş yap</h1>
        <p className="mt-3 text-sm leading-6 text-capa-muted">Yönetici oturumunuzu doğrulamak için e-posta ve parolanızı girin.</p>
        <LoginForm initialNotice={authNoticeFromQuery(notice)} />
      </section>
    </main>
  );
}
