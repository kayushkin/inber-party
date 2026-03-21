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
import WSDebugPanel from './components/WSDebugPanel';
import ErrorTestButtons from './components/ErrorTestButtons';
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
    // Use consistent test environment detection
    const isTest = !!(
      '__playwright' in globalThis || 
      import.meta.env.VITE_NODE_ENV === 'test' ||
      (typeof navigator !== 'undefined' && (
        navigator.userAgent.includes('HeadlessChrome') ||
        navigator.userAgent.includes('Firefox') && navigator.webdriver
      ))
    );
    
    if (isTest) {
      // In test environment, delay WebSocket connection to ensure backend is ready
      console.log('🧪 Test environment detected, using delayed initialization with persistent connections');
      const testTimer = setTimeout(() => {
        connectWebSocket();
        startPolling(15000); // Longer polling interval for tests
      }, 2000); // Reduced delay from 3s to 2s
      
      // Initialize seasonal events (no delay needed)
      updateSeasonalEvent();
      
      return () => {
        clearTimeout(testTimer);
        // In test environment, avoid disconnecting to prevent connection churn
        console.log('🧪 Test environment cleanup: Skipping WebSocket disconnection to prevent churn');
        stopPolling();
      };
    } else {
      // Normal environment, connect immediately
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
    }
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
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading Tavern..." />}>
                  <TavernView />
                </Suspense>
              </ErrorBoundary>
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
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading War Room..." />}>
                  <WarRoom />
                </Suspense>
              </ErrorBoundary>
            } />
            <Route path="guild-chat" element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading Guild Chat..." />}>
                  <GuildMasterChat />
                </Suspense>
              </ErrorBoundary>
            } />
            <Route path="conversations" element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading Conversations..." />}>
                  <AgentConversations />
                </Suspense>
              </ErrorBoundary>
            } />
            <Route path="library" element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading Library..." />}>
                  <Library />
                </Suspense>
              </ErrorBoundary>
            } />
            <Route path="training" element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading Training Grounds..." />}>
                  <TrainingGrounds />
                </Suspense>
              </ErrorBoundary>
            } />
            <Route path="forge" element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading Forge..." />}>
                  <Forge />
                </Suspense>
              </ErrorBoundary>
            } />
            <Route path="stats" element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading Stats..." />}>
                  <StatsView />
                </Suspense>
              </ErrorBoundary>
            } />
            <Route path="compare" element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading Comparison..." />}>
                  <ComparisonView />
                </Suspense>
              </ErrorBoundary>
            } />
            <Route path="parties" element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading Parties..." />}>
                  <Parties />
                </Suspense>
              </ErrorBoundary>
            } />
            <Route path="leaderboard" element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading Leaderboard..." />}>
                  <Leaderboard />
                </Suspense>
              </ErrorBoundary>
            } />
            <Route path="costs" element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading Cost Dashboard..." />}>
                  <CostDashboard />
                </Suspense>
              </ErrorBoundary>
            } />
            <Route path="payouts" element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading Payout Dashboard..." />}>
                  <PayoutDashboard />
                </Suspense>
              </ErrorBoundary>
            } />
            <Route path="chatroom" element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading MMO Chatroom..." />}>
                  <MMOChatroom />
                </Suspense>
              </ErrorBoundary>
            } />
            <Route path="timelapse" element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading Time-lapse View..." />}>
                  <TimeLapseView />
                </Suspense>
              </ErrorBoundary>
            } />
            <Route path="map" element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading Codebase Map..." />}>
                  <MapView />
                </Suspense>
              </ErrorBoundary>
            } />
            <Route path="spectator" element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading Spectator Mode..." />}>
                  <SpectatorMode />
                </Suspense>
              </ErrorBoundary>
            } />
            <Route path="relationships" element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading Relationships..." />}>
                  <RelationshipsView />
                </Suspense>
              </ErrorBoundary>
            } />
            <Route path="create-adventurer" element={
              <ErrorBoundary>
                <Suspense fallback={<LoadingSpinner message="Loading Adventurer Creator..." />}>
                  <CreateAdventurer />
                </Suspense>
              </ErrorBoundary>
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
        
        {/* WebSocket Debug Panel - only visible in development */}
        {import.meta.env.DEV && <WSDebugPanel />}
        
        {/* Error Testing Buttons - only visible in development */}
        {import.meta.env.DEV && <ErrorTestButtons />}
      </BrowserRouter>
      </NotificationProvider>
    </ErrorBoundary>
  );
}

export default App;
