import { AudioLines, Download, History, Settings2, Sparkles, Waves } from "lucide-react";
import type { PropsWithChildren } from "react";
import { Button } from "../atoms/Button";

export type AppTab = "download" | "history" | "advanced";

interface Props extends PropsWithChildren {
  activeTab: AppTab;
  title: string;
  subtitle: string;
  toolsReady: boolean;
  onTabChange: (tab: AppTab) => void;
  onOpenOutput: () => void;
}

const navigation: { id: AppTab; label: string; icon: typeof Download; number: string }[] = [
  { id: "download", label: "Download", icon: Download, number: "01" },
  { id: "history", label: "History", icon: History, number: "02" },
  { id: "advanced", label: "Advanced", icon: Settings2, number: "03" },
];

export function AppShell({
  activeTab,
  title,
  subtitle,
  toolsReady,
  onTabChange,
  onOpenOutput,
  children,
}: Props) {
  return (
    <div className="app-backdrop min-h-screen text-slate-100">
      <div className="mx-auto flex min-h-screen max-w-[1600px]">
        <aside className="hidden w-[238px] shrink-0 flex-col border-r border-white/[0.055] bg-[#0a0e15]/70 px-4 py-6 lg:flex">
          <div className="flex items-center gap-3 px-2">
            <div className="relative grid size-10 place-items-center rounded-[15px] bg-lime-300 text-slate-950 shadow-[0_8px_30px_rgba(190,242,100,0.16)]">
              <AudioLines className="size-5" strokeWidth={2.3} aria-hidden="true" />
            </div>
            <div>
              <p className="text-[15px] font-bold tracking-tight text-white">yumdown</p>
              <p className="mt-0.5 text-[10px] font-semibold uppercase tracking-[0.19em] text-slate-600">
                Audio utility
              </p>
            </div>
          </div>

          <div className="mt-12 px-3 text-[10px] font-bold uppercase tracking-[0.18em] text-slate-600">
            Workspace
          </div>
          <nav aria-label="Main navigation" className="mt-3 grid gap-1">
            {navigation.map(({ id, label, icon: Icon, number }) => {
              const active = activeTab === id;
              return (
                <Button
                  type="button"
                  key={id}
                  variant="quiet"
                  aria-current={active ? "page" : undefined}
                  onClick={() => onTabChange(id)}
                  className={`group min-h-11 justify-start rounded-xl px-3 ${active ? "bg-white/[0.07] text-white hover:bg-white/[0.09]" : "text-slate-500"}`}
                >
                  <Icon
                    className={`size-4 ${active ? "text-lime-200" : "text-slate-600 group-hover:text-slate-300"}`}
                    aria-hidden="true"
                  />
                  <span className="flex-1 text-left">{label}</span>
                  <span
                    className={`text-[10px] tabular-nums ${active ? "text-slate-500" : "text-slate-700"}`}
                  >
                    {number}
                  </span>
                </Button>
              );
            })}
          </nav>

          <div className="mt-auto">
            <div className="overflow-hidden rounded-2xl border border-white/[0.06] bg-gradient-to-br from-lime-200/[0.08] via-white/[0.025] to-transparent p-4">
              <div className="flex items-center justify-between">
                <span className="grid size-8 place-items-center rounded-xl bg-lime-200/10 text-lime-200">
                  <Waves className="size-4" aria-hidden="true" />
                </span>
                <span
                  className={`size-1.5 rounded-full ${toolsReady ? "bg-lime-300" : "bg-amber-300"}`}
                />
              </div>
              <p className="mt-4 text-xs font-semibold text-slate-200">
                {toolsReady ? "Tools are ready" : "Finish tool setup"}
              </p>
              <p className="mt-1 text-[11px] leading-5 text-slate-500">
                {toolsReady
                  ? "Your audio workflow is ready."
                  : "Add yt-dlp and ffmpeg to enable downloads."}
              </p>
              {!toolsReady && (
                <button
                  type="button"
                  onClick={() => onTabChange("advanced")}
                  className="mt-3 inline-flex items-center gap-1 text-[11px] font-semibold text-lime-200 hover:text-lime-100"
                >
                  Setup tools <Sparkles className="size-3" aria-hidden="true" />
                </button>
              )}
            </div>
            <p className="mt-4 px-2 text-[10px] text-slate-700">Yumdown · Local desktop app</p>
          </div>
        </aside>

        <div className="flex min-w-0 flex-1 flex-col">
          <header className="sticky top-0 z-10 border-b border-white/[0.055] bg-[#0c1018]/85 px-5 py-4 backdrop-blur-xl sm:px-8 lg:px-10">
            <div className="mx-auto flex max-w-[1180px] items-center justify-between gap-4">
              <div className="min-w-0">
                <div className="mb-1 flex items-center gap-2 text-[10px] font-semibold uppercase tracking-[0.17em] text-slate-600 lg:hidden">
                  <AudioLines className="size-3.5 text-lime-200" aria-hidden="true" /> Yumdown
                </div>
                <h1 className="truncate text-lg font-semibold tracking-tight text-white sm:text-xl">
                  {title}
                </h1>
                <p className="mt-1 hidden text-xs text-slate-500 sm:block">{subtitle}</p>
              </div>
              <div className="flex shrink-0 items-center gap-2">
                <span
                  className={`hidden items-center gap-1.5 rounded-full border px-2.5 py-1.5 text-[11px] sm:inline-flex ${toolsReady ? "border-lime-200/15 bg-lime-200/[0.06] text-lime-100/80" : "border-amber-200/15 bg-amber-200/[0.05] text-amber-100/80"}`}
                >
                  <span
                    className={`size-1.5 rounded-full ${toolsReady ? "bg-lime-300" : "bg-amber-300"}`}
                  />
                  {toolsReady ? "Tools ready" : "Setup needed"}
                </span>
                <Button
                  type="button"
                  variant="secondary"
                  className="min-h-9 px-3 text-xs"
                  onClick={onOpenOutput}
                >
                  <Download className="size-3.5" aria-hidden="true" />
                  <span className="hidden sm:inline">Open folder</span>
                </Button>
              </div>
            </div>
          </header>

          <div className="border-b border-white/[0.055] bg-[#0a0e15]/65 px-4 py-2.5 lg:hidden">
            <nav className="mx-auto flex max-w-[1180px] gap-1" aria-label="Main navigation">
              {navigation.map(({ id, label, icon: Icon }) => (
                <button
                  type="button"
                  key={id}
                  onClick={() => onTabChange(id)}
                  aria-current={activeTab === id ? "page" : undefined}
                  className={`flex min-h-9 flex-1 items-center justify-center gap-2 rounded-lg px-3 text-xs font-medium transition ${activeTab === id ? "bg-white/[0.07] text-white" : "text-slate-500 hover:text-slate-200"}`}
                >
                  <Icon className="size-3.5" aria-hidden="true" /> {label}
                </button>
              ))}
            </nav>
          </div>

          <main className="w-full flex-1 px-4 py-6 sm:px-8 sm:py-8 lg:px-10">
            <div className="mx-auto grid max-w-[1180px] gap-5">{children}</div>
          </main>
          <footer className="mx-auto flex w-full max-w-[1180px] items-center justify-between px-5 pb-5 text-[10px] text-slate-700 sm:px-8 lg:px-10">
            <span>Use only content you’re authorized to save.</span>
            <span>Yumdown · Private local history</span>
          </footer>
        </div>
      </div>
    </div>
  );
}
