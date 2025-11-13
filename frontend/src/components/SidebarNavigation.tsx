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
  },
  {
    id: "entries" as TabType,
    label: "作業履歴",
    icon: FileText,
  },
  {
    id: "projects" as TabType,
    label: "プロジェクト",
    icon: FolderOpen,
  },
  {
    id: "settings" as TabType,
    label: "設定",
    icon: Settings,
  },
];

export function SidebarNavigation({
  activeTab,
  onTabChange,
}: SidebarNavigationProps) {
  return (
    <header className="w-full bg-card/60 border-b border-border/80 backdrop-blur supports-[backdrop-filter]:bg-card/70">
      <div className="max-w-6xl mx-auto px-4 py-4 md:px-6 lg:px-8 flex items-center gap-4">
        <p className="text-2xl font-semibold tracking-wide text-primary whitespace-nowrap">
          ChronoMe
        </p>

        <nav className="flex items-center gap-2 ml-auto overflow-x-auto whitespace-nowrap">
          {tabs.map((tab) => {
            const Icon = tab.icon;
            const isActive = activeTab === tab.id;

            return (
              <button
                key={tab.id}
                type="button"
                onClick={() => onTabChange(tab.id)}
                className={`flex items-center gap-2 rounded-full border px-4 py-2 text-sm transition-colors ${
                  isActive
                    ? "bg-primary/10 text-primary border-primary/40"
                    : "border-transparent text-muted-foreground hover:bg-accent/60 hover:text-foreground"
                }`}
              >
                <Icon
                  className={`h-5 w-5 ${
                    isActive ? "text-primary" : ""
                  }`}
                />
                <span className="font-medium whitespace-nowrap">
                  {tab.label}
                </span>
              </button>
            );
          })}
        </nav>
      </div>
    </header>
  );
}
