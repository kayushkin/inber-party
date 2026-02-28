import { useEffect } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { useStore } from './store';
import Layout from './components/Layout';
import CampView from './pages/CampView';
import CharacterSheet from './pages/CharacterSheet';
import QuestBoard from './pages/QuestBoard';
import './App.css';

function App() {
  const connectWebSocket = useStore((state) => state.connectWebSocket);
  const disconnectWebSocket = useStore((state) => state.disconnectWebSocket);

  useEffect(() => {
    connectWebSocket();
    return () => disconnectWebSocket();
  }, [connectWebSocket, disconnectWebSocket]);

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<CampView />} />
          <Route path="agent/:id" element={<CharacterSheet />} />
          <Route path="quests" element={<QuestBoard />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
