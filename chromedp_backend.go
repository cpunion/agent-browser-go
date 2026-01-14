package agentbrowser

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/storage"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

// BrowserManager manages the browser lifecycle and provides operations.
type ChromeDPBackend struct {
	allocCtx    context.Context
	allocCancel context.CancelFunc
	ctx         context.Context
	cancel      context.CancelFunc

	// Tab management
	targets     []target.ID
	activeTab   int
	tabContexts map[target.ID]context.Context
	tabCancels  map[target.ID]context.CancelFunc

	// Ref tracking
	refMap  RefMap
	refLock sync.RWMutex

	// State
	launched     atomic.Bool
	headless     bool
	viewport     *Viewport
	consoleLog   []ConsoleMessage
	pageErrors   []PageError
	consoleLock  sync.Mutex
	requests     []TrackedRequest
	requestsLock sync.Mutex

	// Screencast
	screencastCallback func(ScreencastFrame)
	screencastLock     sync.Mutex
}

// LaunchOptions configures browser launch.
type LaunchOptions struct {
	Headless       bool
	Viewport       *Viewport
	ExecutablePath string
	UserDataDir    string // Path to user data directory for persistent profiles
	CDPPort        int
	Headers        map[string]string
}

// NewBrowserManager creates a new browser manager.
func NewChromeDPBackend() *ChromeDPBackend {
	return &ChromeDPBackend{
		tabContexts: make(map[target.ID]context.Context),
		tabCancels:  make(map[target.ID]context.CancelFunc),
		refMap:      make(RefMap),
	}
}

// Launch starts the browser.
func (b *ChromeDPBackend) Launch(opts LaunchOptions) error {
	if b.launched.Load() {
		// Check if headless setting changed
		if b.headless != opts.Headless {
			// Need to relaunch with new settings
			b.Close()
		} else {
			return nil // Already launched with same settings
		}
	}

	// Build chromedp options
	chromedpOpts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.DisableGPU,
		// Anti-detection: hide automation flags
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("excludeSwitches", "enable-automation"),
	}

	// Optional: sandbox and shm flags (via environment variables)
	if os.Getenv("AGENT_BROWSER_NO_SANDBOX") == "1" {
		chromedpOpts = append(chromedpOpts, chromedp.NoSandbox)
	}
	if os.Getenv("AGENT_BROWSER_DISABLE_SHM") == "1" {
		chromedpOpts = append(chromedpOpts, chromedp.Flag("disable-dev-shm-usage", true))
	}

	if opts.Headless {
		chromedpOpts = append(chromedpOpts, chromedp.Headless)
	}

	if opts.ExecutablePath != "" {
		chromedpOpts = append(chromedpOpts, chromedp.ExecPath(opts.ExecutablePath))
	}

	if opts.UserDataDir != "" {
		chromedpOpts = append(chromedpOpts, chromedp.UserDataDir(opts.UserDataDir))
	}

	if opts.Viewport != nil {
		chromedpOpts = append(chromedpOpts,
			chromedp.WindowSize(opts.Viewport.Width, opts.Viewport.Height))
		b.viewport = opts.Viewport
	} else {
		// Default viewport
		chromedpOpts = append(chromedpOpts, chromedp.WindowSize(1280, 720))
		b.viewport = &Viewport{Width: 1280, Height: 720}
	}

	b.headless = opts.Headless

	// Create allocator
	b.allocCtx, b.allocCancel = chromedp.NewExecAllocator(
		context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:], chromedpOpts...)...,
	)

	// Create browser context
	b.ctx, b.cancel = chromedp.NewContext(b.allocCtx)

	// Run an empty action to start the browser
	if err := chromedp.Run(b.ctx); err != nil {
		b.Close()
		return fmt.Errorf("failed to launch browser: %w", err)
	}

	// Get initial target
	targets, err := chromedp.Targets(b.ctx)
	if err != nil {
		b.Close()
		return fmt.Errorf("failed to get targets: %w", err)
	}

	for _, t := range targets {
		if t.Type == "page" {
			b.targets = append(b.targets, t.TargetID)
			b.tabContexts[t.TargetID] = b.ctx
			b.tabCancels[t.TargetID] = b.cancel
			break
		}
	}

	b.launched.Store(true)
	return nil
}

