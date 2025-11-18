import { useState, useEffect } from 'react'
import { AppService } from "../bindings/github.com/votanchat/cloudflared-desktop-tunnel-v3/services"

export function ErrorPopup() {
  const [error, setError] = useState<string>('')

  // Continuously poll for errors
  useEffect(() => {
    const interval = setInterval(async () => {
      try {
        const tunnelError = await AppService.GetLastTunnelError()
        if (tunnelError) {
          setError(tunnelError)
          // Stop tunnel on error
          await AppService.StopAll()
        }
      } catch (err) {
        console.error('Failed to fetch error:', err)
      }
    }, 1000)

    return () => clearInterval(interval)
  }, [])

  if (!error) return null

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-[rgba(255,255,255,0.1)] backdrop-blur-xl rounded-2xl p-6 border border-red-500/50 shadow-2xl max-w-md w-full">
        <div className="flex items-center mb-4">
          <div className="w-12 h-12 bg-red-500 rounded-lg flex items-center justify-center mr-4">
            <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <h2 className="text-xl font-bold text-white">Error</h2>
        </div>
        <p className="text-red-300 mb-4">{error}</p>
        <button
          onClick={() => setError('')}
          className="w-full py-2 px-4 bg-red-500 hover:bg-red-600 text-white font-semibold rounded-lg transition-all"
        >
          Close
        </button>
      </div>
    </div>
  )
}

