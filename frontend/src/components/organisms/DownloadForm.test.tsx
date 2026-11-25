import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { DownloadForm } from "./DownloadForm";
import type { Capabilities, Settings } from "../../types/domain";

const settings: Settings = {
  outputDir: "/home/test/Downloads/Yumdown",
  format: "mp3",
  concurrency: 2,
  audioQuality: "192K",
};

const capabilities: Capabilities = {
  toolsAvailable: true,
  ffprobeFound: false,
  formats: ["mp3", "flac"],
};

function renderForm(onSubmit = vi.fn()) {
  render(
    <DownloadForm
      capabilities={capabilities}
      settings={settings}
      onSettingsChange={vi.fn()}
      onChooseFolder={vi.fn()}
      onOpenFolder={vi.fn()}
      onSubmit={onSubmit}
      submitting={false}
      error=""
    />,
  );
  return onSubmit;
}

const videoURLInputs = () => screen.getAllByLabelText(/^Video URL \d+$/);
const outputNameInputs = () => screen.getAllByLabelText(/^Output name \d+$/);

describe("DownloadForm", () => {
  it("opens with one video row", () => {
    renderForm();
    expect(videoURLInputs()).toHaveLength(1);
    expect(outputNameInputs()).toHaveLength(1);
  });

  it("adds rows and Clear resets the list to one empty row", () => {
    renderForm();
    fireEvent.click(screen.getByRole("button", { name: "Add Video" }));
    expect(videoURLInputs()).toHaveLength(2);
    expect(outputNameInputs()).toHaveLength(2);

    fireEvent.change(videoURLInputs()[0], {
      target: { value: "https://video.example/watch?id=1" },
    });
    fireEvent.change(outputNameInputs()[0], { target: { value: "first" } });
    fireEvent.click(screen.getByRole("button", { name: "Clear all video rows" }));

    expect(videoURLInputs()).toHaveLength(1);
    expect(outputNameInputs()).toHaveLength(1);
    expect(videoURLInputs()[0]).toHaveValue("");
    expect(outputNameInputs()[0]).toHaveValue("");
  });

  it("submits each row's URL and output name together", () => {
    const onSubmit = renderForm();
    fireEvent.change(videoURLInputs()[0], {
      target: { value: "https://video.example/watch?id=1" },
    });
    fireEvent.change(outputNameInputs()[0], { target: { value: "track one" } });
    fireEvent.click(screen.getByRole("button", { name: "Add Video" }));
    fireEvent.change(videoURLInputs()[1], {
      target: { value: "https://video.example/watch?id=2" },
    });
    fireEvent.change(outputNameInputs()[1], { target: { value: "track two" } });
    fireEvent.click(screen.getByRole("button", { name: /add to queue/i }));

    expect(onSubmit).toHaveBeenCalledWith({
      videos: [
        { url: "https://video.example/watch?id=1", filename: "track one" },
        { url: "https://video.example/watch?id=2", filename: "track two" },
      ],
      outputDir: settings.outputDir,
      format: settings.format,
      audioQuality: settings.audioQuality,
    });
  });

  it("shows per-row URL validation before submission", () => {
    const onSubmit = renderForm();
    fireEvent.change(screen.getByLabelText(/^Video URL 1$/), {
      target: { value: "file:///tmp/audio.mp3" },
    });
    expect(screen.getByRole("status")).toHaveTextContent(
      "URL 1 must use http or https without embedded credentials.",
    );
    fireEvent.click(screen.getByRole("button", { name: /add to queue/i }));
    expect(onSubmit).not.toHaveBeenCalled();
  });
});
