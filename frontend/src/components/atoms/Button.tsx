import { Button as BaseButton } from "@base-ui/react/button";
import type { ComponentProps } from "react";

type Variant = "primary" | "secondary" | "quiet" | "danger";

type Props = ComponentProps<typeof BaseButton> & { variant?: Variant };

const variants: Record<Variant, string> = {
  primary:
    "bg-lime-300 text-slate-950 shadow-[0_8px_30px_rgba(190,242,100,0.15)] hover:bg-lime-200 disabled:bg-slate-700 disabled:text-slate-400",
  secondary: "border border-white/10 bg-white/[0.04] text-slate-100 hover:bg-white/[0.08]",
  quiet: "text-slate-400 hover:bg-white/[0.06] hover:text-white",
  danger: "border border-rose-400/20 bg-rose-400/10 text-rose-200 hover:bg-rose-400/20",
};

export function Button({ variant = "secondary", className = "", ...props }: Props) {
  return (
    <BaseButton
      className={`inline-flex min-h-10 items-center justify-center gap-2 rounded-xl px-4 text-sm font-semibold transition duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-lime-300/80 disabled:cursor-not-allowed disabled:opacity-60 ${variants[variant]} ${className}`}
      {...props}
    />
  );
}
