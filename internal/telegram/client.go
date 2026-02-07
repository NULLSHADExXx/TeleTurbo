package telegram

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
)

// TGClient wraps the Telegram client with session management
type TGClient struct {
	client         *telegram.Client
	ctx            context.Context
	cancel         context.CancelFunc
	runCtx         context.Context
	appID          int32
	appHash        string
	authenticated  bool
	authMutex      sync.RWMutex
	sessionStorage *telegram.FileSessionStorage
	ready          chan struct{}

	// Auth flow state
	phoneCodeHash  string
	phoneNumber    string
	authFlow       chan string
}

// NewClient creates a new Telegram client with session persistence
func NewClient(appID int32, appHash string) (*TGClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Setup session storage in user's home directory
	sessionDir := filepath.Join(os.TempDir(), ".teleturbo")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}
	sessionPath := filepath.Join(sessionDir, "session.json")
	
	// Clear any corrupted session
	os.Remove(sessionPath)
	
	sessionStorage := &telegram.FileSessionStorage{Path: sessionPath}
	
	// Create client (appID should be int, not int32)
	client := telegram.NewClient(int(appID), appHash, telegram.Options{
		SessionStorage: sessionStorage,
	})
	
	tgClient := &TGClient{
		client:         client,
		ctx:            ctx,
		cancel:         cancel,
		appID:          appID,
		appHash:        appHash,
		sessionStorage: sessionStorage,
		authFlow:       make(chan string, 1),
		ready:          make(chan struct{}),
	}

	errCh := make(chan error, 1)

	// Start client in background
	go func() {
		if err := client.Run(ctx, func(runCtx context.Context) error {
			// Store the run context so API calls use it
			tgClient.runCtx = runCtx

			// Check if already authenticated
			status, err := client.Auth().Status(runCtx)
			if err != nil {
				fmt.Printf("Auth status error: %v\n", err)
			} else if status.Authorized {
				tgClient.setAuthenticated(true)
			}

			// Signal that client is ready for API calls
			close(tgClient.ready)

			// Keep client running
			<-runCtx.Done()
			return runCtx.Err()
		}); err != nil {
			fmt.Printf("Client error: %v\n", err)
			errCh <- err
		}
	}()

	// Wait for client to be ready or fail
	select {
	case <-tgClient.ready:
		return tgClient, nil
	case err := <-errCh:
		cancel()
		return nil, fmt.Errorf("client failed to start: %w", err)
	case <-time.After(10 * time.Second):
		cancel()
		return nil, fmt.Errorf("client timed out connecting to Telegram")
	}
}

// StartLogin initiates phone authentication
func (t *TGClient) StartLogin(phone string) string {
	// Format phone number (remove spaces, ensure + prefix)
	phone = strings.TrimSpace(phone)
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")

	if !strings.HasPrefix(phone, "+") {
		return "ERROR: Phone number must start with + and country code (e.g., +1234567890)"
	}

	t.phoneNumber = phone

	ctx, cancel := context.WithTimeout(t.runCtx, 30*time.Second)
	defer cancel()

	fmt.Printf("Sending auth code to %s...\n", phone)

	// Request phone code
	result, err := t.client.API().AuthSendCode(ctx, &tg.AuthSendCodeRequest{
		PhoneNumber: phone,
		APIID:       int(t.appID),
		APIHash:     t.appHash,
		Settings:    tg.CodeSettings{},
	})

	if err != nil {
		fmt.Printf("AuthSendCode error: %v\n", err)
		return fmt.Sprintf("ERROR: %v", err)
	}

	// Store phone code hash
	switch sentCode := result.(type) {
	case *tg.AuthSentCode:
		t.phoneCodeHash = sentCode.PhoneCodeHash
		fmt.Printf("Code sent successfully, hash: %s\n", sentCode.PhoneCodeHash)
		return "CODE_SENT"
	case *tg.AuthSentCodeSuccess:
		t.setAuthenticated(true)
		return "LOGIN_SUCCESS"
	default:
		return fmt.Sprintf("ERROR: Unexpected response type: %T", result)
	}
}

// SubmitCode submits the verification code
func (t *TGClient) SubmitCode(code string) string {
	if t.phoneCodeHash == "" {
		return "ERROR: No active login flow"
	}
	
	ctx, cancel := context.WithTimeout(t.runCtx, 30*time.Second)
	defer cancel()

	// Sign in with code
	result, err := t.client.API().AuthSignIn(ctx, &tg.AuthSignInRequest{
		PhoneNumber:   t.phoneNumber,
		PhoneCodeHash: t.phoneCodeHash,
		PhoneCode:     code,
	})
	
	if err != nil {
		// Check if cloud password is required
		if strings.Contains(err.Error(), "SESSION_PASSWORD_NEEDED") {
			return "PASSWORD_REQUIRED"
		}
		fmt.Printf("AuthSignIn error: %v\n", err)
		return fmt.Sprintf("ERROR: %v", err)
	}
	
	switch result.(type) {
	case *tg.AuthAuthorization:
		t.setAuthenticated(true)
		return "LOGIN_SUCCESS"
	case *tg.AuthAuthorizationSignUpRequired:
		return "SIGNUP_REQUIRED"
	default:
		return "ERROR: Unexpected response"
	}
}

