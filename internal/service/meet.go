package service

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/hnam/notafly/internal/config"
	"go.uber.org/zap"
)

// Selectors — extracted so they're easy to update when Google changes UI.
var selectors = struct {
	EmailInput   string
	EmailNext    string
	PasswordInput string
	PasswordNext string
	MicToggle    string
	CameraToggle string
	JoinButton   string
}{
	EmailInput:   `#identifierId`,
	EmailNext:    `#identifierNext`,
	PasswordInput: `#password input[type="password"]`,
	PasswordNext: `#passwordNext`,
	MicToggle:    `div[data-anchor-id="hw0c9"]`,
	CameraToggle: `div[data-anchor-id="psRWwc"]`,
	JoinButton:   `button[jsname="Qx7uuf"]`,
}

const (
	loginURL = "https://accounts.google.com/ServiceLogin?hl=en&passive=true&continue=https://www.google.com/&ec=GAZAAQ"
)

type MeetService struct {
	config *config.Config
	logger *zap.Logger
}

func NewMeetService(cfg *config.Config, logger *zap.Logger) *MeetService {
	return &MeetService{config: cfg, logger: logger}
}

// JoinMeeting logs into Gmail, navigates to the Meet link, disables mic/cam,
// and clicks "Ask to Join". It returns the browser context so the caller can
// keep it alive during recording (cancel to close Chrome).
func (s *MeetService) JoinMeeting(ctx context.Context, meetLink string) (context.Context, context.CancelFunc, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("start-maximized", true),
		// Allow mic/camera access without prompt
		chromedp.Flag("use-fake-ui-for-media-stream", true),
		// Non-headless needed for media permissions on real devices
		chromedp.Flag("headless", false),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)

	browserCtx, browserCancel := chromedp.NewContext(allocCtx,
		chromedp.WithLogf(func(format string, args ...interface{}) {
			s.logger.Debug(fmt.Sprintf(format, args...))
		}),
	)

	// Combined cancel to clean up both browser and allocator
	cancelAll := func() {
		browserCancel()
		allocCancel()
	}

	// Step 1: Login
	s.logger.Info("logging into Gmail")
	if err := s.login(browserCtx); err != nil {
		cancelAll()
		return nil, nil, fmt.Errorf("gmail login failed: %w", err)
	}

	// Step 2: Navigate, disable media, join
	s.logger.Info("joining meeting", zap.String("link", meetLink))
	if err := s.disableMediaAndJoin(browserCtx, meetLink); err != nil {
		cancelAll()
		return nil, nil, fmt.Errorf("join meeting failed: %w", err)
	}

	s.logger.Info("successfully joined meeting")
	return browserCtx, cancelAll, nil
}

func (s *MeetService) login(ctx context.Context) error {
	return chromedp.Run(ctx,
		// Navigate to login page
		chromedp.Navigate(loginURL),

		// Enter email
		chromedp.WaitVisible(selectors.EmailInput, chromedp.ByQuery),
		chromedp.SendKeys(selectors.EmailInput, s.config.EmailID, chromedp.ByQuery),
		chromedp.Click(selectors.EmailNext, chromedp.ByQuery),

		// Wait for password field and enter password
		chromedp.WaitVisible(selectors.PasswordInput, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second), // Google animates the transition
		chromedp.SendKeys(selectors.PasswordInput, s.config.EmailPassword, chromedp.ByQuery),
		chromedp.Click(selectors.PasswordNext, chromedp.ByQuery),

		// Wait for redirect to Google home
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	)
}

func (s *MeetService) disableMediaAndJoin(ctx context.Context, meetLink string) error {
	return chromedp.Run(ctx,
		// Navigate to Meet link
		chromedp.Navigate(meetLink),
		chromedp.Sleep(3*time.Second), // Meet lobby takes time to load

		// Disable microphone
		chromedp.WaitVisible(selectors.MicToggle, chromedp.ByQuery),
		chromedp.Click(selectors.MicToggle, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),

		// Disable camera
		chromedp.WaitVisible(selectors.CameraToggle, chromedp.ByQuery),
		chromedp.Click(selectors.CameraToggle, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),

		// Click "Ask to Join" / "Join now"
		chromedp.WaitVisible(selectors.JoinButton, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		chromedp.Click(selectors.JoinButton, chromedp.ByQuery),
	)
}
