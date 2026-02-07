package telegram

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
)

// DownloadTask represents an active download
type DownloadTask struct {
	ID              string
	MessageLink     string
	Destination     string
	Filename        string
	TotalBytes      int64
	DownloadedBytes int64
	Status          string // pending, downloading, completed, error, cancelled
	Error           string
	StartTime       time.Time
	EndTime         time.Time
	
	// Internal
	client      *TGClient
	ctx         context.Context
	cancelFunc  context.CancelFunc
	mu          sync.RWMutex
	speedSamples  []speedSample
}

type speedSample struct {
	bytes int64
	time  time.Time
}

// DownloadFile initiates a high-speed parallel download
func (t *TGClient) DownloadFile(messageLink, destination string) *DownloadTask {
	taskCtx, cancel := context.WithCancel(t.runCtx)

	task := &DownloadTask{
		ID:          generateRandomID(),
		MessageLink: messageLink,
		Destination: destination,
		Status:      "pending",
		client:      t,
		ctx:         taskCtx,
		cancelFunc:  cancel,
		StartTime:   time.Now(),
	}

	// Start download in background
	go task.execute()

	return task
}

// execute performs the actual download with parallel chunking
func (d *DownloadTask) execute() {
	d.setStatus("downloading")

	// Start speed tracking
	go d.startSpeedTracker()

	// Parse the link
	linkInfo, err := ParseTelegramLink(d.MessageLink)
	if err != nil {
		d.setError(fmt.Sprintf("Failed to parse link: %v", err))
		return
	}

	fmt.Printf("Parsed link: channelID=%d, username=%s, messageID=%d, private=%v\n",
		linkInfo.ChannelID, linkInfo.Username, linkInfo.MessageID, linkInfo.IsPrivate)

	// Resolve channel peer
	var channelPeer *tg.InputPeerChannel
	resolveCtx, resolveCancel := context.WithTimeout(d.ctx, 30*time.Second)
	defer resolveCancel()

	if linkInfo.IsPrivate {
		channelPeer, err = d.client.GetChannelPeer(resolveCtx, linkInfo.ChannelID)
	} else {
		channelPeer, err = d.client.ResolveUsername(resolveCtx, linkInfo.Username)
	}
	if err != nil {
		d.setError(fmt.Sprintf("Failed to resolve channel: %v", err))
		return
	}

	fmt.Printf("Resolved channel: ID=%d, AccessHash=%d\n", channelPeer.ChannelID, channelPeer.AccessHash)

	// Fetch message to get file location
	fileLocation, filename, size, err := d.resolveFileLocation(channelPeer, linkInfo.MessageID)
	if err != nil {
		d.setError(fmt.Sprintf("Failed to resolve file: %v", err))
		return
	}

	d.Filename = filename
	d.TotalBytes = size
	fmt.Printf("File: %s, Size: %d bytes\n", filename, size)

	// Ensure destination directory exists
	destPath := d.Destination
	if destPath == "" {
		destPath = "~/Downloads/TeleTurbo"
	}
	// Expand ~ to home dir
	if len(destPath) > 0 && destPath[0] == '~' {
		home, _ := os.UserHomeDir()
		destPath = filepath.Join(home, destPath[1:])
	}

	if err := os.MkdirAll(destPath, 0755); err != nil {
		d.setError(fmt.Sprintf("Failed to create destination: %v", err))
		return
	}

	// Full file path
	filePath := filepath.Join(destPath, filename)

	// Check if file already exists with same size
	if info, err := os.Stat(filePath); err == nil {
		if info.Size() == size {
			atomic.StoreInt64(&d.DownloadedBytes, size)
			d.setStatus("completed")
			d.EndTime = time.Now()
			fmt.Printf("File already exists: %s\n", filePath)
			return
		}
	}

	// Configure parallel download
	parallelism := runtime.NumCPU() * 2
	if parallelism < 4 {
		parallelism = 4
	}
	if parallelism > 16 {
		parallelism = 16
	}

	// Create downloader
	dl := downloader.NewDownloader()

	// Create progress writer
	file, err := os.Create(filePath)
	if err != nil {
		d.setError(fmt.Sprintf("Failed to create file: %v", err))
		return
	}
	defer file.Close()

	// Wrap file with progress tracking writer
	progressWriter := &progressWriter{
		writer:   file,
		task:     d,
	}

	fmt.Printf("Starting download with %d threads...\n", parallelism)

	_, err = dl.Download(d.client.GetClient().API(), fileLocation).
		WithThreads(parallelism).
		Stream(d.ctx, progressWriter)

	if err != nil {
		file.Close()
		if d.ctx.Err() == context.Canceled {
			d.setStatus("cancelled")
			os.Remove(filePath)
		} else {
			d.setError(fmt.Sprintf("Download failed: %v", err))
		}
		return
	}

	// Mark as completed
	atomic.StoreInt64(&d.DownloadedBytes, d.TotalBytes)
	d.setStatus("completed")
	d.EndTime = time.Now()
	fmt.Printf("Download completed: %s\n", filePath)
}

