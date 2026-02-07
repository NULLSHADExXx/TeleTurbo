package main

import (
	"context"
	"fmt"
	"runtime"

	"TeleTurbo/internal/telegram"
)

// App struct
type App struct {
	ctx      context.Context
	tgClient *telegram.TGClient
	downloads map[string]*telegram.DownloadTask
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		downloads: make(map[string]*telegram.DownloadTask),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// domReady is called after front-end resources have been loaded
func (a App) domReady(ctx context.Context) {
	// Add your action here
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	// Perform your teardown here
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// InitializeTelegramClient creates a new Telegram client
func (a *App) InitializeTelegramClient(appID int32, appHash string) string {
	client, err := telegram.NewClient(appID, appHash)
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}
	a.tgClient = client
	return "CLIENT_INITIALIZED"
}

// StartLogin initiates the phone authentication flow
func (a *App) StartLogin(phone string) string {
	if a.tgClient == nil {
		return "ERROR: Client not initialized"
	}
	return a.tgClient.StartLogin(phone)
}

// SubmitCode submits the verification code
func (a *App) SubmitCode(code string) string {
	if a.tgClient == nil {
		return "ERROR: Client not initialized"
	}
	return a.tgClient.SubmitCode(code)
}

// SubmitPassword submits the 2FA password
func (a *App) SubmitPassword(password string) string {
	if a.tgClient == nil {
		return "ERROR: Client not initialized"
	}
	return a.tgClient.SubmitPassword(password)
}

// IsAuthenticated checks if user is logged in
func (a *App) IsAuthenticated() bool {
	if a.tgClient == nil {
		return false
	}
	return a.tgClient.IsAuthenticated()
}

// GetSystemInfo returns system information for download optimization
func (a *App) GetSystemInfo() map[string]interface{} {
	return map[string]interface{}{
		"cpu_cores":    runtime.NumCPU(),
		"parallelism":  runtime.NumCPU() * 2,
		"os":           runtime.GOOS,
		"arch":         runtime.GOARCH,
	}
}

// StartDownload initiates a parallel download
func (a *App) StartDownload(messageLink string, destination string) string {
	if a.tgClient == nil {
		return "ERROR: Client not initialized"
	}
	
	task := a.tgClient.DownloadFile(messageLink, destination)
	if task == nil {
		return "ERROR: Failed to create download task"
	}
	
	a.downloads[task.ID] = task
	return task.ID
}

// GetDownloadProgress returns the current progress of a download
func (a *App) GetDownloadProgress(downloadID string) map[string]interface{} {
	task, exists := a.downloads[downloadID]
	if !exists {
		return map[string]interface{}{
			"error": "Download not found",
		}
	}
	
	return map[string]interface{}{
		"id":           task.ID,
		"progress":     task.GetProgress(),
		"downloaded":   task.DownloadedBytes,
		"total":        task.TotalBytes,
		"speed":        task.GetSpeed(),
		"status":       task.Status,
		"filename":     task.Filename,
	}
}

// GetAllDownloads returns all active downloads
func (a *App) GetAllDownloads() []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	for _, task := range a.downloads {
		result = append(result, map[string]interface{}{
			"id":         task.ID,
			"progress":   task.GetProgress(),
			"downloaded": task.DownloadedBytes,
			"total":      task.TotalBytes,
			"speed":      task.GetSpeed(),
			"status":     task.Status,
			"filename":   task.Filename,
		})
	}
	return result
}

// CancelDownload cancels an active download
func (a *App) CancelDownload(downloadID string) string {
	task, exists := a.downloads[downloadID]
	if !exists {
		return "ERROR: Download not found"
	}
	
	task.Cancel()
	delete(a.downloads, downloadID)
	return "CANCELLED"
}
