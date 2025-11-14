import { useState, useEffect } from 'react';
import TunnelManager from './components/TunnelManager';
import StatusDisplay from './components/StatusDisplay';
import Settings from './components/Settings';
import './App.css';

function App() {
  const [activeTab, setActiveTab] = useState('tunnel');
  const [tunnelStatus, setTunnelStatus] = useState<any>(null);
  const [wailsReady, setWailsReady] = useState(false);

  // Check if Wails runtime is ready
  useEffect(() => {
    const checkWails = () => {
      if (window.go && window.go.app && window.go.app.App) {
        setWailsReady(true);
        return true;
      }
      return false;
    };

    // Try immediately
    if (checkWails()) return;

    // Otherwise poll every 100ms for up to 5 seconds
    let attempts = 0;
    const interval = setInterval(() => {
      attempts++;
      if (checkWails() || attempts > 50) {
        clearInterval(interval);
      }
    }, 100);

    return () => clearInterval(interval);
  }, []);

  // Fetch tunnel status periodically
  useEffect(() => {
    if (!wailsReady) return;

    const fetchStatus = async () => {
      try {
        const status = await window.go.app.App.GetTunnelStatus();
        setTunnelStatus(status);
      } catch (error) {
        console.error('Failed to fetch tunnel status:', error);
      }
    };

    fetchStatus();
    const interval = setInterval(fetchStatus, 3000);

    return () => clearInterval(interval);
  }, [wailsReady]);

  if (!wailsReady) {
    return (
      <div className="app" style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: '100vh' }}>
        <div style={{ textAlign: 'center', color: 'white' }}>
          <h2>🔄 Loading Wails Runtime...</h2>
          <p>Please make sure you're running: <code>wails dev</code></p>
        </div>
      </div>
    );
  }

  return (
    <div className="app">
      <header className="app-header">
        <h1>🔒 Cloudflared Desktop Tunnel</h1>
        <p className="subtitle">Manage your Cloudflare Tunnels with ease</p>
      </header>

      <nav className="tabs">
        <button
          className={`tab ${activeTab === 'tunnel' ? 'active' : ''}`}
          onClick={() => setActiveTab('tunnel')}
        >
          🏛️ Tunnel
        </button>
        <button
          className={`tab ${activeTab === 'status' ? 'active' : ''}`}
          onClick={() => setActiveTab('status')}
        >
          📊 Status
        </button>
        <button
          className={`tab ${activeTab === 'settings' ? 'active' : ''}`}
          onClick={() => setActiveTab('settings')}
        >
          ⚙️ Settings
        </button>
      </nav>

      <main className="app-content">
        {activeTab === 'tunnel' && <TunnelManager status={tunnelStatus} />}
        {activeTab === 'status' && <StatusDisplay status={tunnelStatus} />}
        {activeTab === 'settings' && <Settings />}
      </main>

      <footer className="app-footer">
        <p>© 2025 Cloudflared Desktop Tunnel | Powered by Wails & React</p>
      </footer>
    </div>
  );
}

export default App;
