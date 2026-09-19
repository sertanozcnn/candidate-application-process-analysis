"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";

type SnackbarTone = "error" | "success" | "info";
type Snackbar = { message: string; tone: SnackbarTone };
type SnackbarContextValue = { showSnackbar: (message: string, tone?: SnackbarTone) => void };

const SnackbarContext = createContext<SnackbarContextValue | null>(null);

const toneClasses: Record<SnackbarTone, string> = {
  error: "border-red-200 bg-red-50 text-red-900",
  success: "border-capa-green/40 bg-capa-green/10 text-capa-ink",
  info: "border-capa-violet/30 bg-capa-violet/10 text-capa-ink",
};

export function SnackbarProvider({ children }: { children: ReactNode }) {
  const [snackbar, setSnackbar] = useState<Snackbar | null>(null);

  const showSnackbar = useCallback((message: string, tone: SnackbarTone = "info") => {
    setSnackbar({ message, tone });
  }, []);

  useEffect(() => {
    if (!snackbar) return;
    const timeout = window.setTimeout(() => setSnackbar(null), 5000);
    return () => window.clearTimeout(timeout);
  }, [snackbar]);

  const value = useMemo(() => ({ showSnackbar }), [showSnackbar]);

  return (
    <SnackbarContext.Provider value={value}>
      {children}
      {snackbar && (
        <div className="pointer-events-none fixed inset-x-4 bottom-4 z-50 flex justify-center sm:inset-x-auto sm:right-6 sm:justify-end" role="status" aria-live="polite">
          <div className={`pointer-events-auto w-full max-w-md rounded-xl border px-4 py-3 text-sm font-medium shadow-lg shadow-capa-ink/10 ${toneClasses[snackbar.tone]}`}>
            {snackbar.message}
          </div>
        </div>
      )}
    </SnackbarContext.Provider>
  );
}

export function useSnackbar() {
  const context = useContext(SnackbarContext);
  if (!context) throw new Error("useSnackbar must be used inside SnackbarProvider");
  return context;
}
