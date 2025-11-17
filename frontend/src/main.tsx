import React from 'react'
import ReactDOM from 'react-dom/client'
import './index.css'
import App from './App'

// Wait for DOM to be ready
const rootElement = document.getElementById('root')
if (!rootElement) {
  throw new Error('Root element not found')
}

ReactDOM.createRoot(rootElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
