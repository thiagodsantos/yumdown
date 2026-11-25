import { CircleAlert, FolderOpen, Search, Trash2 } from "lucide-react";
import type { HistoryEntry, JobStatus } from "../../types/domain";
import { formatBytes, formatDate } from "../../lib/format";
import { safeDisplayURL } from "../../features/downloads/url-utils";
import { Button } from "../atoms/Button";
import { Panel } from "../atoms/Panel";
import { StatusBadge } from "../atoms/StatusBadge";

interface Props {
  entries: HistoryEntry[];
  search: string;
  status: string;
  loading: boolean;
  error: string;
  onSearch: (value: string) => void;
  onStatus: (value: string) => void;
  onClear: () => void;
  onOpenFolder: (path: string) => void;
}

export function HistoryTable({
  entries,
  search,
  status,
  loading,
  error,
  onSearch,
  onStatus,
  onClear,
  onOpenFolder,
}: Props) {
  return (
    <Panel className="overflow-hidden">
      <div className="flex flex-col gap-4 border-b border-white/[0.06] p-5 sm:flex-row sm:items-center sm:justify-between sm:px-6">
        <div className="relative min-w-0 flex-1 sm:max-w-md">
          <Search
            className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-slate-500"
            aria-hidden="true"
          />
          <input
            type="search"
            value={search}
            onChange={(event) => onSearch(event.target.value)}
            placeholder="Search URL, file or error…"
            aria-label="Search download history"
            className="h-10 w-full rounded-xl border border-white/10 bg-slate-950/55 pl-9 pr-3 text-sm text-slate-100 outline-none placeholder:text-slate-600 focus:border-lime-300/40 focus:ring-2 focus:ring-lime-300/10"
          />
        </div>
        <div className="flex items-center gap-2">
          <label htmlFor="history-status" className="sr-only">
            Filter by status
          </label>
          <select
            id="history-status"
            value={status}
            onChange={(event) => onStatus(event.target.value)}
            className="h-10 rounded-xl border border-white/10 bg-slate-950/55 px-3 text-sm text-slate-300 outline-none focus:border-lime-300/40"
          >
            <option value="all">All statuses</option>
            <option value="completed">Completed</option>
            <option value="failed">Failed</option>
            <option value="cancelled">Cancelled</option>
          </select>
          <Button
            type="button"
            variant="quiet"
            className="min-h-10 px-3 text-xs text-rose-200/80"
            onClick={onClear}
            disabled={entries.length === 0 || loading}
          >
            <Trash2 className="size-3.5" aria-hidden="true" />
            Clear
          </Button>
        </div>
      </div>

      {error && (
        <p className="m-5 rounded-xl bg-rose-300/[0.06] p-3 text-sm text-rose-200" role="alert">
          {error}
        </p>
      )}
      {loading ? (
        <div className="grid min-h-48 place-items-center text-sm text-slate-500">
          Loading history…
        </div>
      ) : entries.length === 0 ? (
        <div className="grid min-h-56 place-items-center px-6 py-12 text-center">
          <div>
            <div className="mx-auto grid size-12 place-items-center rounded-2xl border border-white/[0.08] bg-white/[0.03] text-slate-500">
              <FolderOpen className="size-5" aria-hidden="true" />
            </div>
            <h3 className="mt-4 text-sm font-medium text-slate-300">No matching downloads</h3>
            <p className="mt-1 text-sm text-slate-500">
              Completed and failed jobs will appear here.
            </p>
          </div>
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full min-w-[680px] border-collapse text-left">
            <thead>
              <tr className="text-[11px] uppercase tracking-[0.14em] text-slate-600">
                <th scope="col" className="px-6 py-3 font-medium">
                  Source
                </th>
                <th scope="col" className="px-4 py-3 font-medium">
                  Status
                </th>
                <th scope="col" className="px-4 py-3 font-medium">
                  Saved
                </th>
                <th scope="col" className="px-4 py-3 font-medium">
                  Size
                </th>
                <th scope="col" className="px-5 py-3 text-right font-medium">
                  Folder
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/[0.05]">
              {entries.map((entry) => (
                <tr key={entry.id} className="group transition hover:bg-white/[0.02]">
                  <td className="max-w-[360px] px-6 py-4">
                    <div
                      className="truncate text-sm font-medium text-slate-200"
                      title={safeDisplayURL(entry.url)}
                    >
                      {safeDisplayURL(entry.url)}
                    </div>
                    <div
                      className="mt-1 truncate text-xs text-slate-500"
                      title={entry.outputPath || entry.outputDir}
                    >
                      {entry.outputPath || `${entry.outputDir} · ${entry.format.toUpperCase()}`}
                    </div>
                    {entry.error && (
                      <p className="mt-1 truncate text-xs text-rose-200/80" title={entry.error}>
                        {entry.error}
                      </p>
                    )}
                  </td>
                  <td className="px-4 py-4">
                    <StatusBadge status={entry.status as JobStatus} />
                  </td>
                  <td className="whitespace-nowrap px-4 py-4 text-xs text-slate-400">
                    {formatDate(entry.finishedAt)}
                  </td>
                  <td className="whitespace-nowrap px-4 py-4 text-xs tabular-nums text-slate-400">
                    {entry.status === "completed" ? (
                      formatBytes(entry.outputSize)
                    ) : entry.status === "cancelled" ? (
                      "—"
                    ) : (
                      <span className="inline-flex items-center gap-1.5">
                        <CircleAlert className="size-3.5 text-rose-300" aria-hidden="true" />
                        Failed
                      </span>
                    )}
                  </td>
                  <td className="px-5 py-4 text-right">
                    <Button
                      type="button"
                      variant="quiet"
                      className="min-h-8 px-2.5 text-xs"
                      onClick={() => onOpenFolder(entry.outputDir)}
                      aria-label={`Open folder for ${safeDisplayURL(entry.url)}`}
                    >
                      <FolderOpen className="size-3.5" aria-hidden="true" />
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </Panel>
  );
}
