import { FolderOpen } from "lucide-react";
import { Button } from "../atoms/Button";

interface Props {
  value: string;
  onChoose: () => void;
  onOpen?: () => void;
  disabled?: boolean;
}

export function FolderPicker({ value, onChoose, onOpen, disabled = false }: Props) {
  return (
    <div className="flex h-11 w-full min-w-0 max-w-full items-center gap-2 rounded-xl border border-white/10 bg-slate-950/50 py-1 pl-3 pr-2">
      <FolderOpen className="size-4 shrink-0 text-slate-500" aria-hidden="true" />
      <span
        className="min-w-0 flex-1 truncate text-sm text-slate-300"
        title={value || "No folder selected"}
      >
        {value || "Choose an output folder"}
      </span>
      {onOpen && value && (
        <Button
          type="button"
          variant="quiet"
          className="min-h-8 shrink-0 whitespace-nowrap px-2.5 text-xs"
          onClick={onOpen}
          disabled={disabled}
        >
          Open
        </Button>
      )}
      <Button
        type="button"
        variant="secondary"
        className="min-h-8 shrink-0 whitespace-nowrap px-3 text-xs"
        onClick={onChoose}
        disabled={disabled}
      >
        Browse
      </Button>
    </div>
  );
}
