import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { HistoryPage } from "./features/history/HistoryPage";
import { DownloadPage } from "./features/downloads/DownloadPage";
import { SettingsPage } from "./features/settings/SettingsPage";
import { AppShell, type AppTab } from "./components/templates/AppShell";
import { Panel } from "./components/atoms/Panel";
import { backend, onJobEvent } from "./lib/backend";
import type {
  AggregateProgress,
  Capabilities,
  DownloadRequest,
  HistoryEntry,
  Job,
  JobEvent,
  Settings,
} from "./types/domain";

const initialSettings: Settings = {
  outputDir: "",
  format: "mp3",
  concurrency: 2,
  audioQuality: "192K",
};

const initialAggregate: AggregateProgress = {
  totalItems: 0,
  finishedItems: 0,
  percent: 0,
  indeterminate: false,
};

const pageContent: Record<AppTab, { title: string; subtitle: string }> = {
  download: {
    title: "Download audio",
    subtitle: "Turn video links into a clean, organized audio library.",
  },
  history: {
    title: "Download history",
    subtitle: "A private record of the downloads on this device.",
  },
  advanced: {
    title: "Advanced settings",
    subtitle: "Configure defaults and inspect your media-tool setup.",
  },
};

function upsertJobs(previous: Job[], incoming: Job[]) {
  const jobs = new Map(previous.map((job) => [job.id, job]));
  for (const job of incoming) {
    const existing = jobs.get(job.id);
    if (!existing || job.updatedAt >= existing.updatedAt) jobs.set(job.id, job);
  }
  return [...jobs.values()].sort((a, b) => a.createdAt.localeCompare(b.createdAt));
}

function aggregateFromJobs(jobs: Job[]): AggregateProgress {
  if (!jobs.length) return initialAggregate;
  let units = 0;
  let finishedItems = 0;
  let indeterminate = false;
  for (const job of jobs) {
    if (["completed", "failed", "cancelled"].includes(job.status)) {
      units += 1;
      finishedItems += 1;
    } else if (job.status === "queued") {
      continue;
    } else if (job.progress.percent === undefined) {
      indeterminate = true;
    } else {
      units += Math.max(0, Math.min(100, job.progress.percent)) / 100;
    }
  }
  return {
    totalItems: jobs.length,
    finishedItems,
    percent: indeterminate ? undefined : (units / jobs.length) * 100,
    indeterminate,
  };
}