// Close closes the browser.
func (b *ChromeDPBackend) Close() error {
	if !b.launched.Load() {
		return nil
	}

	// Close all tab contexts
	for _, cancel := range b.tabCancels {
		if cancel != nil {
			cancel()
		}
	}

	if b.cancel != nil {
		b.cancel()
	}
	if b.allocCancel != nil {
		b.allocCancel()
	}

	b.launched.Store(false)
	b.targets = nil
	b.tabContexts = make(map[target.ID]context.Context)
	b.tabCancels = make(map[target.ID]context.CancelFunc)
	b.refMap = make(RefMap)

	return nil
}

// IsLaunched returns whether the browser is launched.
func (b *ChromeDPBackend) IsLaunched() bool {
	return b.launched.Load()
}

// Context returns the current browser context.
func (b *ChromeDPBackend) Context() context.Context {
	if len(b.targets) == 0 || b.activeTab >= len(b.targets) {
		return b.ctx
	}
	tid := b.targets[b.activeTab]
	if ctx, ok := b.tabContexts[tid]; ok {
		return ctx
	}
	return b.ctx
}

// Navigate navigates to a URL.
func (b *ChromeDPBackend) Navigate(url string, waitUntil string) (string, string, error) {
	ctx := b.Context()

	var title string
	var currentURL string

	// Simple navigation - WaitReady waits for body to be ready
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Title(&title),
		chromedp.Location(&currentURL),
	)

	if err != nil {
		return "", "", err
	}

	return currentURL, title, nil
}

// Click clicks an element.
func (b *ChromeDPBackend) Click(selector string) error {
	ctx := b.Context()
	sel := b.resolveSelector(selector)
	return chromedp.Run(ctx, chromedp.Click(sel, chromedp.NodeVisible))
}

// Fill clears and fills an input.
func (b *ChromeDPBackend) Fill(selector, value string) error {
	ctx := b.Context()
	sel := b.resolveSelector(selector)
	return chromedp.Run(ctx,
		chromedp.Clear(sel),
		chromedp.SendKeys(sel, value),
	)
}

// Type types text into an element.
func (b *ChromeDPBackend) Type(selector, text string, delay int) error {
	ctx := b.Context()
	sel := b.resolveSelector(selector)

	if delay > 0 {
		// Type with delay between keystrokes not directly supported,
		// we'll type character by character
		if err := chromedp.Run(ctx, chromedp.Focus(sel)); err != nil {
			return err
		}
		for _, char := range text {
			if err := chromedp.Run(ctx, chromedp.SendKeys(sel, string(char))); err != nil {
				return err
			}
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
		return nil
	}

	return chromedp.Run(ctx, chromedp.SendKeys(sel, text))
}

// Press presses a key.
func (b *ChromeDPBackend) Press(key string, selector string) error {
	ctx := b.Context()
	if selector != "" {
		sel := b.resolveSelector(selector)
		return chromedp.Run(ctx,
			chromedp.Focus(sel),
			chromedp.KeyEvent(key),
		)
	}
	return chromedp.Run(ctx, chromedp.KeyEvent(key))
}

// Hover hovers over an element.
func (b *ChromeDPBackend) Hover(selector string) error {
	ctx := b.Context()
	sel := b.resolveSelector(selector)

	var x, y float64
	err := chromedp.Run(ctx,
		chromedp.ScrollIntoView(sel),
		chromedp.Evaluate(fmt.Sprintf(`
			(function() {
				const el = document.querySelector(%q);
				if (!el) return {x: 0, y: 0};
				const rect = el.getBoundingClientRect();
				return {x: rect.left + rect.width/2, y: rect.top + rect.height/2};
			})()
		`, sel), &struct {
			X *float64 `json:"x"`
			Y *float64 `json:"y"`
		}{&x, &y}),
	)
	if err != nil {
		return err
	}
	return chromedp.Run(ctx, chromedp.MouseClickXY(x, y, chromedp.ButtonNone))
}

// Screenshot takes a screenshot.
func (b *ChromeDPBackend) Screenshot(fullPage bool, selector string, quality int) ([]byte, error) {
	ctx := b.Context()

	var buf []byte
	var err error

	if selector != "" {
		sel := b.resolveSelector(selector)
		err = chromedp.Run(ctx, chromedp.Screenshot(sel, &buf))
	} else if fullPage {
		err = chromedp.Run(ctx, chromedp.FullScreenshot(&buf, quality))
	} else {
		err = chromedp.Run(ctx, chromedp.CaptureScreenshot(&buf))
	}

	return buf, err
}

// Evaluate runs JavaScript and returns the result.
func (b *ChromeDPBackend) Evaluate(script string) (interface{}, error) {
	ctx := b.Context()

	var result interface{}
	err := chromedp.Run(ctx, chromedp.Evaluate(script, &result))
	return result, err
}

// GetText gets element text content.
func (b *ChromeDPBackend) GetText(selector string) (string, error) {
	ctx := b.Context()
	sel := b.resolveSelector(selector)

	var text string
	err := chromedp.Run(ctx, chromedp.Text(sel, &text))
	return text, err
}

// GetAttribute gets an element attribute.
func (b *ChromeDPBackend) GetAttribute(selector, attr string) (string, error) {
	ctx := b.Context()
	sel := b.resolveSelector(selector)

	var value string
	var ok bool
	err := chromedp.Run(ctx, chromedp.AttributeValue(sel, attr, &value, &ok))
	if err != nil {
		return "", err
	}
	if !ok {
		return "", nil
	}
	return value, nil
}

// GetHTML gets element HTML.
func (b *ChromeDPBackend) GetHTML(selector string, outer bool) (string, error) {
	ctx := b.Context()
	sel := b.resolveSelector(selector)

	var html string
	if outer {
		err := chromedp.Run(ctx, chromedp.OuterHTML(sel, &html))
		return html, err
	}
	err := chromedp.Run(ctx, chromedp.InnerHTML(sel, &html))
	return html, err
}

// IsVisible checks if element is visible.
func (b *ChromeDPBackend) IsVisible(selector string) (bool, error) {
	ctx := b.Context()
	sel := b.resolveSelector(selector)

	var visible bool
	err := chromedp.Run(ctx, chromedp.Evaluate(fmt.Sprintf(`
		(function() {
			const el = document.querySelector(%q);
			if (!el) return false;
			const style = window.getComputedStyle(el);
			return style.display !== 'none' &&
			       style.visibility !== 'hidden' &&
			       style.opacity !== '0' &&
			       el.offsetParent !== null;
		})()
	`, sel), &visible))

	return visible, err
}

// Wait waits for a condition.
func (b *ChromeDPBackend) Wait(selector string, timeout int, state string) error {
	ctx := b.Context()

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
		defer cancel()
	}

	sel := b.resolveSelector(selector)

	switch state {
	case "hidden":
		return chromedp.Run(ctx, chromedp.WaitNotPresent(sel))
	case "detached":
		return chromedp.Run(ctx, chromedp.WaitNotPresent(sel))
	case "attached":
		return chromedp.Run(ctx, chromedp.WaitReady(sel))
	default: // visible
		return chromedp.Run(ctx, chromedp.WaitVisible(sel))
	}
}

