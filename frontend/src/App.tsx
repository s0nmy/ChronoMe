import { AuthCallback } from "./components/AuthCallback";
import { LoginPage } from "./components/LoginPage";
import { SidebarNavigation } from "./components/SidebarNavigation";
import { TimerScreen } from "./components/TimerScreen";
import { EntriesScreen } from "./components/EntriesScreen";
import { ProjectsScreen } from "./components/ProjectsScreen";
import { SettingsScreen } from "./components/SettingsScreen";
import { AuthProvider } from "./contexts/AuthContext";
import { useChronome } from "./features/app/useChronome";

function AppContent() {
  const {
    user,
    entries,
    projects,
    activeEntries,
    activeTab,
    setActiveTab,
    initializing,
    handleLogin,
    handleSignup,
    handleOAuthLogin,
    handleLogout,
    handleStartEntryTimer,
    handlePauseEntryTimer,
    handleResumeEntryTimer,
    handleUpdateActiveEntry,
    getCurrentElapsedTime,
    handleStopEntryTimer,
    handleCreateManualEntry,
    handleCreateProject,
    handleUpdateProject,
    handleDeleteProject,
    handleExportData,
    handleUpdateEntry,
    handleDeleteEntry,
    handleDeleteAllData,
  } = useChronome();

  if (initializing) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <p className="text-sm text-muted-foreground">ChronoMe を読み込み中です…</p>
      </div>
    );
  }

  if (!user) {
    return (
      <LoginPage onLogin={handleLogin} onSignup={handleSignup} onOAuthLogin={handleOAuthLogin} />
    );
  }

  const renderScreen = () => {
    switch (activeTab) {
      case "timer":
        return (
          <TimerScreen
            projects={projects}
            activeEntries={activeEntries}
            entries={entries}
            onStartEntry={handleStartEntryTimer}
            onPauseEntry={handlePauseEntryTimer}
            onResumeEntry={handleResumeEntryTimer}
            onStopEntry={handleStopEntryTimer}
            onUpdateActiveEntry={handleUpdateActiveEntry}
            onCreateManualEntry={handleCreateManualEntry}
            onCreateProject={handleCreateProject}
            getCurrentElapsedTime={getCurrentElapsedTime}
            onUpdateEntry={handleUpdateEntry}
            onDeleteEntry={handleDeleteEntry}
          />
        );
      case "entries":
        return (
          <EntriesScreen
            entries={entries}
            projects={projects}
            onUpdateEntry={handleUpdateEntry}
            onDeleteEntry={handleDeleteEntry}
            onCreateProject={() => setActiveTab("projects")}
          />
        );
      case "projects":
        return (
          <ProjectsScreen
            projects={projects}
            entries={entries}
            onCreateProject={handleCreateProject}
            onUpdateProject={handleUpdateProject}
            onDeleteProject={handleDeleteProject}
          />
        );
      case "settings":
        return (
          <SettingsScreen
            user={user}
            entries={entries}
            onLogout={handleLogout}
            onExportData={handleExportData}
            onDeleteAllData={handleDeleteAllData}
          />
        );
      default:
        return null;
    }
  };

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <SidebarNavigation activeTab={activeTab} onTabChange={setActiveTab} />
      <main className="flex-1 overflow-y-auto">
        <div className="max-w-6xl mx-auto p-4 md:p-6">{renderScreen()}</div>
      </main>
    </div>
  );
}

export default function App() {
  return (
    <AuthProvider>
      {window.location.pathname === "/auth/callback" ? <AuthCallback /> : <AppContent />}
    </AuthProvider>
  );
}
