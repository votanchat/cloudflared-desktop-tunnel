import { useState, useEffect } from 'react'
// Bindings will be generated when running wails3 dev or wails3 build
// Import path will be: "../bindings/github.com/votanchat/cloudflared-desktop-tunnel-v3/services"
// For now, using a placeholder - update after bindings are generated
// @ts-ignore
import { AppService } from "../bindings/changeme/services"

function App() {
  const [tunnelStatus, setTunnelStatus] = useState<any>(null)
  const [webServerStatus, setWebServerStatus] = useState<any>(null)
  const [config, setConfig] = useState<any>(null)
  const [manualToken, setManualToken] = useState<string>('')
  const [loading, setLoading] = useState<boolean>(false)

  useEffect(() => {
    loadStatus()
    const interval = setInterval(loadStatus, 2000)
    return () => clearInterval(interval)
  }, [])

  const loadStatus = async () => {
    try {
      const [tunnel, webServer, cfg] = await Promise.all([
        AppService.GetTunnelStatus(),
        AppService.GetWebServerStatus(),
        AppService.GetConfig()
      ])
      setTunnelStatus(tunnel)
      setWebServerStatus(webServer)
      setConfig(cfg)
    } catch (err) {
      console.error('Failed to load status:', err)
    }
  }

  const handleStartTunnel = async () => {
    setLoading(true)
    try {
      await AppService.StartTunnel(manualToken || '')
      setManualToken('')
      await loadStatus()
    } catch (err: any) {
      alert('Failed to start tunnel: ' + (err.message || err))
    } finally {
      setLoading(false)
    }
  }

  const handleStopTunnel = async () => {
    setLoading(true)
    try {
      await AppService.StopTunnel()
      await loadStatus()
    } catch (err: any) {
      alert('Failed to stop tunnel: ' + (err.message || err))
    } finally {
      setLoading(false)
    }
  }

  const handleStartWebServer = async () => {
    setLoading(true)
    try {
      await AppService.StartWebServer(0)
      await loadStatus()
    } catch (err: any) {
      alert('Failed to start web server: ' + (err.message || err))
    } finally {
      setLoading(false)
    }
  }

  const handleStopWebServer = async () => {
    setLoading(true)
    try {
      await AppService.StopWebServer()
      await loadStatus()
    } catch (err: any) {
      alert('Failed to stop web server: ' + (err.message || err))
    } finally {
      setLoading(false)
    }
  }

  const isTunnelRunning = tunnelStatus?.running || false
  const isWebServerRunning = webServerStatus?.running || false

  return (
    <div style={styles.container}>
      <div style={styles.content}>
        <h1 style={styles.title}>Cloudflared Desktop Tunnel</h1>

        {/* Tunnel Status */}
        <div style={styles.section}>
          <h2 style={styles.sectionTitle}>Tunnel Status</h2>
          <div style={styles.statusBox}>
            <div style={styles.statusRow}>
              <span style={styles.label}>Status:</span>
              <span style={{
                ...styles.statusBadge,
                backgroundColor: isTunnelRunning ? '#4caf50' : '#f44336'
              }}>
                {isTunnelRunning ? 'Running' : 'Stopped'}
              </span>
            </div>
            {tunnelStatus?.tunnelURL && (
              <div style={styles.statusRow}>
                <span style={styles.label}>URL:</span>
                <span style={styles.value}>{tunnelStatus.tunnelURL}</span>
              </div>
            )}
            {tunnelStatus?.tunnelName && (
              <div style={styles.statusRow}>
                <span style={styles.label}>Name:</span>
                <span style={styles.value}>{tunnelStatus.tunnelName}</span>
              </div>
            )}
          </div>

          <div style={styles.controls}>
            {!isTunnelRunning ? (
              <>
                <input
                  type="text"
                  placeholder="Manual token (optional)"
                  value={manualToken}
                  onChange={(e) => setManualToken(e.target.value)}
                  style={styles.input}
                />
                <button
                  onClick={handleStartTunnel}
                  disabled={loading}
                  style={styles.button}
                >
                  Start Tunnel
                </button>
              </>
            ) : (
              <button
                onClick={handleStopTunnel}
                disabled={loading}
                style={{ ...styles.button, backgroundColor: '#f44336' }}
              >
                Stop Tunnel
              </button>
            )}
          </div>
        </div>

        {/* Web Server Status */}
        <div style={styles.section}>
          <h2 style={styles.sectionTitle}>Web Server Status</h2>
          <div style={styles.statusBox}>
            <div style={styles.statusRow}>
              <span style={styles.label}>Status:</span>
              <span style={{
                ...styles.statusBadge,
                backgroundColor: isWebServerRunning ? '#4caf50' : '#f44336'
              }}>
                {isWebServerRunning ? 'Running' : 'Stopped'}
              </span>
            </div>
            {webServerStatus?.port && (
              <div style={styles.statusRow}>
                <span style={styles.label}>Port:</span>
                <span style={styles.value}>{webServerStatus.port}</span>
              </div>
            )}
          </div>

          <div style={styles.controls}>
            {!isWebServerRunning ? (
              <button
                onClick={handleStartWebServer}
                disabled={loading}
                style={styles.button}
              >
                Start Web Server
              </button>
            ) : (
              <button
                onClick={handleStopWebServer}
                disabled={loading}
                style={{ ...styles.button, backgroundColor: '#f44336' }}
              >
                Stop Web Server
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

const styles = {
  container: {
    width: '100%',
    height: '100vh',
    backgroundColor: '#1a1a1a',
    color: '#ffffff',
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
  },
  content: {
    width: '100%',
    maxWidth: '600px',
    padding: '40px',
  },
  title: {
    fontSize: '28px',
    fontWeight: 'bold',
    marginBottom: '30px',
    textAlign: 'center' as const,
  },
  section: {
    marginBottom: '30px',
    backgroundColor: '#2a2a2a',
    padding: '20px',
    borderRadius: '8px',
  },
  sectionTitle: {
    fontSize: '18px',
    fontWeight: '600',
    marginBottom: '15px',
    color: '#ffffff',
  },
  statusBox: {
    backgroundColor: '#1a1a1a',
    padding: '15px',
    borderRadius: '6px',
    marginBottom: '15px',
  },
  statusRow: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '10px',
  },
  label: {
    fontSize: '14px',
    color: '#aaaaaa',
  },
  value: {
    fontSize: '14px',
    color: '#ffffff',
    fontWeight: '500',
  },
  statusBadge: {
    padding: '4px 12px',
    borderRadius: '12px',
    fontSize: '12px',
    fontWeight: '600',
    color: '#ffffff',
  },
  controls: {
    display: 'flex',
    gap: '10px',
    flexDirection: 'column' as const,
  },
  input: {
    padding: '10px',
    borderRadius: '6px',
    border: '1px solid #444',
    backgroundColor: '#1a1a1a',
    color: '#ffffff',
    fontSize: '14px',
  },
  button: {
    padding: '12px 24px',
    borderRadius: '6px',
    border: 'none',
    backgroundColor: '#4caf50',
    color: '#ffffff',
    fontSize: '14px',
    fontWeight: '600',
    cursor: 'pointer',
    transition: 'background-color 0.2s',
  },
}

export default App
