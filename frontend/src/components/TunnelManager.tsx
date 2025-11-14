import { useState } from 'react';

interface TunnelManagerProps {
  status: any;
}

function TunnelManager({ status }: TunnelManagerProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [manualToken, setManualToken] = useState('');
  const [showTokenInput, setShowTokenInput] = useState(false);

  const handleStart = async () => {
    setIsLoading(true);
    try {
      // Check if Wails runtime is available
      if (!window.go || !window.go.app || !window.go.app.App) {
        throw new Error('Wails runtime not initialized. Please run: wails dev');
      }
      
      // Pass manual token (empty string if not provided)
      await window.go.app.App.StartTunnel(manualToken);
      console.log('Tunnel started successfully');
      
      // Clear token after successful start for security
      if (manualToken) {
        setManualToken('');
      }
    } catch (error: any) {
      console.error('Start tunnel error:', error);
      alert(`Failed to start tunnel: ${error.message || error}`);
    } finally {
      setIsLoading(false);
    }
  };

  const handleStop = async () => {
    setIsLoading(true);
    try {
      if (!window.go || !window.go.app || !window.go.app.App) {
        throw new Error('Wails runtime not initialized. Please run: wails dev');
      }
      
      await window.go.app.App.StopTunnel();
      console.log('Tunnel stopped successfully');
    } catch (error: any) {
      console.error('Stop tunnel error:', error);
      alert(`Failed to stop tunnel: ${error.message || error}`);
    } finally {
      setIsLoading(false);
    }
  };

  const isRunning = status?.running || false;

  return (
    <div className="tunnel-manager">
      <h2>🏛️ Tunnel Control</h2>

      <div className="info-card">
        <div className="info-row">
          <span className="info-label">Status:</span>
          <span className="status-indicator" style={{ display: 'inline-flex', alignItems: 'center', gap: '8px' }}>
            <span className={`status-dot ${isRunning ? 'running' : 'stopped'}`}></span>
            {isRunning ? 'Running' : 'Stopped'}
          </span>
        </div>
        <div className="info-row">
          <span className="info-label">Tunnel Name:</span>
          <span className="info-value">{status?.tunnelName || 'N/A'}</span>
        </div>
      </div>

      {/* Manual Token Input Section */}
      <div className="info-card" style={{ marginBottom: '20px' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '10px' }}>
          <h3 style={{ margin: 0, fontSize: '1rem' }}>🔑 Token Configuration</h3>
          <button
            className="btn"
            onClick={() => setShowTokenInput(!showTokenInput)}
            style={{
              padding: '6px 12px',
              fontSize: '0.85rem',
              background: showTokenInput ? '#f5576c' : '#667eea',
              color: 'white',
              border: 'none',
              borderRadius: '6px',
              cursor: 'pointer'
            }}
          >
            {showTokenInput ? '❌ Hide' : '✏️ Manual Token'}
          </button>
        </div>

        {showTokenInput && (
          <div>
            <label className="form-label" style={{ fontSize: '0.9rem', marginBottom: '8px', display: 'block' }}>
              Paste your Cloudflare Tunnel token here (optional)
            </label>
            <textarea
              className="form-input"
              value={manualToken}
              onChange={(e) => setManualToken(e.target.value)}
              placeholder="eyJhIjoiMTIzNDU2Nzg5MGFiY2RlZiIsInQiOiJhYmNkZWYxMi0zNDU2LTc4OTAtYWJjZC1lZjEyMzQ1Njc4OTAiLCJzIjoi..."
              disabled={isRunning}
              rows={3}
              style={{
                fontFamily: 'monospace',
                fontSize: '0.85rem',
                resize: 'vertical'
              }}
            />
            <p style={{ fontSize: '0.85rem', color: '#6c757d', marginTop: '8px', marginBottom: 0 }}>
              💡 <strong>Tip:</strong> Leave empty to fetch token from backend automatically.
            </p>
          </div>
        )}

        {!showTokenInput && (
          <p style={{ fontSize: '0.9rem', color: '#6c757d', margin: 0 }}>
            Token will be fetched from backend API when you start the tunnel.
          </p>
        )}
      </div>

      <div className="control-panel">
        <button
          className="btn btn-primary"
          onClick={handleStart}
          disabled={isRunning || isLoading}
        >
          {isLoading ? '⏳ Starting...' : '▶️ Start Tunnel'}
        </button>
        <button
          className="btn btn-danger"
          onClick={handleStop}
          disabled={!isRunning || isLoading}
        >
          {isLoading ? '⏳ Stopping...' : '⏸️ Stop Tunnel'}
        </button>
      </div>

      <div>
        <h3>Recent Logs</h3>
        <div className="logs-container">
          {status?.logs && status.logs.length > 0 ? (
            status.logs.map((log: string, index: number) => (
              <div key={index} className="log-line">
                {log}
              </div>
            ))
          ) : (
            <div className="log-line" style={{ opacity: 0.5 }}>No logs available</div>
          )}
        </div>
      </div>
    </div>
  );
}

export default TunnelManager;