// SubmitPassword submits cloud password
func (t *TGClient) SubmitPassword(password string) string {
	ctx, cancel := context.WithTimeout(t.runCtx, 30*time.Second)
	defer cancel()

	// Get password configuration
	passwordConfig, err := t.client.API().AccountGetPassword(ctx)
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}

	// Check if password is actually needed
	if passwordConfig.CurrentAlgo == nil {
		return "ERROR: No password needed - 2FA not enabled on this account"
	}

	// Generate secure random bytes
	secureRandom := make([]byte, 32)
	if _, err := rand.Read(secureRandom); err != nil {
		return fmt.Sprintf("ERROR: failed to generate random: %v", err)
	}

	// Use gotd's proper PasswordHash function
	srpHash, err := auth.PasswordHash(
		[]byte(password),
		passwordConfig.SRPID,
		passwordConfig.SRPB,
		secureRandom,
		passwordConfig.CurrentAlgo,
	)
	if err != nil {
		return fmt.Sprintf("ERROR: failed to compute password hash: %v", err)
	}

	// Submit password
	result, err := t.client.API().AuthCheckPassword(ctx, srpHash)
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}

	switch result.(type) {
	case *tg.AuthAuthorization:
		t.setAuthenticated(true)
		return "LOGIN_SUCCESS"
	case *tg.AuthAuthorizationSignUpRequired:
		return "SIGNUP_REQUIRED"
	default:
		return "ERROR: Unexpected response from password check"
	}
}

// IsAuthenticated returns authentication status
func (t *TGClient) IsAuthenticated() bool {
	t.authMutex.RLock()
	defer t.authMutex.RUnlock()
	return t.authenticated
}

func (t *TGClient) setAuthenticated(value bool) {
	t.authMutex.Lock()
	defer t.authMutex.Unlock()
	t.authenticated = value
}

// Logout terminates the session
func (t *TGClient) Logout() error {
	ctx, cancel := context.WithTimeout(t.ctx, 10*time.Second)
	defer cancel()
	
	_, err := t.client.API().AuthLogOut(ctx)
	if err != nil {
		return err
	}
	
	t.setAuthenticated(false)
	
	// Clear session file
	sessionDir := filepath.Join(os.TempDir(), ".teleturbo")
	sessionPath := filepath.Join(sessionDir, "session.json")
	os.Remove(sessionPath)
	
	return nil
}

// LinkInfo holds parsed Telegram link information
type LinkInfo struct {
	ChannelID int64
	Username  string
	MessageID int
	IsPrivate bool
}

// ParseTelegramLink extracts channel/username and message ID from various link formats
func ParseTelegramLink(link string) (*LinkInfo, error) {
	link = strings.TrimSpace(link)
	link = strings.ReplaceAll(link, "https://", "")
	link = strings.ReplaceAll(link, "http://", "")
	link = strings.TrimPrefix(link, "www.")

	// Private channel format: t.me/c/CHANNEL_ID/MESSAGE_ID
	if strings.Contains(link, "/c/") {
		parts := strings.Split(link, "/")
		for i, part := range parts {
			if part == "c" && i+2 < len(parts) {
				channelID, err := strconv.ParseInt(parts[i+1], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid channel ID: %w", err)
				}
				messageID, err := strconv.Atoi(parts[i+2])
				if err != nil {
					return nil, fmt.Errorf("invalid message ID: %w", err)
				}
				return &LinkInfo{
					ChannelID: channelID,
					MessageID: messageID,
					IsPrivate: true,
				}, nil
			}
		}
	}

	// Public channel format: t.me/USERNAME/MESSAGE_ID
	if strings.HasPrefix(link, "t.me/") {
		parts := strings.Split(link, "/")
		if len(parts) >= 3 {
			username := parts[1]
			messageID, err := strconv.Atoi(parts[2])
			if err != nil {
				return nil, fmt.Errorf("invalid message ID: %w", err)
			}
			return &LinkInfo{
				Username:  username,
				MessageID: messageID,
				IsPrivate: false,
			}, nil
		}
	}

	return nil, fmt.Errorf("unsupported link format. Use https://t.me/c/CHANNEL_ID/MSG_ID or https://t.me/USERNAME/MSG_ID")
}

// GetRunContext returns the run context for API calls
func (t *TGClient) GetRunContext() context.Context {
	return t.runCtx
}

// ResolveUsername resolves a username to a channel InputPeer
func (t *TGClient) ResolveUsername(ctx context.Context, username string) (*tg.InputPeerChannel, error) {
	resolved, err := t.client.API().ContactsResolveUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve username @%s: %w", username, err)
	}

	for _, chat := range resolved.Chats {
		switch c := chat.(type) {
		case *tg.Channel:
			return &tg.InputPeerChannel{
				ChannelID:  c.ID,
				AccessHash: c.AccessHash,
			}, nil
		}
	}

	return nil, fmt.Errorf("could not find channel for @%s", username)
}

// GetChannelPeer gets an InputPeerChannel for a private channel ID by fetching dialogs
func (t *TGClient) GetChannelPeer(ctx context.Context, channelID int64) (*tg.InputPeerChannel, error) {
	// Try to get the channel from the full dialog list
	result, err := t.client.API().MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      100,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get dialogs: %w", err)
	}

	var chats []tg.ChatClass
	switch d := result.(type) {
	case *tg.MessagesDialogs:
		chats = d.Chats
	case *tg.MessagesDialogsSlice:
		chats = d.Chats
	}

	for _, chat := range chats {
		switch c := chat.(type) {
		case *tg.Channel:
			if c.ID == channelID {
				return &tg.InputPeerChannel{
					ChannelID:  c.ID,
					AccessHash: c.AccessHash,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("channel %d not found in your dialogs — make sure you're a member of this channel", channelID)
}

// GetClient returns the underlying Telegram client
func (t *TGClient) GetClient() *telegram.Client {
	return t.client
}

// GetContext returns the client context
func (t *TGClient) GetContext() context.Context {
	return t.ctx
}

// GetDownloader creates a new downloader instance with optimized settings
func (t *TGClient) GetDownloader() *downloader.Downloader {
	return downloader.NewDownloader()
}

// generateRandomID generates a random unique ID
func generateRandomID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
