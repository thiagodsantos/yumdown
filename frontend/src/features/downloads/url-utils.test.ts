import { describe, expect, it } from "vitest";
import { normalizeVideoRows, safeDisplayURL } from "./url-utils";

describe("normalizeVideoRows", () => {
  it("keeps each URL paired with its own trimmed output name", () => {
    expect(
      normalizeVideoRows([
        { url: " https://example.com/a ", filename: " Track A " },
        { url: "", filename: "" },
        { url: "https://example.org/b", filename: "" },
      ]),
    ).toEqual([
      { url: "https://example.com/a", filename: "Track A" },
      { url: "https://example.org/b", filename: "" },
    ]);
  });

  it("rejects non-web URLs and rows with a name but no URL", () => {
    expect(() => normalizeVideoRows([{ url: "file:///tmp/audio.mp3", filename: "" }])).toThrow(
      "must use http or https",
    );
    expect(() => normalizeVideoRows([{ url: "", filename: "Track A" }])).toThrow(
      "Video row 1 needs a URL",
    );
    expect(() => normalizeVideoRows([{ url: "not a URL", filename: "" }])).toThrow("is not valid");
  });

  it("requires at least one URL and limits batches to 100 rows", () => {
    expect(() => normalizeVideoRows([{ url: "", filename: "" }])).toThrow(
      "Add at least one video URL",
    );
    const videos = Array.from({ length: 101 }, (_, index) => ({
      url: `https://example.com/${index}`,
      filename: "",
    }));
    expect(() => normalizeVideoRows(videos)).toThrow("at most 100 videos");
  });

  it("does not reveal URL query parameters in display text", () => {
    expect(safeDisplayURL("https://video.example/watch?id=secret")).toBe("video.example/watch");
  });
});
