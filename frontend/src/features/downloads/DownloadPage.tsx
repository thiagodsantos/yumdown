import type {
  AggregateProgress,
  Capabilities,
  DownloadRequest,
  Job,
  Settings,
} from "../../types/domain";
import { DownloadForm } from "../../components/organisms/DownloadForm";
import { JobQueue } from "../../components/organisms/JobQueue";

interface Props {
  capabilities: Capabilities | null;
  settings: Settings;
  jobs: Job[];
  aggregate: AggregateProgress;
  submitting: boolean;
  error: string;
  onSettingsChange: (settings: Settings) => void;
  onChooseFolder: () => void;
  onOpenFolder: (path: string) => void;
  onStart: (request: DownloadRequest) => void;
  onCancel: (id: string) => void;
}

export function DownloadPage({
  capabilities,
  settings,
  jobs,
  aggregate,
  submitting,
  error,
  onSettingsChange,
  onChooseFolder,
  onOpenFolder,
  onStart,
  onCancel,
}: Props) {
  return (
    <>
      <section className="relative overflow-hidden rounded-2xl border border-white/[0.06] bg-gradient-to-br from-[#18251c] via-[#111923] to-[#11131c] px-6 py-6 sm:px-8 sm:py-7">
        <div className="pointer-events-none absolute -right-10 -top-28 size-72 rounded-full bg-lime-300/[0.08] blur-3xl" />
        <div className="relative flex flex-col justify-between gap-6 sm:flex-row sm:items-end">
          <div className="max-w-xl">
            <div className="mb-3 inline-flex items-center gap-2 rounded-full border border-lime-200/15 bg-lime-200/[0.05] px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.15em] text-lime-100/80">
              <span className="size-1.5 rounded-full bg-lime-300" /> Batch audio extraction
            </div>
            <h2 className="text-2xl font-semibold leading-tight tracking-tight text-white sm:text-[28px]">
              Your videos, in audio.
            </h2>
            <p className="mt-2 max-w-lg text-sm leading-6 text-slate-400">
              Queue multiple links, choose a format, and let each download run independently while
              you keep an eye on progress.
            </p>
          </div>
          <div className="grid grid-cols-2 gap-2 sm:min-w-56">
            <div className="rounded-xl border border-white/[0.07] bg-black/15 px-3.5 py-3">
              <p className="text-[10px] font-semibold uppercase tracking-[0.14em] text-slate-600">
                In queue
              </p>
              <p className="mt-1 text-xl font-semibold tabular-nums text-white">
                {jobs.length.toString().padStart(2, "0")}
              </p>
            </div>
            <div className="rounded-xl border border-white/[0.07] bg-black/15 px-3.5 py-3">
              <p className="text-[10px] font-semibold uppercase tracking-[0.14em] text-slate-600">
                Finished
              </p>
              <p className="mt-1 text-xl font-semibold tabular-nums text-lime-200">
                {aggregate.finishedItems.toString().padStart(2, "0")}
              </p>
            </div>
          </div>
        </div>
      </section>

      <DownloadForm
        capabilities={capabilities}
        settings={settings}
        onSettingsChange={onSettingsChange}
        onChooseFolder={onChooseFolder}
        onOpenFolder={() => onOpenFolder(settings.outputDir)}
        onSubmit={onStart}
        submitting={submitting}
        error={error}
      />
      <JobQueue jobs={jobs} aggregate={aggregate} onCancel={onCancel} onOpenFolder={onOpenFolder} />
    </>
  );
}
