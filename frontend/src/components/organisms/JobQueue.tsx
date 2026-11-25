import { Check, CircleAlert, Clock3, FolderOpen, LoaderCircle, Pause, X } from "lucide-react";
import type { AggregateProgress, Job, JobStatus } from "../../types/domain";
import { formatBytes, formatETA, formatSpeed } from "../../lib/format";
import { Button } from "../atoms/Button";
import { Panel } from "../atoms/Panel";
import { ProgressBar } from "../atoms/ProgressBar";
import { StatusBadge } from "../atoms/StatusBadge";
import { safeDisplayURL } from "../../features/downloads/url-utils";

interface Props {
  jobs: Job[];
  aggregate: AggregateProgress;
  onCancel: (id: string) => void;
  onOpenFolder: (path: string) => void;
}

const activeStates = new Set<JobStatus>(["queued", "downloading", "converting"]);

function prettyStatus(status: JobStatus) {
  if (status === "completed") return <Check className="size-4 text-lime-200" aria-hidden="true" />;
  if (status === "failed")
    return <CircleAlert className="size-4 text-rose-200" aria-hidden="true" />;
  if (status === "cancelled") return <Pause className="size-4 text-amber-200" aria-hidden="true" />;
  if (status === "queued") return <Clock3 className="size-4 text-slate-500" aria-hidden="true" />;
  return <LoaderCircle className="size-4 animate-spin text-cyan-200" aria-hidden="true" />;
}

export function JobQueue({ jobs, aggregate, onCancel, onOpenFolder }: Props) {
  const activeCount = jobs.filter((job) => activeStates.has(job.status)).length;

  return (
    <Panel className="overflow-hidden">
      <div className="flex flex-wrap items-center justify-between gap-4 border-b border-white/[0.06] px-6 py-5 sm:px-7">
        <div>
          <div className="flex items-center gap-2.5">
            <h2 className="text-base font-semibold text-white">Download queue</h2>
            <span className="rounded-full bg-white/[0.06] px-2 py-0.5 text-xs tabular-nums text-slate-400">
              {jobs.length}
            </span>
          </div>
          <p className="mt-1 text-sm text-slate-500">
            {activeCount > 0 ? `${activeCount} active · ` : ""}
            {aggregate.finishedItems} of {aggregate.totalItems} items finished
          </p>
        </div>
        {jobs.length > 0 && (
          <div className="w-full max-w-56 sm:w-44">
            <div className="mb-2 flex items-center justify-between text-[11px] text-slate-500">
              <span>Item progress</span>
              <span>
                {aggregate.indeterminate || aggregate.percent === undefined
                  ? "—"
                  : `${Math.round(aggregate.percent)}%`}
              </span>
            </div>
            <ProgressBar
              label="Item-count queue progress"
              value={aggregate.percent}
              indeterminate={aggregate.indeterminate}
              compact
            />
          </div>
        )}
      </div>

      {jobs.length === 0 ? (
        <div className="grid min-h-48 place-items-center px-6 py-12 text-center">
          <div>
            <div className="mx-auto grid size-12 place-items-center rounded-2xl border border-white/[0.08] bg-white/[0.03] text-slate-500">
              <FolderOpen className="size-5" aria-hidden="true" />
            </div>
            <h3 className="mt-4 text-sm font-medium text-slate-300">Your queue is clear</h3>
            <p className="mt-1 text-sm text-slate-500">
              Add one or more video URLs to start a batch.
            </p>
          </div>
        </div>
      ) : (
        <ol className="divide-y divide-white/[0.05]">
          {jobs.map((job) => {
            const percent = job.progress.percent;
            const canCancel = activeStates.has(job.status);
            const indeterminate =
              (job.status === "downloading" || job.status === "converting") &&
              percent === undefined;
            return (
              <li key={job.id} className="px-6 py-4 sm:px-7">
                <div className="flex items-start gap-3">
                  <div className="mt-1 grid size-8 shrink-0 place-items-center rounded-xl bg-white/[0.035]">
                    {prettyStatus(job.status)}
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <h3
                        className="min-w-0 max-w-full truncate text-sm font-medium text-slate-100"
                        title={job.outputName}
                      >
                        {job.outputName}.{job.format}
                      </h3>
                      <StatusBadge status={job.status} />
                    </div>
                    <p
                      className="mt-1 truncate text-xs text-slate-500"
                      title={safeDisplayURL(job.url)}
                    >
                      {safeDisplayURL(job.url)}
                    </p>
                    {job.error && (
                      <p
                        className="mt-2 flex items-start gap-1.5 text-xs leading-5 text-rose-200"
                        role="status"
                      >
                        <CircleAlert className="mt-0.5 size-3.5 shrink-0" aria-hidden="true" />
                        <span>{job.error}</span>
                      </p>
                    )}
                    {job.status === "completed" && (
                      <p className="mt-2 text-xs text-slate-500">
                        {formatBytes(job.outputSize)} · {job.outputPath || job.outputDir}
                      </p>
                    )}
                    {(job.status === "downloading" || job.status === "converting") && (
                      <div className="mt-3 grid gap-2">
                        <ProgressBar
                          label={`${job.outputName} progress`}
                          value={percent}
                          indeterminate={indeterminate}
                          compact
                        />
                        <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-[11px] tabular-nums text-slate-500">
                          <span>
                            {percent === undefined
                              ? "Progress unavailable"
                              : `${Math.round(percent)}%`}
                          </span>
                          {job.progress.downloadedBytes !== undefined && (
                            <span>{formatBytes(job.progress.downloadedBytes)} downloaded</span>
                          )}
                          {job.progress.speedBytesPerSecond !== undefined && (
                            <span>{formatSpeed(job.progress.speedBytesPerSecond)}</span>
                          )}
                          {job.progress.etaSeconds !== undefined && (
                            <span>~{formatETA(job.progress.etaSeconds)} left</span>
                          )}
                        </div>
                      </div>
                    )}
                  </div>
                  <div className="flex shrink-0 items-center gap-1">
                    {job.status === "completed" && (
                      <Button
                        type="button"
                        variant="quiet"
                        className="min-h-8 px-2.5 text-xs"
                        onClick={() => onOpenFolder(job.outputDir)}
                        aria-label={`Open folder for ${job.outputName}`}
                      >
                        <FolderOpen className="size-3.5" aria-hidden="true" />
                      </Button>
                    )}
                    {canCancel && job.status !== "queued" ? (
                      <Button
                        type="button"
                        variant="quiet"
                        className="min-h-8 px-2.5 text-xs text-slate-500 hover:text-rose-200"
                        onClick={() => onCancel(job.id)}
                        aria-label={`Cancel ${job.outputName}`}
                        title="Cancel download"
                      >
                        <X className="size-4" aria-hidden="true" />
                      </Button>
                    ) : job.status === "queued" ? (
                      <Button
                        type="button"
                        variant="quiet"
                        className="min-h-8 px-2.5 text-xs text-slate-500 hover:text-rose-200"
                        onClick={() => onCancel(job.id)}
                        aria-label={`Cancel ${job.outputName}`}
                        title="Remove from queue"
                      >
                        <X className="size-4" aria-hidden="true" />
                      </Button>
                    ) : null}
                  </div>
                </div>
              </li>
            );
          })}
        </ol>
      )}
    </Panel>
  );
}
