interface StatusDisplayProps {
  status: any;
}

function StatusDisplay({ status }: StatusDisplayProps) {
  const isRunning = status?.running || false;

  return (
    <div className="status-display">
      <h2>📊 Tunnel Status</h2>

      <div className="info-card">
        <div className="info-row">
          <span className="info-label">Connection Status:</span>
          <span className={`status-indicator ${isRunning ? 'running' : 'stopped'}`}>
            <span className={`status-dot ${isRunning ? 'running' : 'stopped'}`}></span>
            {isRunning ? 'Connected' : 'Disconnected'}
          </span>
        </div>
        <div className="info-row">
          <span className="info-label">Tunnel Name:</span>
          <span className="info-value">{status?.tunnelName || 'N/A'}</span>
        </div>
        <div className="info-row">
          <span className="info-label">Tunnel URL:</span>
          <span className="info-value">
            {isRunning ? `https://${status?.tunnelName}.cfargotunnel.com` : 'Not running'}
          </span>
        </div>
      </div>

      <div>
        <h3>Connection Details</h3>
        <div className="info-card">
          <div className="info-row">
            <span className="info-label">Protocol:</span>
            <span className="info-value">QUIC</span>
          </div>
          <div className="info-row">
            <span className="info-label">Connections:</span>
            <span className="info-value">{isRunning ? '4' : '0'}</span>
          </div>
          <div className="info-row">
            <span className="info-label">Edge Location:</span>
            <span className="info-value">{isRunning ? 'Auto-selected' : 'N/A'}</span>
          </div>
        </div>
      </div>

      <div>
        <h3>Full Logs</h3>
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

export default StatusDisplay;
