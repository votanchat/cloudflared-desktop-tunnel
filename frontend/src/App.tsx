import { useState, useEffect } from 'react'
import { AppService } from "../bindings/github.com/votanchat/cloudflared-desktop-tunnel-v3/services"
import { ErrorPopup } from './ErrorPopup'

function App() {
  const [token, setToken] = useState<string>('')
  const [loading, setLoading] = useState<boolean>(false)
  const [status, setStatus] = useState<any>(null)
  const [error, setError] = useState<string>('')

  // Poll status every 2 seconds if tunnel is running
  useEffect(() => {
    if (!status?.running) return

    const interval = setInterval(async () => {
      try {
        const tunnelStatus = await AppService.GetTunnelStatus()
        const webServerStatus = await AppService.GetWebServerStatus()
        
        if (!tunnelStatus.running || !webServerStatus.running) {
          await AppService.StopAll()
        }
        
        setStatus({
          running: tunnelStatus.running && webServerStatus.running,
          tunnel: tunnelStatus,
          webServer: webServerStatus
        })
      } catch (err) {
        console.error('Failed to fetch status:', err)
      }
    }, 2000)

    return () => clearInterval(interval)
  }, [status?.running])

  const handleStart = async () => {
    if (!token.trim()) {
      setError('Please enter a token')
      return
    }

    setLoading(true)
    setError('')
    try {
        await AppService.StartTunnel(token)
        setToken('')
      
      // Get status after successful start
      const tunnelStatus = await AppService.GetTunnelStatus()
      const webServerStatus = await AppService.GetWebServerStatus()
      setStatus({
        running: true,
        tunnel: tunnelStatus,
        webServer: webServerStatus
      })
    } catch (err: any) {
      setError('Starting failed: ' + (err.message || err))
      setStatus(null)
    } finally {
      setLoading(false)
    }
  }

  const handleStop = async () => {
    setLoading(true)
    try {
      await AppService.StopAll()
      setStatus(null)
    } catch (err: any) {
      setError('Stopping failed: ' + (err.message || err))
    } finally {
      setLoading(false)
    }
  }

  // Show status if running
  if (status?.running) {
    return (
      <div className="w-full min-h-screen flex items-center justify-center bg-[#1b2636] p-4">
        <ErrorPopup />
        <div className="w-full max-w-md">
          <div className="bg-[rgba(255,255,255,0.05)] backdrop-blur-xl rounded-2xl p-8 border border-[rgba(255,255,255,0.1)] shadow-2xl">
            <div className="flex justify-center mb-6">
              <div className="w-16 h-16 bg-green-500 rounded-lg flex items-center justify-center">
                <svg className="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
              </div>
            </div>

            <h1 className="text-3xl font-bold text-center text-white mb-2">
              Running
            </h1>
            <p className="text-sm text-center text-[rgba(255,255,255,0.7)] mb-6">
              Tunnel has been started successfully
            </p>

            <div className="space-y-4 mb-6">
              <div className="bg-[rgba(255,255,255,0.05)] rounded-lg p-4">
                <div className="text-sm text-[rgba(255,255,255,0.7)] mb-1">Tunnel Status</div>
                <div className="text-white font-semibold">
                  {status.tunnel?.running ? 'Running' : 'Stopped'}
                </div>
              </div>

              <div className="bg-[rgba(255,255,255,0.05)] rounded-lg p-4">
                <div className="text-sm text-[rgba(255,255,255,0.7)] mb-1">Web Server Status</div>
                <div className="text-white font-semibold">
                  {status.webServer?.running ? 'Running' : 'Stopped'}
                </div>
                {status.webServer?.port && (
                  <div className="text-xs text-[rgba(255,255,255,0.5)] mt-1">
                    Port: {status.webServer.port}
                  </div>
                )}
              </div>
            </div>

            <button
              onClick={handleStop}
              disabled={loading}
              className="w-full py-3 px-4 bg-red-500 hover:bg-red-600 text-white font-semibold rounded-lg transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? 'Stopping...' : 'Stop'}
            </button>
          </div>
        </div>
      </div>
    )
  }

  // Show login form
  return (
    <div className="w-full min-h-screen flex items-center justify-center bg-[#1b2636] p-4">
      <ErrorPopup />
      <div className="w-full max-w-md">
        <div className="bg-[rgba(255,255,255,0.05)] backdrop-blur-xl rounded-2xl p-8 border border-[rgba(255,255,255,0.1)] shadow-2xl">
          <div className="flex justify-center mb-6">
            <div className="w-16 h-16 bg-[#60a5fa] rounded-lg flex items-center justify-center">
              <svg 
                className="w-10 h-10 text-yellow-400" 
                fill="currentColor" 
                viewBox="0 0 20 20"
              >
                <path 
                  fillRule="evenodd" 
                  d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" 
                  clipRule="evenodd" 
                />
              </svg>
            </div>
          </div>

          <h1 className="text-3xl font-bold text-center text-white mb-2">
            Cloudflared Tunnel
          </h1>
          <p className="text-sm text-center text-[rgba(255,255,255,0.7)] mb-6">
            Enter token to start
          </p>

            <div className="mb-6">
              <label className="block text-sm text-[rgba(255,255,255,0.7)] mb-2">
                Token
              </label>
              <input
                type="text"
                placeholder="Enter your token"
                value={token}
                onChange={(e) => setToken(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && handleStart()}
                className="w-full px-4 py-3 rounded-lg bg-[rgba(255,255,255,0.05)] border border-[rgba(255,255,255,0.1)] text-white placeholder-[rgba(255,255,255,0.4)] focus:outline-none focus:border-[#60a5fa] focus:ring-2 focus:ring-[#60a5fa]/20 transition-all"
              />
          </div>

          {error && (
            <div className="mb-4 p-3 bg-red-500/20 border border-red-500/50 rounded-lg text-red-300 text-sm">
              {error}
            </div>
          )}

          <button
            onClick={handleStart}
            disabled={loading || !token.trim()}
            className="w-full py-3 px-4 bg-[#60a5fa] hover:bg-[#3b82f6] text-white font-semibold rounded-lg transition-all disabled:opacity-50 disabled:cursor-not-allowed shadow-lg hover:shadow-xl"
          >
            {loading ? 'Starting...' : 'Start'}
          </button>
        </div>
      </div>
    </div>
  )
}

export default App
