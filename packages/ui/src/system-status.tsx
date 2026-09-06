export type SystemStatusState = "ready" | "offline";

type SystemStatusProps = {
  label: string;
  detail: string;
  state: SystemStatusState;
};

export function SystemStatus({ label, detail, state }: SystemStatusProps) {
  const isReady = state === "ready";

  return (
    <div className="inline-flex max-w-full items-center gap-3 rounded-full border border-capa-line bg-capa-panel/75 px-3.5 py-2 text-left shadow-sm shadow-capa-violet/5 backdrop-blur">
      <span
        className={[
          "relative flex size-2.5 shrink-0 rounded-full",
          isReady ? "bg-emerald-400" : "bg-capa-coral",
        ].join(" ")}
        aria-hidden="true"
      >
        <span
          className={[
            "absolute inset-0 rounded-full opacity-40 blur-[2px]",
            isReady ? "bg-emerald-400" : "bg-capa-coral",
          ].join(" ")}
        />
      </span>
      <span className="min-w-0">
        <span className="block truncate text-[11px] font-semibold text-capa-ink">{label}</span>
        <span className="block truncate text-[10px] font-medium text-capa-muted">{detail}</span>
      </span>
    </div>
  );
}
