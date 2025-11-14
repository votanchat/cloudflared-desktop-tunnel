import { useState } from 'react';

interface TunnelManagerProps {
  status: any;
}

function TunnelManager({ status }: TunnelManagerProps) {
  const [isLoading, setIsLoading] = useState(false);

  const handleStart = async () => {
    setIsLoading(true);
    try {
      // @ts-ignore
      await window.go.app.App.StartTunnel();
    } catch (error: any) {
      alert(`Failed to start tunnel: ${error.message || error}`);
    } finally {
      setIsLoading(false);
    }
  };

  const handleStop = async () => {
    setIsLoading(true);
    try {
      // @ts-ignore
      await window.go.app.App.StopTunnel();
    } catch (error: any) {
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
