import type { JobStatus } from "../../types/domain";

const labels: Record<JobStatus, string> = {
  queued: "Queued",
  downloading: "Downloading",
  converting: "Converting",
  completed: "Complete",
  failed: "Failed",
  cancelled: "Cancelled",
};

const styles: Record<JobStatus, string> = {
  queued: "border-white/10 bg-white/[0.04] text-slate-400",
  downloading: "border-cyan-300/20 bg-cyan-300/10 text-cyan-200",
  converting: "border-violet-300/20 bg-violet-300/10 text-violet-200",
  completed: "border-lime-300/20 bg-lime-300/10 text-lime-200",
  failed: "border-rose-300/20 bg-rose-300/10 text-rose-200",
  cancelled: "border-amber-300/20 bg-amber-300/10 text-amber-200",
};

export function StatusBadge({ status }: { status: JobStatus }) {
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-[11px] font-semibold tracking-wide ${styles[status]}`}
    >
      <span className="size-1.5 rounded-full bg-current" aria-hidden="true" />
      {labels[status]}
    </span>
  );
}