// progressWriter wraps an io.Writer to track bytes written
type progressWriter struct {
	writer io.Writer
	task   *DownloadTask
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if n > 0 {
		atomic.AddInt64(&pw.task.DownloadedBytes, int64(n))
	}
	return n, err
}

// resolveFileLocation fetches message and extracts file location
func (d *DownloadTask) resolveFileLocation(peer *tg.InputPeerChannel, messageID int) (tg.InputFileLocationClass, string, int64, error) {
	ctx, cancel := context.WithTimeout(d.ctx, 30*time.Second)
	defer cancel()

	api := d.client.GetClient().API()

	// Fetch message using the resolved channel with proper access hash
	messages, err := api.ChannelsGetMessages(ctx, &tg.ChannelsGetMessagesRequest{
		Channel: &tg.InputChannel{
			ChannelID:  peer.ChannelID,
			AccessHash: peer.AccessHash,
		},
		ID: []tg.InputMessageClass{&tg.InputMessageID{ID: messageID}},
	})

	if err != nil {
		return nil, "", 0, fmt.Errorf("failed to fetch message: %w", err)
	}

	// Extract file info from message
	var msgList []tg.MessageClass
	switch m := messages.(type) {
	case *tg.MessagesChannelMessages:
		msgList = m.Messages
	case *tg.MessagesMessages:
		msgList = m.Messages
	case *tg.MessagesMessagesSlice:
		msgList = m.Messages
	}

	if len(msgList) == 0 {
		return nil, "", 0, fmt.Errorf("no messages found")
	}

	msg := msgList[0]
	var media tg.MessageMediaClass

	switch m := msg.(type) {
	case *tg.Message:
		if m.Media == nil {
			return nil, "", 0, fmt.Errorf("message has no media")
		}
		media = m.Media
	default:
		return nil, "", 0, fmt.Errorf("unsupported message type: %T", msg)
	}

	// Extract file location and info based on media type
	return d.extractFileInfo(media)
}

// extractFileInfo extracts file location from media
func (d *DownloadTask) extractFileInfo(media tg.MessageMediaClass) (tg.InputFileLocationClass, string, int64, error) {
	switch m := media.(type) {
	case *tg.MessageMediaDocument:
		doc, ok := m.Document.(*tg.Document)
		if !ok {
			return nil, "", 0, fmt.Errorf("invalid document")
		}
		
		// Get filename from attributes
		filename := "download.bin"
		for _, attr := range doc.Attributes {
			if fileAttr, ok := attr.(*tg.DocumentAttributeFilename); ok {
				filename = fileAttr.FileName
				break
			}
		}
		
		location := &tg.InputDocumentFileLocation{
			ID:            doc.ID,
			AccessHash:    doc.AccessHash,
			FileReference: doc.FileReference,
		}
		
		return location, filename, doc.Size, nil
		
	case *tg.MessageMediaPhoto:
		photo, ok := m.Photo.(*tg.Photo)
		if !ok {
			return nil, "", 0, fmt.Errorf("invalid photo")
		}
		
		// Get largest photo size
		var largest *tg.PhotoSize
		var maxSize int64
		
		for _, size := range photo.Sizes {
			if ps, ok := size.(*tg.PhotoSize); ok {
				if int64(ps.Size) > maxSize {
					maxSize = int64(ps.Size)
					largest = ps
				}
			}
		}
		
		if largest == nil {
			return nil, "", 0, fmt.Errorf("no photo sizes found")
		}
		
		filename := fmt.Sprintf("photo_%d.jpg", photo.ID)
		location := &tg.InputPhotoFileLocation{
			ID:            photo.ID,
			AccessHash:    photo.AccessHash,
			FileReference: photo.FileReference,
			ThumbSize:     largest.Type,
		}
		
		return location, filename, maxSize, nil
		
	default:
		return nil, "", 0, fmt.Errorf("unsupported media type: %T", media)
	}
}

