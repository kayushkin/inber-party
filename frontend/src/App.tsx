import { useEffect } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { useStore } from './store';
import Layout from './components/Layout';
import ErrorBoundary, { AgentErrorFallback, QuestErrorFallback } from './components/ErrorBoundary';
import AsyncErrorBoundary from './components/AsyncErrorBoundary';
import TavernView from './pages/TavernView';
import CharacterSheet from './pages/CharacterSheet';
import QuestBoard from './pages/QuestBoard';
import WarRoom from './pages/WarRoom';
import StatsView from './pages/StatsView';
import ComparisonView from './pages/ComparisonView';
import GuildMasterChat from './pages/GuildMasterChat';
import AgentConversations from './pages/AgentConversations';
import Library from './pages/Library';
import TrainingGrounds from './pages/TrainingGrounds';
import Forge from './pages/Forge';
import AgentQuarters from './pages/AgentQuarters';
import CreateAdventurer from './pages/CreateAdventurer';
import Parties from './pages/Parties';
import Leaderboard from './pages/Leaderboard';
import CostDashboard from './pages/CostDashboard';
import { AchievementToastContainer } from './components/AchievementToast';
import './App.css';

function App() {
  const connectWebSocket = useStore((s) => s.connectWebSocket);
  const disconnectWebSocket = useStore((s) => s.disconnectWebSocket);
  const startPolling = useStore((s) => s.startPolling);
  const stopPolling = useStore((s) => s.stopPolling);
  const theme = useStore((s) => s.theme);
  const achievementToasts = useStore((s) => s.achievementToasts);
  const removeAchievementToast = useStore((s) => s.removeAchievementToast);

  useEffect(() => {
    connectWebSocket();
    startPolling(10000);
    return () => {
      disconnectWebSocket();
      stopPolling();
    };
  }, [connectWebSocket, disconnectWebSocket, startPolling, stopPolling]);

  // Apply theme on mount and when theme changes
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
  }, [theme]);

  return (
    <ErrorBoundary onError={(error, errorInfo) => {
      console.error('App-level error:', error, errorInfo);
      // Could integrate with error reporting service here
    }}>
      <BrowserRouter>
        <AsyncErrorBoundary>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route index element={<TavernView />} />
              <Route path="agent/:id" element={
                <ErrorBoundary fallback={AgentErrorFallback}>
                  <CharacterSheet />
                </ErrorBoundary>
              } />
              <Route path="quarters/:id" element={
                <ErrorBoundary fallback={AgentErrorFallback}>
                  <AgentQuarters />
                </ErrorBoundary>
              } />
              <Route path="quests" element={
                <ErrorBoundary fallback={QuestErrorFallback}>
                  <QuestBoard />
                </ErrorBoundary>
              } />
              <Route path="war-room" element={<WarRoom />} />
              <Route path="guild-chat" element={<GuildMasterChat />} />
              <Route path="conversations" element={<AgentConversations />} />
              <Route path="library" element={<Library />} />
              <Route path="training" element={<TrainingGrounds />} />
              <Route path="forge" element={<Forge />} />
              <Route path="stats" element={<StatsView />} />
              <Route path="compare" element={<ComparisonView />} />
              <Route path="parties" element={<Parties />} />
              <Route path="leaderboard" element={<Leaderboard />} />
              <Route path="costs" element={<CostDashboard />} />
              <Route path="create-adventurer" element={<CreateAdventurer />} />
            </Route>
          </Routes>
        </AsyncErrorBoundary>
        
        <ErrorBoundary>
          <AchievementToastContainer
            achievements={achievementToasts}
            onRemove={removeAchievementToast}
          />
        </ErrorBoundary>
      </BrowserRouter>
    </ErrorBoundary>
  );
}

export default App;
