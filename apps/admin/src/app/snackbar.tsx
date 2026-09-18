"use client";

import { useEffect, useState } from "react";

type SnackbarProps = {
  message: string;
  tone?: "error" | "success" | "info";
  onClose: () => void;
};

export function Snackbar({ message, tone = "info", onClose }: SnackbarProps) {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const enter = requestAnimationFrame(() => setVisible(true));
    return () => cancelAnimationFrame(enter);
  }, []);

  function dismiss() {
    setVisible(false);
    setTimeout(onClose, 250);
  }

  const styles = {
    error: "border-capa-brand/30 bg-capa-brand text-white",
    success: "border-capa-green/40 bg-capa-green text-white",
    info: "border-capa-violet/30 bg-capa-violet text-white",
  };

  return (
    <div className="pointer-events-none fixed inset-x-4 bottom-8 z-50 flex justify-center sm:inset-x-auto sm:right-6 sm:bottom-6 sm:w-[min(24rem,calc(100vw-3rem))]" role={tone === "error" ? "alert" : "status"} aria-live={tone === "error" ? "assertive" : "polite"}>
      <div className={`pointer-events-auto flex w-full translate-y-4 items-start gap-3 rounded-xl border px-4 pb-4 pt-3 text-sm shadow-xl opacity-0 transition-all duration-300 ease-out ${visible ? "translate-y-0 opacity-100" : ""} ${styles[tone]}`}>
        <p className="min-w-0 flex-1 leading-5">{message}</p>
        <button type="button" onClick={dismiss} aria-label="Bildirimi kapat" className="shrink-0 rounded-md px-1 text-lg leading-5 opacity-90 transition-opacity hover:opacity-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white">×</button>
      </div>
    </div>
  );
}