// startSpeedTracker monitors download speed in the background
func (d *DownloadTask) startSpeedTracker() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var lastBytes int64

	for {
		select {
		case <-ticker.C:
			currentBytes := atomic.LoadInt64(&d.DownloadedBytes)
			now := time.Now()

			d.mu.Lock()
			d.speedSamples = append(d.speedSamples, speedSample{
				bytes: currentBytes - lastBytes,
				time:  now,
			})

			// Keep only last 10 samples
			if len(d.speedSamples) > 10 {
				d.speedSamples = d.speedSamples[len(d.speedSamples)-10:]
			}
			d.mu.Unlock()

			lastBytes = currentBytes

		case <-d.ctx.Done():
			return
		}
	}
}

// GetProgress returns download progress percentage
func (d *DownloadTask) GetProgress() float64 {
	if d.TotalBytes == 0 {
		return 0
	}
	
	downloaded := atomic.LoadInt64(&d.DownloadedBytes)
	return float64(downloaded) / float64(d.TotalBytes) * 100
}

// GetSpeed returns current download speed in bytes/second
func (d *DownloadTask) GetSpeed() float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	if len(d.speedSamples) < 2 {
		return 0
	}
	
	var totalBytes int64
	var totalTime time.Duration
	
	for i := 1; i < len(d.speedSamples); i++ {
		totalBytes += d.speedSamples[i].bytes
		totalTime += d.speedSamples[i].time.Sub(d.speedSamples[i-1].time)
	}
	
	if totalTime == 0 {
		return 0
	}
	
	return float64(totalBytes) / totalTime.Seconds()
}

// GetETA returns estimated time to completion
func (d *DownloadTask) GetETA() time.Duration {
	speed := d.GetSpeed()
	if speed == 0 {
		return 0
	}
	
	remaining := d.TotalBytes - atomic.LoadInt64(&d.DownloadedBytes)
	seconds := float64(remaining) / speed
	
	return time.Duration(seconds) * time.Second
}

// Cancel stops the download
func (d *DownloadTask) Cancel() {
	d.cancelFunc()
	d.setStatus("cancelled")
}

// Internal state management
func (d *DownloadTask) setStatus(status string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Status = status
}

func (d *DownloadTask) setError(err string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Status = "error"
	d.Error = err
}

// GetStatus returns current status
func (d *DownloadTask) GetStatus() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Status
}

// GetError returns error message if any
func (d *DownloadTask) GetError() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Error
}

// Helper function to copy with progress
func copyWithProgress(dst io.Writer, src io.Reader, total int64, progress chan<- int64) (int64, error) {
	var written int64
	buf := make([]byte, 32*1024) // 32KB buffer
	
	for {
		nr, err := src.Read(buf)
		if nr > 0 {
			nw, err := dst.Write(buf[:nr])
			if nw > 0 {
				written += int64(nw)
				select {
				case progress <- written:
				default:
				}
			}
			if err != nil {
				return written, err
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if err != nil {
			if err == io.EOF {
				return written, nil
			}
			return written, err
		}
	}
}
