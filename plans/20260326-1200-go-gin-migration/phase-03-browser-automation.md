# Phase 3: Browser Automation (chromedp)

**Priority:** P0
**Status:** ⬜ Not started

---

## Context

Current Python uses Selenium WebDriver to:
1. Login to Gmail (`Glogin`)
2. Navigate to Meet link, disable mic/camera (`turnOffMicCam`)
3. Click "Ask to Join" button (`AskToJoin`)

### Current Python Selectors (fragile — Google can change anytime)

| Action | Selector |
|--------|----------|
| Email input | `By.ID, "identifierId"` |
| Next button | `By.ID, "identifierNext"` |
| Password input | `By.XPATH, '//*[@id="password"]/div[1]/div/div[1]/input'` |
| Password next | `By.ID, "passwordNext"` |
| Mic toggle | `div[jscontroller="t2mBxb"][data-anchor-id="hw0c9"]` |
| Camera toggle | `div[jscontroller="bwqwSd"][data-anchor-id="psRWwc"]` |
| Join button | `button[jsname="Qx7uuf"]` |

## Go Implementation: chromedp

```go
// internal/service/meet.go
type MeetService struct {
    config *config.Config
    logger *zap.Logger
}

func (s *MeetService) JoinMeeting(ctx context.Context, meetLink string) error {
    opts := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.Flag("disable-blink-features", "AutomationControlled"),
        chromedp.Flag("start-maximized", true),
    )

    allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
    defer cancel()

    ctx, cancel = chromedp.NewContext(allocCtx)
    defer cancel()

    // 1. Login to Gmail
    if err := s.login(ctx); err != nil {
        return fmt.Errorf("gmail login failed: %w", err)
    }

    // 2. Navigate & disable mic/cam
    if err := s.disableMediaAndJoin(ctx, meetLink); err != nil {
        return fmt.Errorf("join meeting failed: %w", err)
    }

    return nil
}
```

### Key Differences from Python

| Aspect | Python (Selenium) | Go (chromedp) |
|--------|-------------------|---------------|
| Driver | Needs ChromeDriver binary | Uses Chrome DevTools Protocol directly |
| Waits | `implicitly_wait()`, `time.sleep()` | `chromedp.WaitVisible()`, `chromedp.WaitReady()` |
| Selectors | By.ID, By.CSS_SELECTOR, By.XPATH | `chromedp.ByID`, `chromedp.ByQuery`, `chromedp.BySearch` |
| Error handling | try/except | error returns |
| Headless | opt.add_argument('--headless') | `chromedp.Headless` flag |

## Implementation Steps

- [ ] Create `internal/service/meet.go` with `MeetService` struct
- [ ] Implement `login(ctx)` — Gmail auth flow via chromedp
- [ ] Implement `disableMediaAndJoin(ctx, meetLink)` — mic/cam off + join
- [ ] Add configurable selectors (in case Google changes UI)
- [ ] Add proper waits (replace `time.sleep` with `WaitVisible`)
- [ ] Add context timeout for each operation
- [ ] Handle common errors: element not found, timeout, auth failure

## Risk: chromedp vs Google Meet

Google Meet is heavy JS SPA. chromedp works well but:
- Must run with **headed mode** (not headless) for media permissions
- Chrome must have real audio device access for recording
- Fallback option: use `go-rod/rod` which has higher-level API

## Success Criteria

- Bot can login to Gmail via chromedp
- Bot can join a Google Meet link
- Mic and camera are disabled before joining
- Proper error messages on failure
