import type { PropsWithChildren } from "react";

export function Panel({ children, className = "" }: PropsWithChildren<{ className?: string }>) {
  return (
    <section
      className={`rounded-2xl border border-white/[0.075] bg-slate-900/65 shadow-[0_18px_70px_rgba(0,0,0,0.18)] ${className}`}
    >
      {children}
    </section>
  );
}
