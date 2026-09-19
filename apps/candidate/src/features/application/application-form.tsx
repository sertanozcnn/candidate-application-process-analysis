"use client";

import { Button } from "@capa/ui";
import { FormEvent, useEffect, useRef, useState } from "react";

import { ApplicationRequestError, submitApplication, type ApplicationInput } from "@/lib/application-api";
import { useSnackbar } from "@/components/snackbar";
import { useInteractionTracker } from "@/features/interaction/interaction-tracker";

const initialValues: ApplicationInput = {
  full_name: "",
  email: "",
  position_code: "",
  experience: "",
};

const positions = [
  ["frontend", "Frontend geliştirici"],
  ["backend", "Backend geliştirici"],
  ["fullstack", "Fullstack geliştirici"],
] as const;

type FormErrors = Partial<Record<keyof ApplicationInput, string>>;
type SubmitState = "idle" | "loading" | "success" | "error";

function validate(values: ApplicationInput): FormErrors {
  const errors: FormErrors = {};
  const name = values.full_name.trim();
  const email = values.email.trim();
  const experience = values.experience.trim();

  if (name.length < 2 || name.length > 120) errors.full_name = "Ad soyad 2–120 karakter arasında olmalı.";
  if (!/^\S+@\S+\.\S+$/.test(email) || email.length > 254) errors.email = "Geçerli bir e-posta adresi yazın.";
  if (!positions.some(([code]) => code === values.position_code)) errors.position_code = "Bir pozisyon seçin.";
  if (experience.length < 20 || experience.length > 5000) errors.experience = "Deneyim açıklaması 20–5000 karakter arasında olmalı.";

  return errors;
}

function getSubmitMessage(error: unknown): string {
  if (error instanceof ApplicationRequestError) {
    if (error.status === 400) return "Bilgileri kontrol edip tekrar deneyin.";
    if (error.status === 409) return "Bu pozisyon için aynı e-posta ile daha önce başvuru yapılmış.";
    if (error.status >= 500) return "Sunucuda geçici bir sorun oluştu. Lütfen daha sonra tekrar deneyin.";
  }
  return "Bağlantı kurulamadı. İnternet bağlantınızı kontrol edip tekrar deneyin.";
}

function getValidationMessage(errors: FormErrors): string {
  if (errors.full_name) return "Ad soyad alanını kontrol edin.";
  if (errors.email) return "E-posta adresini kontrol edin.";
  if (errors.position_code) return "Bir pozisyon seçin.";
  return "Deneyim açıklamasını kontrol edin.";
}

export function ApplicationForm() {
  const [values, setValues] = useState<ApplicationInput>(initialValues);
  const [errors, setErrors] = useState<FormErrors>({});
  const [submitState, setSubmitState] = useState<SubmitState>("idle");
  const { showSnackbar } = useSnackbar();
  const tracker = useInteractionTracker(values.position_code);

  function updateField(field: keyof ApplicationInput, value: string) {
    setValues((current) => ({ ...current, [field]: value }));
    setErrors((current) => ({ ...current, [field]: undefined }));
    if (field === "position_code") tracker.track("field_focus", "position_code");
    if (submitState === "error") setSubmitState("idle");
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const nextErrors = validate(values);
    setErrors(nextErrors);
    if (Object.keys(nextErrors).length > 0) {
      setSubmitState("error");
      for (const field of Object.keys(nextErrors) as Array<keyof ApplicationInput>) tracker.track("validation_error", field);
      showSnackbar(getValidationMessage(nextErrors), "error");
      return;
    }

    setSubmitState("loading");
    try {
      await submitApplication({
        full_name: values.full_name.trim(),
        email: values.email.trim().toLowerCase(),
        position_code: values.position_code,
        experience: values.experience.trim(),
      }, tracker.sessionID);
      setSubmitState("success");
      tracker.finish();
      showSnackbar("Başvurun başarıyla alındı.", "success");
    } catch (error) {
      setSubmitState("error");
      showSnackbar(getSubmitMessage(error), "error");
    }
  }

  if (submitState === "success") {
    return (
      <div className="rounded-lg border border-capa-green/40 bg-capa-green/10 p-6 text-left sm:p-8" role="status" aria-live="polite">
        <p className="text-xs font-semibold uppercase tracking-[0.16em] text-capa-green">Başvurun alındı</p>
        <h3 className="mt-3 text-2xl font-semibold tracking-tight">Teşekkürler.</h3>
        <p className="mt-3 text-sm leading-6 text-capa-muted">Başvurun başarıyla kaydedildi. Ek bir işlem yapmana gerek yok.</p>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} noValidate className="grid gap-5" aria-describedby="form-status">
      <div className="grid gap-5 sm:grid-cols-2">
        <Field label="Ad soyad" name="full_name" value={values.full_name} error={errors.full_name} onChange={(value) => updateField("full_name", value)} onFocus={() => tracker.track("field_focus", "full_name")} onBlur={() => tracker.track("field_blur", "full_name")} onPaste={() => tracker.track("field_paste", "full_name", { count: 1 })} placeholder="Örn. Test Kullanıcı" />
        <Field label="E-posta" name="email" type="email" value={values.email} error={errors.email} onChange={(value) => updateField("email", value)} onFocus={() => tracker.track("field_focus", "email")} onBlur={() => tracker.track("field_blur", "email")} onPaste={() => tracker.track("field_paste", "email", { count: 1 })} placeholder="test@example.invalid" />
      </div>
      <PositionSelect value={values.position_code} error={errors.position_code} onChange={(value) => updateField("position_code", value)} onFocus={() => tracker.track("field_focus", "position_code")} onBlur={() => tracker.track("field_blur", "position_code")} />
      <label className="grid gap-2 text-sm font-semibold" htmlFor="experience">
        Deneyim
        <textarea id="experience" name="experience" value={values.experience} onChange={(event) => updateField("experience", event.target.value)} onFocus={() => tracker.track("field_focus", "experience")} onBlur={() => tracker.track("field_blur", "experience")} onPaste={() => tracker.track("field_paste", "experience", { count: 1 })} aria-invalid={Boolean(errors.experience)} aria-describedby={errors.experience ? "experience-error" : undefined} placeholder="Deneyimini ve üzerinde çalıştığın işleri kısaca anlat." rows={6} className="resize-y rounded-lg border border-capa-line bg-capa-panel px-4 py-3 font-normal leading-6 text-capa-ink outline-none transition placeholder:text-capa-muted/70 focus:border-capa-violet focus:ring-4 focus:ring-capa-violet/20" />
        <span className="text-xs font-normal text-capa-muted">En az 20 karakter.</span>
        {errors.experience && <span id="experience-error" className="sr-only">{errors.experience}</span>}
      </label>
      <div id="form-status" aria-live="polite" className="sr-only">Form gönderim durumu</div>
      <Button type="submit" disabled={submitState === "loading"} arrow={false} className="w-full sm:w-fit">{submitState === "loading" ? "Gönderiliyor…" : "Başvuruyu gönder"}</Button>
    </form>
  );
}

