export type AudioFormat = "mp3" | "m4a" | "opus" | "wav" | "flac";
export type JobStatus =
  "queued" | "downloading" | "converting" | "completed" | "failed" | "cancelled";

export interface VideoRequest {
  url: string;
  filename: string;
}

export interface DownloadRequest {
  videos: VideoRequest[];
  outputDir: string;
  format: AudioFormat;
  audioQuality: string;
}

export interface Progress {
  percent?: number;
  downloadedBytes?: number;
  totalBytes?: number;
  speedBytesPerSecond?: number;
  etaSeconds?: number;
}

export interface Job {
  id: string;
  url: string;
  outputDir: string;
  outputName: string;
  format: AudioFormat;
  audioQuality: string;
  status: JobStatus;
  progress: Progress;
  outputPath?: string;
  outputSize?: number;
  error?: string;
  createdAt: string;
  updatedAt: string;
}

export interface HistoryEntry {
  id: string;
  url: string;
  outputDir: string;
  outputPath?: string;
  format: AudioFormat;
  status: JobStatus;
  outputSize?: number;
  error?: string;
  finishedAt: string;
}

export interface Settings {
  outputDir: string;
  format: AudioFormat;
  concurrency: number;
  audioQuality: string;
}

export interface AggregateProgress {
  totalItems: number;
  finishedItems: number;
  percent?: number;
  indeterminate: boolean;
}

export interface JobEvent {
  job: Job;
  aggregate: AggregateProgress;
}

export interface Capabilities {
  toolsAvailable: boolean;
  ytDlpVersion?: string;
  ffmpegVersion?: string;
  ffprobeFound: boolean;
  toolsDirectory?: string;
  formats: AudioFormat[];
  message?: string;
}
