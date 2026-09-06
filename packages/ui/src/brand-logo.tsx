export function BrandLogo() {
  return (
    <span className="inline-flex items-center gap-2.5" aria-label="CAPA">
      <span aria-hidden="true" className="grid size-8 grid-cols-2 gap-0.5 -rotate-6">
        <span className="rounded-tl-xl rounded-br-sm bg-linear-to-br from-capa-coral to-capa-rose" />
        <span className="rounded-tr-xl rounded-bl-sm bg-linear-to-br from-capa-rose to-capa-violet" />
        <span className="rounded-bl-xl rounded-tr-sm bg-linear-to-br from-capa-green to-capa-violet" />
        <span className="rounded-br-xl rounded-tl-sm bg-linear-to-br from-capa-violet to-capa-brand" />
      </span>
      <span aria-hidden="true" className="bg-linear-to-r from-capa-ink via-capa-brand to-capa-violet bg-clip-text text-3xl font-extrabold tracking-[-0.06em] text-transparent">capa.</span>
    </span>
  );
}