type FieldProps = { label: string; name: keyof ApplicationInput; type?: string; value: string; error?: string; placeholder: string; onChange: (value: string) => void; onFocus: () => void; onBlur: () => void; onPaste: () => void };

function Field({ label, name, type = "text", value, error, placeholder, onChange, onFocus, onBlur, onPaste }: FieldProps) {
  const errorId = `${name}-error`;
  return (
    <label className="grid gap-2 text-sm font-semibold" htmlFor={name}>
      {label}
      <input id={name} name={name} type={type} value={value} onChange={(event) => onChange(event.target.value)} onFocus={onFocus} onBlur={onBlur} onPaste={onPaste} placeholder={placeholder} autoComplete={name === "full_name" ? "name" : "email"} aria-invalid={Boolean(error)} aria-describedby={error ? errorId : undefined} className="h-12 rounded-lg border border-capa-line bg-capa-panel px-4 font-normal text-capa-ink outline-none transition focus:border-capa-violet focus:ring-4 focus:ring-capa-violet/20" />
      {error && <span id={errorId} className="sr-only">{error}</span>}
    </label>
  );
}

function PositionSelect({ value, error, onChange, onFocus, onBlur }: { value: string; error?: string; onChange: (value: string) => void; onFocus: () => void; onBlur: () => void }) {
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const selected = positions.find(([code]) => code === value);

  useEffect(() => {
    function closeOnOutsideClick(event: MouseEvent) {
      if (!containerRef.current?.contains(event.target as Node)) setOpen(false);
    }
    document.addEventListener("mousedown", closeOnOutsideClick);
    return () => document.removeEventListener("mousedown", closeOnOutsideClick);
  }, []);

  function selectPosition(code: string) {
    onChange(code);
    setOpen(false);
  }

  return (
    <div ref={containerRef} className="grid gap-2 text-sm font-semibold">
      <span id="position-label">Pozisyon</span>
      <div className="relative">
        <button id="position_code" type="button" aria-labelledby="position-label" aria-haspopup="listbox" aria-expanded={open} aria-controls="position-options" aria-describedby={error ? "position_code-error" : undefined} data-invalid={Boolean(error)} onClick={() => setOpen((current) => !current)} onFocus={onFocus} onBlur={onBlur} onKeyDown={(event) => { if (event.key === "Escape") setOpen(false); }} className={`flex h-14 w-full items-center justify-between rounded-xl border bg-capa-bg/70 px-4 text-left font-normal outline-none transition hover:border-capa-violet/50 focus:border-capa-violet focus:ring-4 focus:ring-capa-violet/20 ${open ? "border-capa-violet ring-4 ring-capa-violet/20" : "border-capa-line"}`}>
          <span className={selected ? "text-capa-ink" : "text-capa-muted"}>{selected?.[1] ?? "Pozisyon seçin"}</span>
          <span aria-hidden="true" className={`size-2.5 shrink-0 -translate-y-1/4 rotate-45 border-b-2 border-r-2 border-capa-violet transition-transform ${open ? "translate-y-1/4 rotate-[225deg]" : ""}`} />
        </button>
        {open && (
          <div id="position-options" role="listbox" aria-labelledby="position-label" className="absolute inset-x-0 top-[calc(100%+0.5rem)] z-20 overflow-hidden rounded-xl border border-capa-line bg-capa-panel p-1.5 shadow-xl shadow-capa-violet/10">
            {positions.map(([code, label]) => (
              <button key={code} type="button" role="option" aria-selected={value === code} onClick={() => selectPosition(code)} className={`flex w-full items-center justify-between rounded-lg px-3 py-3 text-left text-sm font-medium transition-colors ${value === code ? "bg-capa-violet/10 text-capa-violet" : "text-capa-ink hover:bg-capa-surface hover:text-capa-violet"}`}>
                {label}
                {value === code && <span aria-hidden="true" className="text-base">✓</span>}
              </button>
            ))}
          </div>
        )}
      </div>
      {error && <span id="position_code-error" className="sr-only">{error}</span>}
    </div>
  );
}
