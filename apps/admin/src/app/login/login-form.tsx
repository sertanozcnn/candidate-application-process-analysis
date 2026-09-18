"use client";

import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";

import { ADMIN_AUTH_NOTICES, ADMIN_AUTH_ROUTES, getAdminCsrfToken, type AuthNotice } from "@/lib/admin-auth";
import { Snackbar } from "../snackbar";

type LoginFormProps = {
  initialNotice?: AuthNotice;
};

export function LoginForm({ initialNotice }: LoginFormProps) {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [notice, setNotice] = useState<AuthNotice | null>(initialNotice ?? null);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setNotice(null);

    const form = new FormData(event.currentTarget);
    const email = String(form.get("email") ?? "");
    const password = String(form.get("password") ?? "");

    try {
      const csrfToken = await getAdminCsrfToken();
      const response = await fetch(ADMIN_AUTH_ROUTES.login, {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-CSRF-Token": csrfToken },
        body: JSON.stringify({ email, password }),
      });

      if (response.status === 401) {
        setNotice(ADMIN_AUTH_NOTICES.invalidCredentials);
        return;
      }
      if (!response.ok) {
        setNotice(ADMIN_AUTH_NOTICES.loginFailed);
        return;
      }

      router.replace("/");
      router.refresh();
    } catch {
      setNotice(ADMIN_AUTH_NOTICES.networkError);
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <form onSubmit={submit} className="mt-8 grid gap-5">
        <div className="grid gap-2">
          <label htmlFor="email" className="text-sm font-semibold">E-posta</label>
          <input id="email" name="email" type="email" autoComplete="email" required className="h-12 rounded-lg border border-capa-line bg-capa-panel px-4 text-sm outline-none focus:border-capa-violet focus:ring-4 focus:ring-capa-violet/20" />
        </div>
        <div className="grid gap-2">
          <label htmlFor="password" className="text-sm font-semibold">Parola</label>
          <input id="password" name="password" type="password" autoComplete="current-password" required className="h-12 rounded-lg border border-capa-line bg-capa-panel px-4 text-sm outline-none focus:border-capa-violet focus:ring-4 focus:ring-capa-violet/20" />
        </div>
        <button type="submit" disabled={loading} className="h-12 rounded-full bg-capa-brand px-6 text-sm font-semibold text-white transition hover:bg-capa-violet focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-capa-violet/30 disabled:opacity-60">
          {loading ? "Giriş yapılıyor..." : "Giriş yap"}
        </button>
      </form>
      {notice && <Snackbar message={notice.message} tone={notice.tone} onClose={() => setNotice(null)} />}
    </>
  );
}
