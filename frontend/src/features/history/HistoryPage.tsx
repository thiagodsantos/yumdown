import { useEffect, useState } from "react";
import { HistoryTable } from "../../components/organisms/HistoryTable";
import { Panel } from "../../components/atoms/Panel";
import type { HistoryEntry } from "../../types/domain";

interface Props {
  entries: HistoryEntry[];
  loading: boolean;
  error: string;
  onSearch: (search: string, status: string) => void;
  onClear: () => void;
  onOpenFolder: (path: string) => void;
}

export function HistoryPage({ entries, loading, error, onSearch, onClear, onOpenFolder }: Props) {
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("all");

  useEffect(() => {
    const timer = window.setTimeout(() => onSearch(search, status), 180);
    return () => window.clearTimeout(timer);
  }, [onSearch, search, status]);

  return (
    <>
      <Panel className="flex items-center justify-between gap-4 px-5 py-4 sm:px-6">
        <div>
          <p className="text-sm font-semibold text-slate-100">Your local archive</p>
          <p className="mt-1 text-xs text-slate-500">
            History stays on this device. Clearing it never removes audio files.
          </p>
        </div>
        <span className="rounded-lg border border-white/[0.07] bg-white/[0.025] px-3 py-2 text-xs tabular-nums text-slate-400">
          {entries.length} shown
        </span>
      </Panel>
      <HistoryTable
        entries={entries}
        search={search}
        status={status}
        loading={loading}
        error={error}
        onSearch={setSearch}
        onStatus={setStatus}
        onClear={onClear}
        onOpenFolder={onOpenFolder}
      />
    </>
  );
}
