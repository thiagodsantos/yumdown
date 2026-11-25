interface Props {
  value?: number;
  label: string;
  indeterminate?: boolean;
  compact?: boolean;
}

export function ProgressBar({ value, label, indeterminate = false, compact = false }: Props) {
  const normalized = Math.max(0, Math.min(100, value ?? 0));
  return (
    <div
      className={`progress-track ${compact ? "h-1.5" : "h-2"} ${indeterminate ? "is-indeterminate" : ""}`}
      role="progressbar"
      aria-label={label}
      aria-valuemin={0}
      aria-valuemax={100}
      aria-valuenow={indeterminate ? undefined : normalized}
      aria-valuetext={indeterminate ? "Progress is being determined" : `${Math.round(normalized)}%`}
    >
      <span
        className="progress-fill"
        style={indeterminate ? undefined : { width: `${normalized}%` }}
      />
    </div>
  );
}
