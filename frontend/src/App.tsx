import { useEffect } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { useStore } from './store';
import Layout from './components/Layout';
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
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<TavernView />} />
          <Route path="agent/:id" element={<CharacterSheet />} />
          <Route path="quarters/:id" element={<AgentQuarters />} />
          <Route path="quests" element={<QuestBoard />} />
          <Route path="war-room" element={<WarRoom />} />
          <Route path="guild-chat" element={<GuildMasterChat />} />
          <Route path="conversations" element={<AgentConversations />} />
          <Route path="library" element={<Library />} />
          <Route path="training" element={<TrainingGrounds />} />
          <Route path="forge" element={<Forge />} />
          <Route path="stats" element={<StatsView />} />
          <Route path="compare" element={<ComparisonView />} />
          <Route path="parties" element={<Parties />} />
          <Route path="create-adventurer" element={<CreateAdventurer />} />
        </Route>
      </Routes>
      
      <AchievementToastContainer
        achievements={achievementToasts}
        onRemove={removeAchievementToast}
      />
    </BrowserRouter>
  );
}

export default App;
