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
import './App.css';

function App() {
  const connectWebSocket = useStore((s) => s.connectWebSocket);
  const disconnectWebSocket = useStore((s) => s.disconnectWebSocket);
  const startPolling = useStore((s) => s.startPolling);
  const stopPolling = useStore((s) => s.stopPolling);
  const theme = useStore((s) => s.theme);

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
          <Route path="quests" element={<QuestBoard />} />
          <Route path="war-room" element={<WarRoom />} />
          <Route path="guild-chat" element={<GuildMasterChat />} />
          <Route path="conversations" element={<AgentConversations />} />
          <Route path="stats" element={<StatsView />} />
          <Route path="compare" element={<ComparisonView />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