// WaitForTimeout waits for specified milliseconds.
func (b *ChromeDPBackend) WaitForTimeout(ms int) error {
	time.Sleep(time.Duration(ms) * time.Millisecond)
	return nil
}

// Title gets the page title.
func (b *ChromeDPBackend) Title() (string, error) {
	ctx := b.Context()
	var title string
	err := chromedp.Run(ctx, chromedp.Title(&title))
	return title, err
}

// URL gets the current URL.
func (b *ChromeDPBackend) URL() (string, error) {
	ctx := b.Context()
	var url string
	err := chromedp.Run(ctx, chromedp.Location(&url))
	return url, err
}

// Back navigates back.
func (b *ChromeDPBackend) Back() error {
	ctx := b.Context()
	return chromedp.Run(ctx, chromedp.NavigateBack())
}

// Forward navigates forward.
func (b *ChromeDPBackend) Forward() error {
	ctx := b.Context()
	return chromedp.Run(ctx, chromedp.NavigateForward())
}

// Reload reloads the page.
func (b *ChromeDPBackend) Reload() error {
	ctx := b.Context()
	return chromedp.Run(ctx, chromedp.Reload())
}

// SetViewport sets the viewport size.
func (b *ChromeDPBackend) SetViewport(width, height int) error {
	ctx := b.Context()
	return chromedp.Run(ctx, chromedp.EmulateViewport(int64(width), int64(height)))
}

// Count counts matching elements.
func (b *ChromeDPBackend) Count(selector string) (int, error) {
	ctx := b.Context()
	sel := b.resolveSelector(selector)

	var count int
	err := chromedp.Run(ctx, chromedp.Evaluate(fmt.Sprintf(`
		document.querySelectorAll(%q).length
	`, sel), &count))

	return count, err
}

