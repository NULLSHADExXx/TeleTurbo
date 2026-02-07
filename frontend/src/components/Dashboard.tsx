import { useState, useEffect, useCallback } from 'react'
import DownloadItem from './DownloadItem'

interface DashboardProps {
  onLogout: () => void
}

interface DownloadInfo {
  id: string
  filename: string
  progress: number
  downloaded: number
  total: number
  speed: number
  status: string
}

function Dashboard({ onLogout }: DashboardProps) {
  const [link, setLink] = useState('')
  const [destination, setDestination] = useState('~/Downloads/TeleTurbo')
  const [downloads, setDownloads] = useState<DownloadInfo[]>([])
  const [systemInfo, setSystemInfo] = useState<Record<string, any> | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    loadSystemInfo()
    const interval = setInterval(refreshDownloads, 1000)
    return () => clearInterval(interval)
  }, [])

  const loadSystemInfo = async () => {
    try {
      const info = await window.go.main.App.GetSystemInfo()
      setSystemInfo(info)
    } catch (err) {
      console.error('Failed to load system info:', err)
    }
  }

  const refreshDownloads = useCallback(async () => {
    try {
      const allDownloads = await window.go.main.App.GetAllDownloads()
      setDownloads(allDownloads)
    } catch (err) {
      console.error('Failed to refresh downloads:', err)
    }
  }, [])

  const handleDownload = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!link.trim()) return

    setLoading(true)
    try {
      await window.go.main.App.StartDownload(link.trim(), destination)
      setLink('')
      refreshDownloads()
    } catch (err) {
      console.error('Failed to start download:', err)
    }
    setLoading(false)
  }

  const handleCancel = async (id: string) => {
    try {
      await window.go.main.App.CancelDownload(id)
      refreshDownloads()
    } catch (err) {
      console.error('Failed to cancel download:', err)
    }
  }

  const formatSpeed = (bytesPerSecond: number): string => {
    if (bytesPerSecond === 0) return '-'
    const units = ['B/s', 'KB/s', 'MB/s', 'GB/s']
    let unitIndex = 0
    let speed = bytesPerSecond
    
    while (speed >= 1024 && unitIndex < units.length - 1) {
      speed /= 1024
      unitIndex++
    }
    
    return `${speed.toFixed(1)} ${units[unitIndex]}`
  }

  const formatSize = (bytes: number): string => {
    if (bytes === 0) return '0 B'
    const units = ['B', 'KB', 'MB', 'GB', 'TB']
    let unitIndex = 0
    let size = bytes
    
    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024
      unitIndex++
    }
    
    return `${size.toFixed(1)} ${units[unitIndex]}`
  }

  return (
    <div className="dashboard">
      <header className="dashboard-header">
        <div className="header-left">
          <h1>TeleTurbo</h1>
          <span className="version">v1.0.0</span>
        </div>
        <div className="header-right">
          {systemInfo && (
            <div className="system-info">
              <span className="badge">
                {systemInfo.parallelism} Threads
              </span>
              <span className="badge">
                {systemInfo.cpu_cores} Cores
              </span>
            </div>
          )}
          <button className="btn-logout" onClick={onLogout}>
            Logout
          </button>
        </div>
      </header>

      <main className="dashboard-main">
        <section className="download-section">
          <form onSubmit={handleDownload} className="download-form">
            <div className="input-row">
              <div className="input-group flex-grow">
                <label>Telegram Link</label>
                <input
                  type="text"
                  value={link}
                  onChange={(e) => setLink(e.target.value)}
                  placeholder="https://t.me/c/1234567890/123 or t.me/channel/123"
                  required
                />
              </div>
              
              <div className="input-group">
                <label>Destination</label>
                <input
                  type="text"
                  value={destination}
                  onChange={(e) => setDestination(e.target.value)}
                  placeholder="~/Downloads"
                />
              </div>
              
              <button 
                type="submit" 
                className="btn-primary btn-download"
                disabled={loading}
              >
                {loading ? 'Starting...' : 'Download'}
              </button>
            </div>
          </form>
        </section>

        <section className="downloads-section">
          <h2>Active Downloads</h2>
          
          {downloads.length === 0 ? (
            <div className="empty-state">
              <p>No active downloads</p>
              <p className="hint">Paste a Telegram link above to start downloading</p>
            </div>
          ) : (
            <div className="downloads-list">
              {downloads.map((download) => (
                <DownloadItem
                  key={download.id}
                  download={download}
                  onCancel={() => handleCancel(download.id)}
                  formatSpeed={formatSpeed}
                  formatSize={formatSize}
                />
              ))}
            </div>
          )}
        </section>

        <section className="info-section">
          <div className="info-card">
            <h3>How it works</h3>
            <ul>
              <li>Paste any Telegram file link (private or public channels)</li>
              <li>Files download with parallel connections for maximum speed</li>
              <li>Supports videos, photos, documents, and any media</li>
              <li>Automatically resumes interrupted downloads</li>
            </ul>
          </div>
          
          <div className="info-card">
            <h3>Supported Links</h3>
            <ul>
              <li><code>https://t.me/c/CHANNEL_ID/MESSAGE_ID</code> (Private)</li>
              <li><code>https://t.me/USERNAME/MESSAGE_ID</code> (Public)</li>
              <li><code>t.me/c/CHANNEL_ID/MESSAGE_ID</code> (Short)</li>
            </ul>
          </div>
        </section>
      </main>
    </div>
  )
}

export default Dashboard
