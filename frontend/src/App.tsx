import { useState, useEffect } from 'react';
import TunnelManager from './components/TunnelManager';
import StatusDisplay from './components/StatusDisplay';
import Settings from './components/Settings';
import './App.css';

function App() {
  const [activeTab, setActiveTab] = useState('tunnel');
  const [tunnelStatus, setTunnelStatus] = useState<any>(null);

  // Fetch tunnel status periodically
  useEffect(() => {
    const fetchStatus = async () => {
      try {
        // Call Wails backend method
        // This will be automatically bound by Wails
        // @ts-ignore
        if (window.go && window.go.app && window.go.app.App) {
          // @ts-ignore
          const status = await window.go.app.App.GetTunnelStatus();
          setTunnelStatus(status);
        }
      } catch (error) {
        console.error('Failed to fetch tunnel status:', error);
      }
    };

    fetchStatus();
    const interval = setInterval(fetchStatus, 3000); // Update every 3 seconds

    return () => clearInterval(interval);
  }, []);

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
