import { useEffect, Suspense, lazy } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { useStore } from './store';
import Layout from './components/Layout';
import ErrorBoundary from './components/ErrorBoundary';
import OfflineIndicator from './components/OfflineIndicator';
import LoadingSpinner from './components/LoadingSpinner';
import { AchievementToastContainer } from './components/AchievementToast';
import { NotificationProvider } from './contexts/NotificationContextProvider';
import NotificationContainer from './components/NotificationContainer';
import SeasonalDecorations from './components/SeasonalDecorations';
import './App.css';

// Lazy load page components for code splitting
const TavernView = lazy(() => import('./pages/TavernView'));
const CharacterSheet = lazy(() => import('./pages/CharacterSheet'));
const QuestBoard = lazy(() => import('./pages/QuestBoard'));
const BountyBoard = lazy(() => import('./pages/BountyBoard'));
const WarRoom = lazy(() => import('./pages/WarRoom'));
const StatsView = lazy(() => import('./pages/StatsView'));
const ComparisonView = lazy(() => import('./pages/ComparisonView'));
const GuildMasterChat = lazy(() => import('./pages/GuildMasterChat'));
const AgentConversations = lazy(() => import('./pages/AgentConversations'));
const Library = lazy(() => import('./pages/Library'));
const TrainingGrounds = lazy(() => import('./pages/TrainingGrounds'));
const Forge = lazy(() => import('./pages/Forge'));
const AgentQuarters = lazy(() => import('./pages/AgentQuarters'));
const CreateAdventurer = lazy(() => import('./pages/CreateAdventurer'));
const Parties = lazy(() => import('./pages/Parties'));
const Leaderboard = lazy(() => import('./pages/Leaderboard'));
const CostDashboard = lazy(() => import('./pages/CostDashboard'));
const PayoutDashboard = lazy(() => import('./pages/PayoutDashboard'));
const MMOChatroom = lazy(() => import('./pages/MMOChatroom'));
const TimeLapseView = lazy(() => import('./pages/TimeLapseView'));
const MapView = lazy(() => import('./pages/MapView'));
const SpectatorMode = lazy(() => import('./pages/SpectatorMode'));
const RelationshipsView = lazy(() => import('./pages/RelationshipsView'));

