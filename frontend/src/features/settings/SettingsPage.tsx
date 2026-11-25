import { SettingsForm } from "../../components/organisms/SettingsForm";
import type { Capabilities, Settings } from "../../types/domain";

interface Props {
  settings: Settings;
  capabilities: Capabilities | null;
  saving: boolean;
  saved: boolean;
  error: string;
  onSettingsChange: (settings: Settings) => void;
  onChooseFolder: () => void;
  onOpenFolder: () => void;
  onSave: () => void;
  onRefreshTools: () => void;
}

export function SettingsPage(props: Props) {
  return (
    <>
      <div className="max-w-2xl">
        <p className="text-sm leading-6 text-slate-400">
          Tune default output behavior and check which media tools this installation can use.
          Advanced options are intentionally limited to supported, validated choices.
        </p>
      </div>
      <SettingsForm {...props} />
    </>
  );
}
