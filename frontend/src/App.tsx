import { useEffect } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { useStore } from './store';
import Layout from './components/Layout';
import CampView from './pages/CampView';
import CharacterSheet from './pages/CharacterSheet';
import QuestBoard from './pages/QuestBoard';
import StatsView from './pages/StatsView';
import './App.css';

function App() {
  const connectWebSocket = useStore((s) => s.connectWebSocket);
  const disconnectWebSocket = useStore((s) => s.disconnectWebSocket);
  const startPolling = useStore((s) => s.startPolling);
  const stopPolling = useStore((s) => s.stopPolling);

  useEffect(() => {
    connectWebSocket();
    startPolling(10000);
    return () => {
      disconnectWebSocket();
      stopPolling();
    };
  }, [connectWebSocket, disconnectWebSocket, startPolling, stopPolling]);

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<CampView />} />
          <Route path="agent/:id" element={<CharacterSheet />} />
          <Route path="quests" element={<QuestBoard />} />
          <Route path="stats" element={<StatsView />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
