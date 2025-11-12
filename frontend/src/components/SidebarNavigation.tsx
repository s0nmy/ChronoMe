import React from "react";
import {
  Clock,
  FileText,
  FolderOpen,
  Settings,
} from "lucide-react";

export type TabType =
  | "timer"
  | "entries"
  | "projects"
  | "settings";

interface SidebarNavigationProps {
  activeTab: TabType;
  onTabChange: (tab: TabType) => void;
}

const tabs = [
  {
    id: "timer" as TabType,
    label: "時間記録",
    icon: Clock,
    description: "リアルタイムの計測",
  },
  {
    id: "entries" as TabType,
    label: "エントリ",
    icon: FileText,
    description: "過去の記録",
  },
  {
    id: "projects" as TabType,
    label: "プロジェクト",
    icon: FolderOpen,
    description: "案件の管理",
  },
  {
    id: "settings" as TabType,
    label: "設定",
    icon: Settings,
    description: "アプリ設定",
  },
];

export function SidebarNavigation({
  activeTab,
  onTabChange,
}: SidebarNavigationProps) {
  return (
    <aside className="min-h-screen w-64 lg:w-72 bg-card/60 border-r border-border/80 backdrop-blur supports-[backdrop-filter]:bg-card/70">
      <div className="px-4 py-4 border-b border-border/70">
        <p className="text-sm font-semibold tracking-wide text-primary">
          Timecard
        </p>
        <p className="text-xs text-muted-foreground">
          あなたの作業を可視化
        </p>
      </div>

      <nav className="flex flex-col gap-1 p-2 md:p-3">
        {tabs.map((tab) => {
          const Icon = tab.icon;
          const isActive = activeTab === tab.id;

          return (
            <button
              key={tab.id}
              type="button"
              onClick={() => onTabChange(tab.id)}
              className={`flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors ${
                isActive
                  ? "bg-primary/10 text-primary"
                  : "text-muted-foreground hover:bg-accent hover:text-foreground"
              }`}
            >
              <Icon
                className={`h-5 w-5 ${
                  isActive ? "text-primary" : ""
                }`}
              />
              <div className="flex flex-col items-start">
                <span className="font-medium">{tab.label}</span>
                <span className="text-[11px] text-muted-foreground">
                  {tab.description}
                </span>
              </div>
            </button>
          );
        })}
      </nav>
    </aside>
  );
}
