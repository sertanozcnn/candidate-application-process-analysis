"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";

import { ADMIN_AUTH_NOTICES, ADMIN_AUTH_ROUTES, getAdminCsrfToken, type AuthNotice } from "@/lib/admin-auth";
import { Snackbar } from "./snackbar";

export function LogoutButton() {
  const router = useRouter();
  const triggerRef = useRef<HTMLButtonElement>(null);
  const cancelRef = useRef<HTMLButtonElement>(null);
  const confirmRef = useRef<HTMLButtonElement>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [loading, setLoading] = useState(false);
  const [notice, setNotice] = useState<AuthNotice | null>(null);

  const closeModal = useCallback(() => {
    setModalVisible(false);
    if (loading) return;
    setTimeout(() => {
      setModalOpen(false);
      requestAnimationFrame(() => triggerRef.current?.focus());
    }, 250);
  }, [loading]);

  useEffect(() => {
    if (!modalOpen) return;

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    const focusTimer = setTimeout(() => cancelRef.current?.focus(), 30);
    const visibleTimer = setTimeout(() => setModalVisible(true), 10);

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        event.preventDefault();
        closeModal();
        return;
      }
      if (event.key !== "Tab") return;

      const focusable = [cancelRef.current, confirmRef.current].filter(Boolean) as HTMLButtonElement[];
      if (focusable.length === 0) return;
      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("keydown", handleKeyDown);
      document.body.style.overflow = previousOverflow;
      clearTimeout(focusTimer);
      clearTimeout(visibleTimer);
    };
  }, [closeModal, modalOpen]);

  async function logout() {
    setLoading(true);
    setNotice(null);
    try {
      const csrfToken = await getAdminCsrfToken();
      const response = await fetch(ADMIN_AUTH_ROUTES.logout, {
        method: "POST",
        headers: { "X-CSRF-Token": csrfToken },
      });
      if (!response.ok) throw new Error("logout failed");

      router.replace("/login?notice=logout-success");
      router.refresh();
    } catch {
      setLoading(false);
      setModalVisible(false);
      setModalOpen(false);
      setNotice(ADMIN_AUTH_NOTICES.logoutFailed);
    }
  }

  return (
    <>
      <button ref={triggerRef} type="button" onClick={() => setModalOpen(true)} className="h-10 rounded-full border border-capa-line px-5 text-sm font-semibold text-capa-ink focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-capa-violet/30">Çıkış yap</button>
      {notice && <Snackbar message={notice.message} tone={notice.tone} onClose={() => setNotice(null)} />}
      {modalOpen && (
        <div className={`fixed inset-0 z-40 flex items-center justify-center bg-capa-ink/0 px-5 transition-colors duration-300 ${modalVisible ? "bg-capa-ink/40" : ""}`} onMouseDown={(event) => { if (event.target === event.currentTarget) closeModal(); }}>
          <div role="dialog" aria-modal="true" aria-labelledby="logout-title" aria-describedby="logout-description" className={`w-full max-w-md scale-95 rounded-xl border border-capa-line bg-capa-panel p-6 opacity-0 shadow-2xl transition-all duration-300 ${modalVisible ? "scale-100 opacity-100" : ""}`}>
            <h2 id="logout-title" className="text-xl font-semibold">Oturumu kapat?</h2>
            <p id="logout-description" className="mt-3 text-sm leading-6 text-capa-muted">Yönetici çalışma alanından çıkış yapmak istediğinize emin misiniz?</p>
            <div className="mt-6 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
              <button ref={cancelRef} type="button" onClick={closeModal} disabled={loading} className="h-11 rounded-full border border-capa-line px-5 text-sm font-semibold text-capa-ink focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-capa-violet/30 disabled:opacity-60">Vazgeç</button>
              <button ref={confirmRef} type="button" onClick={logout} disabled={loading} aria-busy={loading} className="h-11 rounded-full bg-capa-brand px-5 text-sm font-semibold text-white focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-capa-violet/30 disabled:opacity-60">{loading ? "Çıkış yapılıyor..." : "Çıkış yap"}</button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
