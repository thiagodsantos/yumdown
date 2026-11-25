import { ArrowUpRight, AudioLines, Plus, ShieldCheck, Sparkles, Trash2 } from "lucide-react";
import { useState, type FormEvent } from "react";
import type { Capabilities, DownloadRequest, Settings, VideoRequest } from "../../types/domain";
import { normalizeVideoRows, validateVideoURL } from "../../features/downloads/url-utils";
import { Button } from "../atoms/Button";
import { Panel } from "../atoms/Panel";
import { FieldLabel } from "../molecules/FieldLabel";
import { FolderPicker } from "../molecules/FolderPicker";

interface Props {
  capabilities: Capabilities | null;
  settings: Settings;
  onSettingsChange: (settings: Settings) => void;
  onChooseFolder: () => void;
  onOpenFolder: () => void;
  onSubmit: (request: DownloadRequest) => void;
  submitting: boolean;
  error: string;
}

const emptyVideo: VideoRequest = { url: "", filename: "" };

export function DownloadForm({
  capabilities,
  settings,
  onSettingsChange,
  onChooseFolder,
  onOpenFolder,
  onSubmit,
  submitting,
  error,
}: Props) {
  const [videos, setVideos] = useState<VideoRequest[]>([{ ...emptyVideo }]);
  const [validationError, setValidationError] = useState("");
  const formats = capabilities?.formats ?? [];
  const ready = Boolean(capabilities?.toolsAvailable && formats.length > 0);
  const rowIssues = videos.map((video, index) => {
    const url = video.url.trim();
    if (!url) return video.filename.trim() ? `Video row ${index + 1} needs a URL.` : null;
    return validateVideoURL(url, index + 1);
  });
  const nonEmptyRows = videos.filter((video) => video.url.trim() || video.filename.trim()).length;
  const inlineURLIssues = rowIssues.filter((issue): issue is string => issue !== null);
  if (nonEmptyRows > 100) inlineURLIssues.push("A batch can contain at most 100 videos.");

  function updateVideo(index: number, field: keyof VideoRequest, value: string) {
    setVideos((current) =>
      current.map((video, currentIndex) =>
        currentIndex === index ? { ...video, [field]: value } : video,
      ),
    );
  }

  function addVideo() {
    setVideos((current) => (current.length < 100 ? [...current, { ...emptyVideo }] : current));
  }

  function clearVideos() {
    setVideos([{ ...emptyVideo }]);
    setValidationError("");
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setValidationError("");
    let normalizedVideos: VideoRequest[];
    try {
      normalizedVideos = normalizeVideoRows(videos);
    } catch (error) {
      if (inlineURLIssues.length === 0) {
        setValidationError(
          error instanceof Error ? error.message : "Check the video rows and try again.",
        );
      }
      return;
    }
    if (!settings.outputDir) {
      setValidationError("Choose an output folder before starting.");
      return;
    }
    if (!settings.format || !formats.includes(settings.format)) {
      setValidationError("Choose an audio format supported by the installed tools.");
      return;
    }
    onSubmit({
      videos: normalizedVideos,
      outputDir: settings.outputDir,
      format: settings.format,
      audioQuality: settings.audioQuality,
    });
  }

  return (
    <Panel className="overflow-hidden">
      <div className="border-b border-white/[0.06] px-6 py-5 sm:px-7">
        <div className="flex items-start gap-4">
          <div className="grid size-11 shrink-0 place-items-center rounded-2xl bg-lime-300/10 text-lime-200 ring-1 ring-lime-200/15">
            <AudioLines className="size-5" aria-hidden="true" />
          </div>
          <div className="min-w-0 flex-1">
            <h2 className="text-base font-semibold text-white">Create an audio batch</h2>
            <p className="mt-1 text-sm leading-6 text-slate-400">
              Add a video URL and optional output name on each row.
            </p>
          </div>
          <span className="hidden items-center gap-1.5 rounded-full border border-white/[0.07] bg-white/[0.025] px-3 py-1.5 text-[11px] text-slate-400 sm:inline-flex">
            <ShieldCheck className="size-3.5 text-lime-300" aria-hidden="true" />
            Local processing
          </span>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="grid gap-5 p-6 sm:p-7">
        <section aria-labelledby="videos-heading" className="grid gap-3">
          <div className="flex items-center justify-between gap-3">
            <div>
              <h3 id="videos-heading" className="text-sm font-semibold text-slate-200">
                Video URLs <span className="text-lime-300">*</span>
              </h3>
              <p id="url-hint" className="mt-1 text-xs text-slate-500">
                One video per row. Only use links to content you have permission to download.
              </p>
            </div>
            <Button
              type="button"
              variant="quiet"
              className="min-h-9 shrink-0 px-2.5 text-xs"
              onClick={clearVideos}
              disabled={submitting}
              aria-label="Clear all video rows"
            >
              <Trash2 className="size-3.5" aria-hidden="true" />
              Clear
            </Button>
          </div>

          <div className="grid gap-3">
            {videos.map((video, index) => {
              const urlId = `video-url-${index}`;
              const filenameId = `video-filename-${index}`;
              const issueId = `video-error-${index}`;
              const issue = rowIssues[index];
              return (
                <div
                  key={index}
                  className="grid gap-3 rounded-xl border border-white/[0.07] bg-slate-950/35 p-3 sm:grid-cols-[minmax(0,1.4fr)_minmax(0,1fr)] sm:items-start"
                >
                  <div className="grid min-w-0 gap-1.5">
                    <label htmlFor={urlId} className="text-xs font-medium text-slate-400">
                      Video URL {index + 1} <span className="text-lime-300">*</span>
                    </label>
                    <input
                      id={urlId}
                      aria-label={`Video URL ${index + 1}`}
                      type="url"
                      inputMode="url"
                      value={video.url}
                      onChange={(event) => updateVideo(index, "url", event.target.value)}
                      placeholder="https://video.example/watch?v=…"
                      className="h-10 min-w-0 rounded-lg border border-white/10 bg-slate-950/65 px-3 text-sm text-slate-100 outline-none transition placeholder:text-slate-600 focus:border-lime-300/40 focus:ring-2 focus:ring-lime-300/10"
                      aria-invalid={Boolean(issue)}
                      aria-describedby={issue ? issueId : "url-hint"}
                      disabled={submitting}
                    />
                    {issue && (
                      <p id={issueId} role="status" className="text-xs leading-5 text-amber-200">
                        {issue}
                      </p>
                    )}
                  </div>
                  <div className="grid min-w-0 gap-1.5">
                    <label htmlFor={filenameId} className="text-xs font-medium text-slate-400">
                      Output name <span className="text-slate-600">· optional</span>
                    </label>
                    <input
                      id={filenameId}
                      aria-label={`Output name ${index + 1}`}
                      value={video.filename}
                      onChange={(event) => updateVideo(index, "filename", event.target.value)}
                      placeholder="Generated if left blank"
                      className="h-10 min-w-0 rounded-lg border border-white/10 bg-slate-950/65 px-3 text-sm text-slate-100 outline-none transition placeholder:text-slate-600 focus:border-lime-300/40 focus:ring-2 focus:ring-lime-300/10"
                      disabled={submitting}
                    />
                  </div>
                </div>
              );
            })}
          </div>

          <div className="flex flex-wrap items-center justify-between gap-3">
            <Button
              type="button"
              variant="secondary"
              className="min-h-9 px-3 text-xs"
              onClick={addVideo}
              disabled={submitting || videos.length >= 100}
            >
              <Plus className="size-3.5" aria-hidden="true" />
              Add Video
            </Button>
            <span className="text-xs tabular-nums text-slate-600">{nonEmptyRows}/100 videos</span>
          </div>
          {nonEmptyRows > 100 && (
            <p role="status" className="text-xs text-rose-200">
              A batch can contain at most 100 videos.
            </p>
          )}
        </section>

        <div className="grid min-w-0 gap-4 sm:grid-cols-[10rem_minmax(0,1fr)]">
          <div className="min-w-0">
            <FieldLabel label="Audio format" htmlFor="output-format" required>
              <select
                id="output-format"
                value={settings.format}
                onChange={(event) =>
                  onSettingsChange({
                    ...settings,
                    format: event.target.value as Settings["format"],
                  })
                }
                className="h-11 w-full min-w-0 rounded-xl border border-white/10 bg-slate-950/60 px-3.5 text-sm text-slate-100 outline-none transition focus:border-lime-300/40 focus:ring-2 focus:ring-lime-300/10 disabled:opacity-50"
                disabled={submitting || formats.length === 0}
              >
                {formats.length === 0 ? (
                  <option value={settings.format}>No supported formats detected</option>
                ) : (
                  formats.map((format) => (
                    <option key={format} value={format}>
                      {format.toUpperCase()}
                    </option>
                  ))
                )}
              </select>
            </FieldLabel>
          </div>
          <div className="min-w-0">
            <FieldLabel label="Save to" htmlFor="destination-folder" required>
              <div id="destination-folder" className="min-w-0">
                <FolderPicker
                  value={settings.outputDir}
                  onChoose={onChooseFolder}
                  onOpen={onOpenFolder}
                  disabled={submitting}
                />
              </div>
            </FieldLabel>
          </div>
        </div>

        {(validationError || error) && (
          <div
            id="download-error"
            role="alert"
            className="rounded-xl border border-rose-300/15 bg-rose-300/[0.06] px-4 py-3 text-sm leading-5 text-rose-200"
          >
            {validationError || error}
          </div>
        )}

        {!ready && (
          <div className="flex gap-3 rounded-xl border border-amber-200/15 bg-amber-200/[0.05] p-4 text-sm text-amber-100/90">
            <Sparkles className="mt-0.5 size-4 shrink-0 text-amber-200" aria-hidden="true" />
            <div>
              <p className="font-medium">Media tools need setup</p>
              <p className="mt-1 leading-5 text-amber-100/65">
                {capabilities?.message || "Install yt-dlp and ffmpeg to enable downloads."} Add
                details in Advanced.
              </p>
            </div>
          </div>
        )}

        <div className="flex flex-col gap-3 border-t border-white/[0.06] pt-5 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-xs leading-5 text-slate-500">
            {settings.concurrency} downloads can run at once. Settings can be saved in Advanced.
          </p>
          <Button
            type="submit"
            variant="primary"
            disabled={submitting || !ready}
            className="min-h-11 px-5"
          >
            {submitting ? "Adding to queue…" : "Add to queue"}
            {!submitting && <ArrowUpRight className="size-4" aria-hidden="true" />}
          </Button>
        </div>
      </form>
    </Panel>
  );
}
