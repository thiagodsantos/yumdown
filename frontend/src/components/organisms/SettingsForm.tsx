import { Check, ExternalLink, Info, RefreshCw, ShieldCheck, Wrench } from "lucide-react";
import { Button } from "../atoms/Button";
import { Panel } from "../atoms/Panel";
import { FieldLabel } from "../molecules/FieldLabel";
import { FolderPicker } from "../molecules/FolderPicker";
import type { Capabilities, Settings } from "../../types/domain";

interface Props {
  settings: Settings;
  capabilities: Capabilities | null;
  saving: boolean;
  saved: boolean;
  error: string;
  onSettingsChange: (settings: Settings) => void;
  onChooseFolder: () => void;
  onOpenFolder: () => void;
  onSave: () => void;
  onRefreshTools: () => void;
}

const qualityOptions = ["64K", "96K", "128K", "160K", "192K", "256K", "320K"];

export function SettingsForm({
  settings,
  capabilities,
  saving,
  saved,
  error,
  onSettingsChange,
  onChooseFolder,
  onOpenFolder,
  onSave,
  onRefreshTools,
}: Props) {
  const formats = capabilities?.formats ?? [];
  return (
    <div className="grid gap-5 xl:grid-cols-[minmax(0,1fr)_360px]">
      <Panel className="p-6 sm:p-7">
        <div className="mb-6 flex items-start gap-4">
          <div className="grid size-11 shrink-0 place-items-center rounded-2xl bg-violet-300/10 text-violet-200 ring-1 ring-violet-200/15">
            <Wrench className="size-5" aria-hidden="true" />
          </div>
          <div>
            <h2 className="text-base font-semibold text-white">Download defaults</h2>
            <p className="mt-1 text-sm leading-6 text-slate-400">
              Saved locally and applied to new batches.
            </p>
          </div>
        </div>

        <div className="grid gap-5">
          <FieldLabel label="Default output folder">
            <FolderPicker
              value={settings.outputDir}
              onChoose={onChooseFolder}
              onOpen={onOpenFolder}
              disabled={saving}
            />
          </FieldLabel>

          <div className="grid gap-4 sm:grid-cols-2">
            <FieldLabel label="Default audio format" htmlFor="settings-format">
              <select
                id="settings-format"
                value={settings.format}
                onChange={(event) =>
                  onSettingsChange({
                    ...settings,
                    format: event.target.value as Settings["format"],
                  })
                }
                className="h-11 rounded-xl border border-white/10 bg-slate-950/60 px-3.5 text-sm text-slate-100 outline-none focus:border-lime-300/40"
                disabled={formats.length === 0 || saving}
              >
                {formats.length === 0 ? (
                  <option value={settings.format}>No supported formats detected</option>
                ) : (
                  formats.map((format) => (
                    <option value={format} key={format}>
                      {format.toUpperCase()}
                    </option>
                  ))
                )}
              </select>
            </FieldLabel>
            <FieldLabel label="Audio quality" htmlFor="settings-quality" hint="For lossy formats">
              <select
                id="settings-quality"
                value={settings.audioQuality}
                onChange={(event) =>
                  onSettingsChange({ ...settings, audioQuality: event.target.value })
                }
                className="h-11 rounded-xl border border-white/10 bg-slate-950/60 px-3.5 text-sm text-slate-100 outline-none focus:border-lime-300/40"
                disabled={saving}
              >
                {qualityOptions.map((quality) => (
                  <option key={quality} value={quality}>
                    {quality}
                  </option>
                ))}
              </select>
            </FieldLabel>
          </div>

          <FieldLabel
            label="Simultaneous downloads"
            htmlFor="settings-concurrency"
            hint="Maximum 4"
          >
            <select
              id="settings-concurrency"
              value={settings.concurrency}
              onChange={(event) =>
                onSettingsChange({ ...settings, concurrency: Number(event.target.value) })
              }
              className="h-11 rounded-xl border border-white/10 bg-slate-950/60 px-3.5 text-sm text-slate-100 outline-none focus:border-lime-300/40"
              disabled={saving}
            >
              {[1, 2, 3, 4].map((count) => (
                <option key={count} value={count}>
                  {count} {count === 1 ? "download" : "downloads"}
                </option>
              ))}
            </select>
          </FieldLabel>

          <div className="flex flex-col gap-3 border-t border-white/[0.06] pt-5 sm:flex-row sm:items-center sm:justify-between">
            <div className="min-h-5 text-sm" aria-live="polite">
              {error ? (
                <span className="text-rose-200">{error}</span>
              ) : saved ? (
                <span className="inline-flex items-center gap-1.5 text-lime-200">
                  <Check className="size-4" aria-hidden="true" />
                  Settings saved
                </span>
              ) : (
                <span className="text-slate-500">Changes are not saved yet.</span>
              )}
            </div>
            <Button type="button" variant="primary" onClick={onSave} disabled={saving}>
              {saving ? "Saving…" : "Save settings"}
            </Button>
          </div>
        </div>
      </Panel>

      <div className="grid content-start gap-5">
        <Panel className="overflow-hidden">
          <div className="flex items-center justify-between border-b border-white/[0.06] px-5 py-4">
            <div className="flex items-center gap-2.5">
              <span
                className={`size-2 rounded-full ${capabilities?.toolsAvailable ? "bg-lime-300 shadow-[0_0_12px_rgba(190,242,100,0.55)]" : "bg-amber-300"}`}
              />
              <h2 className="text-sm font-semibold text-white">Media tools</h2>
            </div>
            <Button
              type="button"
              variant="quiet"
              className="min-h-8 px-2"
              onClick={onRefreshTools}
              aria-label="Check media tools again"
            >
              <RefreshCw className="size-3.5" aria-hidden="true" />
            </Button>
          </div>
          <div className="grid gap-4 p-5">
            <div className="flex items-start gap-3 rounded-xl border border-white/[0.06] bg-white/[0.02] p-3.5">
              <ShieldCheck className="mt-0.5 size-4 shrink-0 text-lime-200" aria-hidden="true" />
              <div>
                <p className="text-sm font-medium text-slate-200">
                  {capabilities?.toolsAvailable ? "Ready to process" : "Setup required"}
                </p>
                <p className="mt-1 break-words text-xs leading-5 text-slate-500">
                  {capabilities?.message || "Tool status has not been checked."}
                </p>
              </div>
            </div>
            {capabilities?.toolsAvailable && (
              <dl className="grid gap-2.5 text-xs">
                <div className="flex justify-between gap-3">
                  <dt className="text-slate-500">yt-dlp</dt>
                  <dd className="truncate text-slate-300">{capabilities.ytDlpVersion}</dd>
                </div>
                <div className="flex justify-between gap-3">
                  <dt className="text-slate-500">ffmpeg</dt>
                  <dd className="truncate text-slate-300">
                    {capabilities.ffmpegVersion?.split(" ").slice(0, 3).join(" ")}
                  </dd>
                </div>
                <div className="flex justify-between gap-3">
                  <dt className="text-slate-500">ffprobe</dt>
                  <dd className="text-slate-300">
                    {capabilities.ffprobeFound ? "Available" : "Optional · not found"}
                  </dd>
                </div>
                <div className="grid gap-1">
                  <dt className="text-slate-500">Detected formats</dt>
                  <dd className="text-slate-300">
                    {formats.length
                      ? formats.map((format) => format.toUpperCase()).join(" · ")
                      : "None"}
                  </dd>
                </div>
              </dl>
            )}
            {capabilities?.toolsDirectory && (
              <div className="border-t border-white/[0.06] pt-3">
                <p className="text-[11px] uppercase tracking-wider text-slate-600">
                  Tool directory
                </p>
                <p className="mt-1 break-all text-xs text-slate-400">
                  {capabilities.toolsDirectory}
                </p>
              </div>
            )}
          </div>
        </Panel>

        <Panel className="p-5">
          <div className="flex gap-3">
            <Info className="mt-0.5 size-4 shrink-0 text-cyan-200" aria-hidden="true" />
            <div>
              <h3 className="text-sm font-semibold text-slate-200">Install tools yourself</h3>
              <p className="mt-1.5 text-xs leading-5 text-slate-500">
                Yumdown does not download or update executables automatically. Add the matching
                yt-dlp and ffmpeg binaries to the app’s tools folder, or set{" "}
                <code className="text-slate-300">YUMDOWN_TOOLS_DIR</code> to an absolute directory.
              </p>
              <p className="mt-2 inline-flex items-center gap-1 text-xs text-slate-400">
                <ExternalLink className="size-3" aria-hidden="true" />
                Use official releases and review their licenses.
              </p>
            </div>
          </div>
        </Panel>
      </div>
    </div>
  );
}