// NewTab creates a new tab.
func (b *ChromeDPBackend) NewTab(url string) (int, error) {
	// Create new target
	ctx := b.Context()

	var targetID target.ID
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		createTarget := target.CreateTarget("about:blank")
		tid, err := createTarget.Do(ctx)
		if err != nil {
			return err
		}
		targetID = tid
		return nil
	})); err != nil {
		return 0, err
	}

	// Create context for new tab
	newCtx, newCancel := chromedp.NewContext(b.allocCtx, chromedp.WithTargetID(targetID))

	b.targets = append(b.targets, targetID)
	b.tabContexts[targetID] = newCtx
	b.tabCancels[targetID] = newCancel
	b.activeTab = len(b.targets) - 1

	// Navigate if URL provided
	if url != "" && url != "about:blank" {
		if _, _, err := b.Navigate(url, "load"); err != nil {
			return 0, err
		}
	}

	return b.activeTab, nil
}

// SwitchTab switches to a tab by index.
func (b *ChromeDPBackend) SwitchTab(index int) error {
	if index < 0 || index >= len(b.targets) {
		return fmt.Errorf("tab index out of range: %d", index)
	}
	b.activeTab = index
	return nil
}

// CloseTab closes a tab.
func (b *ChromeDPBackend) CloseTab(index int) error {
	if index < 0 || index >= len(b.targets) {
		return fmt.Errorf("tab index out of range: %d", index)
	}

	tid := b.targets[index]
	if cancel, ok := b.tabCancels[tid]; ok {
		cancel()
		delete(b.tabContexts, tid)
		delete(b.tabCancels, tid)
	}

	// Remove from targets
	b.targets = append(b.targets[:index], b.targets[index+1:]...)

	// Adjust active tab
	if b.activeTab >= len(b.targets) {
		b.activeTab = len(b.targets) - 1
	}
	if b.activeTab < 0 {
		b.activeTab = 0
	}

	return nil
}

// ListTabs returns info about all tabs.
func (b *ChromeDPBackend) ListTabs() ([]TabInfo, error) {
	tabs := make([]TabInfo, len(b.targets))

	for i, tid := range b.targets {
		ctx := b.tabContexts[tid]
		var url, title string

		if ctx != nil {
			chromedp.Run(ctx,
				chromedp.Location(&url),
				chromedp.Title(&title),
			)
		}

		tabs[i] = TabInfo{
			Index:  i,
			URL:    url,
			Title:  title,
			Active: i == b.activeTab,
		}
	}

	return tabs, nil
}

// resolveSelector resolves refs to actual selectors.
func (b *ChromeDPBackend) resolveSelector(selector string) string {
	// Check if it's a ref
	ref := ParseRef(selector)
	if ref == "" {
		return selector
	}

	b.refLock.RLock()
	defer b.refLock.RUnlock()

	if info, ok := b.refMap[ref]; ok {
		return info.Selector
	}

	// Return original if ref not found
	return selector
}

// IsRef checks if a selector is a ref.
func IsRef(selector string) bool {
	return ParseRef(selector) != ""
}

// ParseRef extracts ref ID from selector.
func ParseRef(selector string) string {
	if strings.HasPrefix(selector, "@") {
		return selector[1:]
	}
	if strings.HasPrefix(selector, "ref=") {
		return selector[4:]
	}
	if matched, _ := regexp.MatchString(`^e\d+$`, selector); matched {
		return selector
	}
	return ""
}

// GetSnapshot gets an enhanced accessibility snapshot.
func (b *ChromeDPBackend) GetSnapshot(opts SnapshotOptions) (*EnhancedSnapshot, error) {
	ctx := b.Context()

	// Use JavaScript to get accessibility tree
	script := `
	(function getAccessibilityTree() {
		function getRole(el) {
			return el.getAttribute('role') ||
				   (el.tagName === 'A' ? 'link' :
				   (el.tagName === 'BUTTON' ? 'button' :
				   (el.tagName === 'INPUT' && el.type === 'text' ? 'textbox' :
				   (el.tagName === 'INPUT' && el.type === 'checkbox' ? 'checkbox' :
				   (el.tagName === 'INPUT' && el.type === 'radio' ? 'radio' :
				   (el.tagName === 'SELECT' ? 'combobox' :
				   (el.tagName === 'TEXTAREA' ? 'textbox' :
				   (el.tagName.match(/^H[1-6]$/) ? 'heading' :
				   el.tagName.toLowerCase()))))))));
		}

		function getName(el) {
			return el.getAttribute('aria-label') ||
				   el.getAttribute('title') ||
				   (el.tagName === 'IMG' ? el.alt : '') ||
				   el.innerText?.slice(0, 50) || '';
		}

		function buildTree(el, depth) {
			if (!el || depth > 10) return null;
			if (el.nodeType !== 1) return null;
			if (window.getComputedStyle(el).display === 'none') return null;

			const role = getRole(el);
			const name = getName(el).trim();
			const children = [];

			for (const child of el.children) {
				const childNode = buildTree(child, depth + 1);
				if (childNode) children.push(childNode);
			}

			return { role, name, children };
		}

		return buildTree(document.body, 0);
	})()
	`

	var treeData *AXNode
	err := chromedp.Run(ctx, chromedp.Evaluate(script, &treeData))

	if err != nil {
		return nil, fmt.Errorf("failed to get accessibility tree: %w", err)
	}

	// Build snapshot from tree data
	snapshot := BuildSnapshotFromNodes(treeData, opts)

	// Update ref map
	b.refLock.Lock()
	b.refMap = snapshot.Refs
	b.refLock.Unlock()

	return snapshot, nil
}