export default function App() {
  const [activeTab, setActiveTab] = useState<AppTab>("download");
  const [settings, setSettings] = useState<Settings>(initialSettings);
  const [capabilities, setCapabilities] = useState<Capabilities | null>(null);
  const [jobs, setJobs] = useState<Job[]>([]);
  const jobsRef = useRef<Job[]>([]);
  const [aggregate, setAggregate] = useState<AggregateProgress>(initialAggregate);
  const [history, setHistory] = useState<HistoryEntry[]>([]);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [historyError, setHistoryError] = useState("");
  const [initializing, setInitializing] = useState(true);
  const [startupError, setStartupError] = useState("");
  const [formError, setFormError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [saving, setSaving] = useState(false);
  const [settingsSaved, setSettingsSaved] = useState(false);
  const [settingsError, setSettingsError] = useState("");
  const [toast, setToast] = useState("");

  const page = pageContent[activeTab];
  const toolsReady = Boolean(capabilities?.toolsAvailable && capabilities.formats.length > 0);
  const handleJobEvent = useCallback((event: JobEvent) => {
    const updated = upsertJobs(jobsRef.current, [event.job]);
    jobsRef.current = updated;
    setJobs(updated);
    setAggregate(event.aggregate);
  }, []);

  useEffect(() => {
    let mounted = true;
    const unsubscribe = onJobEvent((event) => {
      if (mounted) handleJobEvent(event);
    });

    async function initialize() {
      try {
        const startupIssue = await backend.startupError();
        if (startupIssue) throw new Error(startupIssue);
        const [loadedSettings, loadedCapabilities, loadedJobs] = await Promise.all([
          backend.getSettings(),
          backend.getCapabilities(),
          backend.getJobs(),
        ]);
        if (!mounted) return;
        const normalizedSettings = { ...loadedSettings };
        if (
          loadedCapabilities.formats.length > 0 &&
          !loadedCapabilities.formats.includes(normalizedSettings.format)
        ) {
          normalizedSettings.format = loadedCapabilities.formats[0];
        }
        setSettings(normalizedSettings);
        setCapabilities(loadedCapabilities);
        const currentJobs = upsertJobs(loadedJobs, jobsRef.current);
        jobsRef.current = currentJobs;
        setJobs(currentJobs);
        setAggregate(aggregateFromJobs(currentJobs));
      } catch (error) {
        if (mounted)
          setStartupError(error instanceof Error ? error.message : "Yumdown could not initialize.");
      } finally {
        if (mounted) setInitializing(false);
      }
    }

    void initialize();
    return () => {
      mounted = false;
      unsubscribe();
    };
  }, [handleJobEvent]);

  const loadHistory = useCallback(async (search: string, status: string) => {
    setHistoryLoading(true);
    setHistoryError("");
    try {
      const entries = await backend.getHistory(search, status);
      setHistory(entries);
    } catch (error) {
      setHistoryError(error instanceof Error ? error.message : "Could not load download history.");
    } finally {
      setHistoryLoading(false);
    }
  }, []);

  const handleHistorySearch = useCallback(
    (search: string, status: string) => {
      void loadHistory(search, status);
    },
    [loadHistory],
  );

  const chooseOutputFolder = useCallback(async () => {
    try {
      const folder = await backend.selectOutputFolder();
      if (folder) {
        setSettings((current) => ({ ...current, outputDir: folder }));
        setSettingsSaved(false);
      }
    } catch (error) {
      setToast(error instanceof Error ? error.message : "Could not select a folder.");
    }
  }, []);

  const openFolder = useCallback(async (path: string) => {
    if (!path) return;
    try {
      await backend.openFolder(path);
      setToast("");
    } catch (error) {
      setToast(error instanceof Error ? error.message : "Could not open that folder.");
    }
  }, []);

  const openOutputFolder = useCallback(async () => {
    try {
      await backend.openOutputFolder();
      setToast("");
    } catch (error) {
      setToast(error instanceof Error ? error.message : "Could not open the output folder.");
    }
  }, []);

  const startDownloads = useCallback(async (request: DownloadRequest) => {
    setSubmitting(true);
    setFormError("");
    try {
      const created = await backend.startDownloads(request);
      const updated = upsertJobs(jobsRef.current, created);
      jobsRef.current = updated;
      setJobs(updated);
      setAggregate(aggregateFromJobs(updated));
      setToast(
        `${created.length} ${created.length === 1 ? "download" : "downloads"} added to the queue.`,
      );
    } catch (error) {
      setFormError(
        error instanceof Error ? error.message : "Could not add downloads to the queue.",
      );
    } finally {
      setSubmitting(false);
    }
  }, []);

  const cancelDownload = useCallback(async (id: string) => {
    try {
      await backend.cancelDownload(id);
    } catch (error) {
      setToast(error instanceof Error ? error.message : "Could not cancel that download.");
    }
  }, []);

  const changeSettings = useCallback((next: Settings) => {
    setSettings(next);
    setSettingsSaved(false);
    setSettingsError("");
  }, []);

  const saveSettings = useCallback(async () => {
    setSaving(true);
    setSettingsError("");
    setSettingsSaved(false);
    try {
      const saved = await backend.saveSettings(settings);
      setSettings(saved);
      setSettingsSaved(true);
    } catch (error) {
      setSettingsError(error instanceof Error ? error.message : "Could not save settings.");
    } finally {
      setSaving(false);
    }
  }, [settings]);

  const refreshTools = useCallback(async () => {
    try {
      const updated = await backend.getCapabilities();
      setCapabilities(updated);
      setSettings((current) => ({
        ...current,
        format: updated.formats.includes(current.format)
          ? current.format
          : (updated.formats[0] ?? current.format),
      }));
    } catch (error) {
      setToast(error instanceof Error ? error.message : "Could not check the media tools.");
    }
  }, []);

  const clearHistory = useCallback(async () => {
    if (!window.confirm("Clear all download history? Saved audio files will not be deleted."))
      return;
    try {
      await backend.clearHistory();
      setHistory([]);
      setHistoryError("");
    } catch (error) {
      setHistoryError(error instanceof Error ? error.message : "Could not clear download history.");
    }
  }, []);

  const body = useMemo(() => {
    if (initializing) {
      return (
        <Panel className="grid min-h-64 place-items-center text-sm text-slate-500">
          Preparing your workspace…
        </Panel>
      );
    }
    if (startupError) {
      return (
        <Panel className="grid min-h-64 place-items-center p-8 text-center">
          <div className="max-w-lg">
            <h2 className="text-base font-semibold text-white">Yumdown could not start</h2>
            <p className="mt-2 text-sm leading-6 text-rose-200">{startupError}</p>
            <p className="mt-3 text-xs leading-5 text-slate-500">
              Restart the application after resolving this local setup issue.
            </p>
          </div>
        </Panel>
      );
    }
    if (activeTab === "download") {
      return (
        <>
          {toast && (
            <div
              role="status"
              className="rounded-xl border border-lime-200/15 bg-lime-200/5 px-4 py-3 text-sm text-lime-100"
            >
              {toast}
            </div>
          )}
          <DownloadPage
            capabilities={capabilities}
            settings={settings}
            jobs={jobs}
            aggregate={aggregate}
            submitting={submitting}
            error={formError}
            onSettingsChange={changeSettings}
            onChooseFolder={() => void chooseOutputFolder()}
            onOpenFolder={(path) => void openFolder(path)}
            onStart={(request) => void startDownloads(request)}
            onCancel={(id) => void cancelDownload(id)}
          />
        </>
      );
    }
    if (activeTab === "history") {
      return (
        <>
          {toast && (
            <div
              role="status"
              className="rounded-xl border border-amber-200/15 bg-amber-200/5 px-4 py-3 text-sm text-amber-100"
            >
              {toast}
            </div>
          )}
          <HistoryPage
            entries={history}
            loading={historyLoading}
            error={historyError}
            onSearch={handleHistorySearch}
            onClear={() => void clearHistory()}
            onOpenFolder={(path) => void openFolder(path)}
          />
        </>
      );
    }
    return (
      <>
        {toast && (
          <div
            role="status"
            className="rounded-xl border border-amber-200/15 bg-amber-200/5 px-4 py-3 text-sm text-amber-100"
          >
            {toast}
          </div>
        )}
        <SettingsPage
          settings={settings}
          capabilities={capabilities}
          saving={saving}
          saved={settingsSaved}
          error={settingsError}
          onSettingsChange={changeSettings}
          onChooseFolder={() => void chooseOutputFolder()}
          onOpenFolder={() => void openOutputFolder()}
          onSave={() => void saveSettings()}
          onRefreshTools={() => void refreshTools()}
        />
      </>
    );
  }, [
    activeTab,
    aggregate,
    cancelDownload,
    capabilities,
    changeSettings,
    chooseOutputFolder,
    clearHistory,
    formError,
    history,
    historyError,
    historyLoading,
    initializing,
    jobs,
    handleHistorySearch,
    openFolder,
    openOutputFolder,
    refreshTools,
    saveSettings,
    settings,
    settingsError,
    settingsSaved,
    startupError,
    startDownloads,
    submitting,
    toast,
    saving,
  ]);

  return (
    <AppShell
      activeTab={activeTab}
      title={page.title}
      subtitle={page.subtitle}
      toolsReady={toolsReady}
      onTabChange={setActiveTab}
      onOpenOutput={() => void openOutputFolder()}
    >
      {body}
    </AppShell>
  );
}
