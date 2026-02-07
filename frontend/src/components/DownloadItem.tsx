interface DownloadInfo {
  id: string
  filename: string
  progress: number
  downloaded: number
  total: number
  speed: number
  status: string
}

interface DownloadItemProps {
  download: DownloadInfo
  onCancel: () => void
  formatSpeed: (bytes: number) => string
  formatSize: (bytes: number) => string
}

function DownloadItem({ download, onCancel, formatSpeed, formatSize }: DownloadItemProps) {
  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'downloading':
        return '#3498db'
      case 'completed':
        return '#2ecc71'
      case 'error':
        return '#e74c3c'
      case 'cancelled':
        return '#95a5a6'
      default:
        return '#7f8c8d'
    }
  }

  const getStatusText = (status: string): string => {
    switch (status) {
      case 'pending':
        return 'Pending...'
      case 'downloading':
        return 'Downloading'
      case 'completed':
        return 'Completed'
      case 'error':
        return 'Error'
      case 'cancelled':
        return 'Cancelled'
      default:
        return status
    }
  }

  return (
    <div className="download-item">
      <div className="download-info">
        <div className="download-filename">{download.filename || 'Unknown File'}</div>
        <div className="download-meta">
          <span 
            className="status-badge"
            style={{ backgroundColor: getStatusColor(download.status) }}
          >
            {getStatusText(download.status)}
          </span>
          <span className="size-info">
            {formatSize(download.downloaded)} / {formatSize(download.total)}
          </span>
          {download.status === 'downloading' && (
            <span className="speed-info">{formatSpeed(download.speed)}</span>
          )}
        </div>
      </div>

      <div className="download-progress">
        <div className="progress-bar">
          <div 
            className="progress-fill"
            style={{ 
              width: `${download.progress}%`,
              backgroundColor: getStatusColor(download.status)
            }}
          />
        </div>
        <span className="progress-text">{download.progress.toFixed(1)}%</span>
      </div>

      {download.status === 'downloading' && (
        <button 
          className="btn-cancel"
          onClick={onCancel}
          title="Cancel download"
        >
          ✕
        </button>
      )}
    </div>
  )
}

export default DownloadItem
