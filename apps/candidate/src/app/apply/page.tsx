import { SystemStatus } from "@capa/ui";

import { ApplicationForm } from "@/features/application/application-form";
import { getApiHealth } from "@/lib/api-health";
import { CandidateShell } from "@/components/candidate-shell";

export const dynamic = "force-dynamic";

export default async function ApplyPage() {
  const apiHealth = await getApiHealth();

  return (
    <CandidateShell>
      <section aria-labelledby="apply-title" className="w-full min-w-0 px-5 py-12 sm:px-10 sm:py-20">
        <div className="mx-auto grid w-full min-w-0 max-w-[1040px] gap-10 lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)] lg:items-start lg:gap-20">
          <div className="min-w-0 lg:sticky lg:top-8">
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-capa-brand">Başvuru formu</p>
            <h1 id="apply-title" className="mt-4 text-4xl font-semibold leading-tight tracking-[-0.04em] sm:text-5xl">Sıradaki adımını anlat.</h1>
            <p className="mt-5 max-w-md text-sm leading-7 text-capa-muted sm:text-base">Kısa ve açık bilgiler yeterli. Form gönderilene kadar bilgilerin yalnızca bu sayfada tutulur.</p>
            <div className="mt-7"><SystemStatus {...apiHealth} /></div>
          </div>
          <div className="min-w-0 rounded-2xl border border-capa-line bg-capa-panel/90 p-5 shadow-xl shadow-capa-violet/5 sm:p-8">
            <ApplicationForm />
          </div>
        </div>
      </section>
    </CandidateShell>
  );
}