function App() {
  const connectWebSocket = useStore((s) => s.connectWebSocket);
  const disconnectWebSocket = useStore((s) => s.disconnectWebSocket);
  const startPolling = useStore((s) => s.startPolling);
  const stopPolling = useStore((s) => s.stopPolling);
  const theme = useStore((s) => s.theme);
  const achievementToasts = useStore((s) => s.achievementToasts);
  const removeAchievementToast = useStore((s) => s.removeAchievementToast);
  const updateSeasonalEvent = useStore((s) => s.updateSeasonalEvent);

  useEffect(() => {
    connectWebSocket();
    startPolling(10000);
    
    // Initialize seasonal events
    updateSeasonalEvent();
    
    // Check for seasonal event changes every hour
    const seasonalInterval = setInterval(updateSeasonalEvent, 60 * 60 * 1000);
    
    return () => {
      disconnectWebSocket();
      stopPolling();
      clearInterval(seasonalInterval);
    };
  }, [connectWebSocket, disconnectWebSocket, startPolling, stopPolling, updateSeasonalEvent]);

  // Apply theme on mount and when theme changes
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
  }, [theme]);

  return (
    <ErrorBoundary onError={(error, errorInfo) => {
      console.error('App-level error:', error, errorInfo);
      // Could integrate with error reporting service here
    }}>
      <NotificationProvider>
        <OfflineIndicator />
        <BrowserRouter>
        <Routes>
          <Route path="/" element={<Layout />}>
            <Route index element={
              <Suspense fallback={<LoadingSpinner message="Loading Tavern..." />}>
                <TavernView />
              </Suspense>
            } />
            <Route path="agent/:id" element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading Character Sheet..." />}>
                  <CharacterSheet />
                </Suspense>
              </ErrorBoundary>
            } />
            <Route path="quarters/:id" element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading Agent Quarters..." />}>
                  <AgentQuarters />
                </Suspense>
              </ErrorBoundary>
            } />
            <Route path="quests" element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading Quest Board..." />}>
                  <QuestBoard />
                </Suspense>
              </ErrorBoundary>
            } />
            <Route path="bounties" element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading Bounty Board..." />}>
                  <BountyBoard />
                </Suspense>
              </ErrorBoundary>
            } />
            <Route path="war-room" element={
              <Suspense fallback={<LoadingSpinner message="Loading War Room..." />}>
                <WarRoom />
              </Suspense>
            } />
            <Route path="guild-chat" element={
              <Suspense fallback={<LoadingSpinner message="Loading Guild Chat..." />}>
                <GuildMasterChat />
              </Suspense>
            } />
            <Route path="conversations" element={
              <Suspense fallback={<LoadingSpinner message="Loading Conversations..." />}>
                <AgentConversations />
              </Suspense>
            } />
            <Route path="library" element={
              <Suspense fallback={<LoadingSpinner message="Loading Library..." />}>
                <Library />
              </Suspense>
            } />
            <Route path="training" element={
              <Suspense fallback={<LoadingSpinner message="Loading Training Grounds..." />}>
                <TrainingGrounds />
              </Suspense>
            } />
            <Route path="forge" element={
              <Suspense fallback={<LoadingSpinner message="Loading Forge..." />}>
                <Forge />
              </Suspense>
            } />
            <Route path="stats" element={
              <Suspense fallback={<LoadingSpinner message="Loading Stats..." />}>
                <StatsView />
              </Suspense>
            } />
            <Route path="compare" element={
              <Suspense fallback={<LoadingSpinner message="Loading Comparison..." />}>
                <ComparisonView />
              </Suspense>
            } />
            <Route path="parties" element={
              <Suspense fallback={<LoadingSpinner message="Loading Parties..." />}>
                <Parties />
              </Suspense>
            } />
            <Route path="leaderboard" element={
              <Suspense fallback={<LoadingSpinner message="Loading Leaderboard..." />}>
                <Leaderboard />
              </Suspense>
            } />
            <Route path="costs" element={
              <Suspense fallback={<LoadingSpinner message="Loading Cost Dashboard..." />}>
                <CostDashboard />
              </Suspense>
            } />
            <Route path="payouts" element={
              <Suspense fallback={<LoadingSpinner message="Loading Payout Dashboard..." />}>
                <PayoutDashboard />
              </Suspense>
            } />
            <Route path="chatroom" element={
              <Suspense fallback={<LoadingSpinner message="Loading MMO Chatroom..." />}>
                <MMOChatroom />
              </Suspense>
            } />
            <Route path="timelapse" element={
              <Suspense fallback={<LoadingSpinner message="Loading Time-lapse View..." />}>
                <TimeLapseView />
              </Suspense>
            } />
            <Route path="map" element={
              <Suspense fallback={<LoadingSpinner message="Loading Codebase Map..." />}>
                <MapView />
              </Suspense>
            } />
            <Route path="spectator" element={
              <Suspense fallback={<LoadingSpinner message="Loading Spectator Mode..." />}>
                <SpectatorMode />
              </Suspense>
            } />
            <Route path="relationships" element={
              <Suspense fallback={<LoadingSpinner message="Loading Relationships..." />}>
                <RelationshipsView />
              </Suspense>
            } />
            <Route path="create-adventurer" element={
              <Suspense fallback={<LoadingSpinner message="Loading Adventurer Creator..." />}>
                <CreateAdventurer />
              </Suspense>
            } />
          </Route>
        </Routes>
        
        <ErrorBoundary>
          <AchievementToastContainer
            achievements={achievementToasts}
            onRemove={removeAchievementToast}
          />
        </ErrorBoundary>
        
        <NotificationContainer />
        <SeasonalDecorations />
      </BrowserRouter>
      </NotificationProvider>
    </ErrorBoundary>
  );
}

export default App;
