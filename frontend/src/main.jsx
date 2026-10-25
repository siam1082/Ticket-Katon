import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.jsx'

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <App />
  </StrictMode>,
)

// Hide the splash screen from index.html. It stays for at least ~0.9s after the page started
// loading, so it does not flash by on a fast connection.
const splash = document.getElementById('boot-splash')
if (splash) {
  const wait = Math.max(0, 900 - performance.now())
  setTimeout(() => {
    splash.classList.add('hide')
    setTimeout(() => splash.remove(), 500)
  }, wait)
}
