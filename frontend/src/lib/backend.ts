import * as App from "../../wailsjs/go/main/App";
import { EventsOn } from "../../wailsjs/runtime/runtime";
import type { downloader } from "../../wailsjs/go/models";
import type {
  Capabilities,
  DownloadRequest,
  HistoryEntry,
  Job,
  JobEvent,
  Settings,
} from "../types/domain";

export const backend = {
  startDownloads: async (request: DownloadRequest): Promise<Job[]> =>
    (await App.StartDownloads(
      request as unknown as downloader.DownloadRequest,
    )) as unknown as Job[],
  getJobs: async (): Promise<Job[]> => (await App.GetJobs()) as unknown as Job[],
  cancelDownload: (id: string) => App.CancelDownload(id),
  getHistory: async (search = "", status = "all"): Promise<HistoryEntry[]> =>
    (await App.GetHistory(search, status)) as unknown as HistoryEntry[],
  clearHistory: () => App.ClearHistory(),
  getSettings: async (): Promise<Settings> => (await App.GetSettings()) as unknown as Settings,
  saveSettings: async (settings: Settings): Promise<Settings> =>
    (await App.SaveSettings(settings)) as unknown as Settings,
  getCapabilities: async (): Promise<Capabilities> =>
    (await App.GetCapabilities()) as unknown as Capabilities,
  selectOutputFolder: () => App.SelectOutputFolder(),
  openFolder: (path: string) => App.OpenFolder(path),
  openOutputFolder: () => App.OpenOutputFolder(),
  startupError: () => App.StartupError(),
};

export function onJobEvent(callback: (event: JobEvent) => void): () => void {
  if (!("runtime" in window)) return () => undefined;
  return EventsOn("download:job", (event) => callback(event as JobEvent));
}
