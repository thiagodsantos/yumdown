import type { PropsWithChildren, ReactNode } from "react";

interface Props extends PropsWithChildren {
  label: string;
  hint?: ReactNode;
  htmlFor?: string;
  required?: boolean;
}

export function FieldLabel({ label, hint, htmlFor, required = false, children }: Props) {
  return (
    <div className="grid gap-2">
      <div className="flex items-center justify-between gap-3">
        <label htmlFor={htmlFor} className="text-sm font-medium text-slate-200">
          {label} {required && <span className="text-lime-300">*</span>}
        </label>
        {hint && <span className="text-xs text-slate-500">{hint}</span>}
      </div>
      {children}
    </div>
  );
}