// GetRefMap returns the current ref map.
func (b *ChromeDPBackend) GetRefMap() RefMap {
	b.refLock.RLock()
	defer b.refLock.RUnlock()

	// Return a copy
	result := make(RefMap, len(b.refMap))
	for k, v := range b.refMap {
		result[k] = v
	}
	return result
}

// Check checks a checkbox.
func (b *ChromeDPBackend) Check(selector string) error {
	ctx := b.Context()
	sel := b.resolveSelector(selector)

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		// Check if already checked
		var checked bool
		if err := chromedp.Evaluate(fmt.Sprintf(`document.querySelector(%q).checked`, sel), &checked).Do(ctx); err != nil {
			return err
		}
		if checked {
			return nil
		}
		return chromedp.Click(sel).Do(ctx)
	}))
}

// Uncheck unchecks a checkbox.
func (b *ChromeDPBackend) Uncheck(selector string) error {
	ctx := b.Context()
	sel := b.resolveSelector(selector)

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		// Check if already unchecked
		var checked bool
		if err := chromedp.Evaluate(fmt.Sprintf(`document.querySelector(%q).checked`, sel), &checked).Do(ctx); err != nil {
			return err
		}
		if !checked {
			return nil
		}
		return chromedp.Click(sel).Do(ctx)
	}))
}

// Select selects dropdown option(s).
func (b *ChromeDPBackend) Select(selector string, values []string) error {
	ctx := b.Context()
	sel := b.resolveSelector(selector)

	return chromedp.Run(ctx, chromedp.SetValue(sel, values[0]))
}

// Focus focuses an element.
func (b *ChromeDPBackend) Focus(selector string) error {
	ctx := b.Context()
	sel := b.resolveSelector(selector)
	return chromedp.Run(ctx, chromedp.Focus(sel))
}

// Clear clears an input.
func (b *ChromeDPBackend) Clear(selector string) error {
	ctx := b.Context()
	sel := b.resolveSelector(selector)
	return chromedp.Run(ctx, chromedp.Clear(sel))
}

// ScrollIntoView scrolls element into view.
func (b *ChromeDPBackend) ScrollIntoView(selector string) error {
	ctx := b.Context()
	sel := b.resolveSelector(selector)
	return chromedp.Run(ctx, chromedp.ScrollIntoView(sel))
}

// Scroll scrolls the page.
func (b *ChromeDPBackend) Scroll(direction string, amount int) error {
	ctx := b.Context()

	dx, dy := 0, 0
	switch direction {
	case "up":
		dy = -amount
	case "down":
		dy = amount
	case "left":
		dx = -amount
	case "right":
		dx = amount
	}

	return chromedp.Run(ctx, chromedp.Evaluate(fmt.Sprintf(`window.scrollBy(%d, %d)`, dx, dy), nil))
}

// DoubleClick double-clicks an element.
func (b *ChromeDPBackend) DoubleClick(selector string) error {
	ctx := b.Context()
	sel := b.resolveSelector(selector)
	return chromedp.Run(ctx, chromedp.DoubleClick(sel))
}

// Content gets page HTML content.
func (b *ChromeDPBackend) Content() (string, error) {
	ctx := b.Context()
	var html string
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		node, err := dom.GetDocument().Do(ctx)
		if err != nil {
			return err
		}
		html, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
		return err
	}))
	return html, err
}

