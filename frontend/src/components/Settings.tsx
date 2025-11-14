import { useState, useEffect } from 'react';

function Settings() {
  const [config, setConfig] = useState<any>(null);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    try {
      // @ts-ignore
      if (window.go && window.go.app && window.go.app.App) {
        // @ts-ignore
        const cfg = await window.go.app.App.GetConfig();
        setConfig(cfg);
      }
    } catch (error) {
      console.error('Failed to load config:', error);
    }
  };

  const handleSave = async () => {
    setIsSaving(true);
    try {
      // @ts-ignore
      await window.go.app.App.UpdateConfig(config);
      alert('Settings saved successfully!');
    } catch (error: any) {
      alert(`Failed to save settings: ${error.message || error}`);
    } finally {
      setIsSaving(false);
    }
  };

  const handleChange = (field: string, value: any) => {
    setConfig({ ...config, [field]: value });
  };

  if (!config) {
    return <div>Loading settings...</div>;
  }

  return (
    <div className="settings">
      <h2>⚙️ Settings</h2>

      <div className="form-group">
        <label className="form-label">Backend URL</label>
        <input
          type="text"
          className="form-input"
          value={config.backendURL || ''}
          onChange={(e) => handleChange('backendURL', e.target.value)}
          placeholder="https://api.example.com"
        />
      </div>

      <div className="form-group">
        <label className="form-label">Tunnel Name</label>
        <input
          type="text"
          className="form-input"
          value={config.tunnelName || ''}
          onChange={(e) => handleChange('tunnelName', e.target.value)}
          placeholder="my-tunnel"
        />
      </div>

      <div className="form-group">
        <label className="form-label">Token Refresh Interval (seconds)</label>
        <input
          type="number"
          className="form-input"
          value={config.refreshInterval || 300}
          onChange={(e) => handleChange('refreshInterval', parseInt(e.target.value))}
          min="60"
          max="3600"
        />
      </div>

      <div className="form-group checkbox-group">
        <input
          type="checkbox"
          id="autoStart"
          checked={config.autoStart || false}
          onChange={(e) => handleChange('autoStart', e.target.checked)}
        />
        <label htmlFor="autoStart" className="form-label" style={{ marginBottom: 0 }}>
          Auto-start tunnel on application startup
        </label>
      </div>

      <div className="form-group checkbox-group">
        <input
          type="checkbox"
          id="minimizeToTray"
          checked={config.minimizeToTray || false}
          onChange={(e) => handleChange('minimizeToTray', e.target.checked)}
        />
        <label htmlFor="minimizeToTray" className="form-label" style={{ marginBottom: 0 }}>
          Minimize to system tray
        </label>
      </div>

      <button
        className="btn btn-primary"
        onClick={handleSave}
        disabled={isSaving}
      >
        {isSaving ? '⏳ Saving...' : '💾 Save Settings'}
      </button>

      <div className="info-card" style={{ marginTop: '30px' }}>
        <h3>About</h3>
        <div className="info-row">
          <span className="info-label">Version:</span>
          <span className="info-value">1.0.0</span>
        </div>
        <div className="info-row">
          <span className="info-label">Framework:</span>
          <span className="info-value">Wails v2 + React + TypeScript</span>
        </div>
        <div className="info-row">
          <span className="info-label">Repository:</span>
          <span className="info-value">
            <a 
              href="https://github.com/votanchat/cloudflared-desktop-tunnel" 
              target="_blank" 
              rel="noopener noreferrer"
              style={{ color: '#667eea', textDecoration: 'none' }}
            >
              GitHub
            </a>
          </span>
        </div>
      </div>
    </div>
  );
}

export default Settings;
