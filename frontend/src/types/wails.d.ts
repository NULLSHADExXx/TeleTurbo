// Global Wails runtime bindings
declare global {
  interface Window {
    go: {
      main: {
        App: {
          // Auth methods
          IsAuthenticated(): Promise<boolean>
          InitializeTelegramClient(appID: number, appHash: string): Promise<string>
          StartLogin(phone: string): Promise<string>
          SubmitCode(code: string): Promise<string>
          SubmitPassword(password: string): Promise<string>
          
          // System info
          GetSystemInfo(): Promise<Record<string, any>>
          
          // Download methods
          StartDownload(link: string, destination: string): Promise<string>
          GetAllDownloads(): Promise<Array<{
            id: string
            filename: string
            progress: number
            downloaded: number
            total: number
            speed: number
            status: string
          }>>
          GetDownloadProgress(id: string): Promise<{
            id: string
            filename: string
            progress: number
            downloaded: number
            total: number
            speed: number
            status: string
          }>
          CancelDownload(id: string): Promise<string>
        }
      }
    }
  }
}

export {}