// SetContent sets page HTML content.
func (b *ChromeDPBackend) SetContent(html string) error {
	ctx := b.Context()
	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		frameTree, err := page.GetFrameTree().Do(ctx)
		if err != nil {
			return err
		}
		return page.SetDocumentContent(frameTree.Frame.ID, html).Do(ctx)
	}))
}

// GetInputValue gets input element value.
func (b *ChromeDPBackend) GetInputValue(selector string) (string, error) {
	ctx := b.Context()
	sel := b.resolveSelector(selector)
	var value string
	err := chromedp.Run(ctx, chromedp.Value(sel, &value))
	return value, err
}

// SetValue sets input value directly.
func (b *ChromeDPBackend) SetValue(selector, value string) error {
	ctx := b.Context()
	sel := b.resolveSelector(selector)
	return chromedp.Run(ctx, chromedp.SetValue(sel, value))
}

// IsEnabled checks if element is enabled.
func (b *ChromeDPBackend) IsEnabled(selector string) (bool, error) {
	ctx := b.Context()
	sel := b.resolveSelector(selector)

	var disabled bool
	err := chromedp.Run(ctx, chromedp.Evaluate(fmt.Sprintf(`
		document.querySelector(%q).disabled === true
	`, sel), &disabled))

	return !disabled, err
}

// IsChecked checks if checkbox is checked.
func (b *ChromeDPBackend) IsChecked(selector string) (bool, error) {
	ctx := b.Context()
	sel := b.resolveSelector(selector)

	var checked bool
	err := chromedp.Run(ctx, chromedp.Evaluate(fmt.Sprintf(`
		document.querySelector(%q).checked === true
	`, sel), &checked))

	return checked, err
}

// GetBoundingBox gets element bounding box.
func (b *ChromeDPBackend) GetBoundingBox(selector string) (*BoundingBox, error) {
	ctx := b.Context()
	sel := b.resolveSelector(selector)

	var box *dom.BoxModel
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		nodes, err := dom.GetDocument().Do(ctx)
		if err != nil {
			return err
		}

		nodeID, err := dom.QuerySelector(nodes.NodeID, sel).Do(ctx)
		if err != nil {
			return err
		}

		box, err = dom.GetBoxModel().WithNodeID(nodeID).Do(ctx)
		return err
	}))

	if err != nil {
		return nil, err
	}

	if box == nil || box.Content == nil || len(box.Content) < 4 {
		return nil, fmt.Errorf("could not get bounding box")
	}

	return &BoundingBox{
		X:      box.Content[0],
		Y:      box.Content[1],
		Width:  box.Content[2] - box.Content[0],
		Height: box.Content[5] - box.Content[1],
	}, nil
}

// GetCookies gets cookies.
func (b *ChromeDPBackend) GetCookies() ([]Cookie, error) {
	ctx := b.Context()

	var netCookies []*network.Cookie
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		netCookies, err = storage.GetCookies().Do(ctx)
		return err
	}))

	if err != nil {
		return nil, err
	}

	cookies := make([]Cookie, len(netCookies))
	for i, c := range netCookies {
		cookies[i] = Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Expires:  int64(c.Expires),
			HTTPOnly: c.HTTPOnly,
			Secure:   c.Secure,
			SameSite: string(c.SameSite),
		}
	}

	return cookies, nil
}

// Shortcuts for semantic locators

// GetByRole finds element by ARIA role.
func (b *ChromeDPBackend) GetByRole(role, name string) string {
	if name != "" {
		return fmt.Sprintf(`[role="%s"][aria-label="%s"], [role="%s"]:has-text("%s")`, role, name, role, name)
	}
	return fmt.Sprintf(`[role="%s"]`, role)
}

// GetByText finds element by text.
func (b *ChromeDPBackend) GetByText(text string, exact bool) string {
	if exact {
		return fmt.Sprintf(`text="%s"`, text)
	}
	return fmt.Sprintf(`text=%s`, text)
}

// GetByLabel finds element by label.
func (b *ChromeDPBackend) GetByLabel(label string) string {
	return fmt.Sprintf(`[aria-label="%s"], label:has-text("%s") + input, label:has-text("%s") input`, label, label, label)
}

// GetByPlaceholder finds element by placeholder.
func (b *ChromeDPBackend) GetByPlaceholder(placeholder string) string {
	return fmt.Sprintf(`[placeholder="%s"]`, placeholder)
}

// GetByTestId finds element by data-testid.
func (b *ChromeDPBackend) GetByTestId(testId string) string {
	return fmt.Sprintf(`[data-testid="%s"]`, testId)
}

// Private helper: convert string to int with default
func atoi(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
