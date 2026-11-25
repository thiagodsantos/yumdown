import type { VideoRequest } from "../../types/domain";

export function normalizeVideoRows(rows: VideoRequest[]): VideoRequest[] {
  const activeRows = rows
    .map((video, index) => ({
      rowNumber: index + 1,
      url: video.url.trim(),
      filename: video.filename.trim(),
    }))
    .filter((video) => video.url || video.filename);

  if (activeRows.length === 0) throw new Error("Add at least one video URL.");
  if (activeRows.length > 100) throw new Error("A batch can contain at most 100 videos.");

  return activeRows.map((video) => {
    if (!video.url) throw new Error(`Video row ${video.rowNumber} needs a URL.`);
    const issue = validateVideoURL(video.url, video.rowNumber);
    if (issue) throw new Error(issue);
    return { url: video.url, filename: video.filename };
  });
}

export function validateVideoURL(value: string, index: number): string | null {
  let parsed: URL;
  try {
    parsed = new URL(value);
  } catch {
    return `URL ${index} is not valid.`;
  }
  if (
    (parsed.protocol !== "http:" && parsed.protocol !== "https:") ||
    !parsed.hostname ||
    parsed.username ||
    parsed.password
  ) {
    return `URL ${index} must use http or https without embedded credentials.`;
  }
  return null;
}

export function safeDisplayURL(value: string): string {
  try {
    const parsed = new URL(value);
    return `${parsed.hostname}${parsed.pathname === "/" ? "" : parsed.pathname}`;
  } catch {
    return value;
  }
}
