package agentbrowser

import (
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"

	"github.com/playwright-community/playwright-go"
)

// PlaywrightBackend implements BrowserBackend using playwright-go.
type PlaywrightBackend struct {
	pw        *playwright.Playwright
	browser   playwright.Browser
	pages     []playwright.Page
	context   playwright.BrowserContext
	launched  atomic.Bool
	headless  bool
	viewport  *Viewport
	refMap    RefMap
	refLock   sync.RWMutex
	activeTab int
}

// NewPlaywrightBackend creates a new Playwright backend.
func NewPlaywrightBackend() *PlaywrightBackend {
	return &PlaywrightBackend{
		refMap: make(RefMap),
		pages:  make([]playwright.Page, 0),
	}
}

// Lifecycle

func (p *PlaywrightBackend) Launch(opts LaunchOptions) error {
	if p.launched.Load() {
		// Check if headless setting changed
		if p.headless != opts.Headless {
			// Need to relaunch with new settings
			p.Close()
		} else {
			return nil // Already launched with same settings
		}
	}

	var err error
	p.pw, err = playwright.Run()
	if err != nil {
		return fmt.Errorf("failed to start playwright: %w", err)
	}

	p.headless = opts.Headless
	if opts.Viewport != nil {
		p.viewport = opts.Viewport
	} else {
		p.viewport = &Viewport{Width: 1280, Height: 720}
	}

	// Launch browser with anti-detection arguments
	args := []string{
		"--disable-blink-features=AutomationControlled",
		"--disable-infobars",
	}

	// Optional: sandbox and shm flags (via environment variables)
	if os.Getenv("AGENT_BROWSER_NO_SANDBOX") == "1" {
		args = append(args, "--no-sandbox")
	}
	if os.Getenv("AGENT_BROWSER_DISABLE_SHM") == "1" {
		args = append(args, "--disable-dev-shm-usage")
	}

	// Use persistent context if UserDataDir is specified
	if opts.UserDataDir != "" {
		// Launch persistent context (like Python's launch_persistent_context)
		contextOpts := playwright.BrowserTypeLaunchPersistentContextOptions{
			Headless:          &opts.Headless,
			Args:              args,
			IgnoreDefaultArgs: []string{"--enable-automation"},
		}
		if opts.ExecutablePath != "" {
			contextOpts.ExecutablePath = &opts.ExecutablePath
		}
		if opts.Locale != "" {
			contextOpts.Locale = &opts.Locale
		}
		if p.viewport != nil {
			contextOpts.Viewport = &playwright.Size{
				Width:  p.viewport.Width,
				Height: p.viewport.Height,
			}
		}

		p.context, err = p.pw.Chromium.LaunchPersistentContext(opts.UserDataDir, contextOpts)
		if err != nil {
			_ = p.pw.Stop()
			return fmt.Errorf("failed to launch persistent context: %w", err)
		}

		// Get the first page
		pages := p.context.Pages()
		if len(pages) > 0 {
			p.pages = []playwright.Page{pages[0]}
			p.activeTab = 0
		}
	} else {
		// Regular browser launch
		launchOpts := playwright.BrowserTypeLaunchOptions{
			Headless:          &opts.Headless,
			Args:              args,
			IgnoreDefaultArgs: []string{"--enable-automation"},
		}
		if opts.ExecutablePath != "" {
			launchOpts.ExecutablePath = &opts.ExecutablePath
		}

		p.browser, err = p.pw.Chromium.Launch(launchOpts)
		if err != nil {
			_ = p.pw.Stop()
			return fmt.Errorf("failed to launch browser: %w", err)
		}

		// Create context
		contextOpts := playwright.BrowserNewContextOptions{}
		if opts.Locale != "" {
			contextOpts.Locale = &opts.Locale
		}
		if p.viewport != nil {
			contextOpts.Viewport = &playwright.Size{
				Width:  p.viewport.Width,
				Height: p.viewport.Height,
			}
		}

		p.context, err = p.browser.NewContext(contextOpts)
		if err != nil {
			_ = p.browser.Close()
			_ = p.pw.Stop()
			return fmt.Errorf("failed to create context: %w", err)
		}

		// Create initial page
		page, err := p.context.NewPage()
		if err != nil {
			_ = p.context.Close()
			_ = p.browser.Close()
			_ = p.pw.Stop()
			return fmt.Errorf("failed to create page: %w", err)
		}

		p.pages = append(p.pages, page)
		p.activeTab = 0
	}

	p.launched.Store(true)
	return nil
}

func (p *PlaywrightBackend) Close() error {
	if !p.launched.Load() {
		return nil
	}

	for _, page := range p.pages {
		if page != nil {
			page.Close()
		}
	}
	if p.context != nil {
		p.context.Close()
	}
	if p.browser != nil {
		p.browser.Close()
	}
	if p.pw != nil {
		_ = p.pw.Stop()
	}

	p.launched.Store(false)
	p.pages = nil
	return nil
}

func (p *PlaywrightBackend) IsLaunched() bool {
	return p.launched.Load()
}

// Navigation

func (p *PlaywrightBackend) Navigate(url string, waitUntil string) (string, string, error) {
	page := p.getCurrentPage()
	if page == nil {
		return "", "", fmt.Errorf("browser not launched")
	}

	var waitOpt playwright.WaitUntilState
	switch waitUntil {
	case "networkidle":
		waitOpt = *playwright.WaitUntilStateNetworkidle
	case "domcontentloaded":
		waitOpt = *playwright.WaitUntilStateDomcontentloaded
	default:
		waitOpt = *playwright.WaitUntilStateLoad
	}

	_, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: &waitOpt,
	})
	if err != nil {
		return "", "", err
	}

	currentURL := page.URL()
	title, _ := page.Title()

	return currentURL, title, nil
}

func (p *PlaywrightBackend) Back() error {
	page := p.getCurrentPage()
	if page == nil {
		return fmt.Errorf("browser not launched")
	}
	_, err := page.GoBack()
	return err
}

func (p *PlaywrightBackend) Forward() error {
	page := p.getCurrentPage()
	if page == nil {
		return fmt.Errorf("browser not launched")
	}
	_, err := page.GoForward()
	return err
}

func (p *PlaywrightBackend) Reload() error {
	page := p.getCurrentPage()
	if page == nil {
		return fmt.Errorf("browser not launched")
	}
	_, err := page.Reload()
	return err
}

// Interaction

func (p *PlaywrightBackend) Click(selector string) error {
	page := p.getCurrentPage()
	if page == nil {
		return fmt.Errorf("browser not launched")
	}
	sel := p.resolveSelector(selector)
	return page.Click(sel)
}

func (p *PlaywrightBackend) Fill(selector, value string) error {
	page := p.getCurrentPage()
	if page == nil {
		return fmt.Errorf("browser not launched")
	}
	sel := p.resolveSelector(selector)
	return page.Fill(sel, value)
}

func (p *PlaywrightBackend) Type(selector, text string, delay int) error {
	page := p.getCurrentPage()
	if page == nil {
		return fmt.Errorf("browser not launched")
	}
	sel := p.resolveSelector(selector)

	if delay > 0 {
		delayFloat := float64(delay)
		return page.Type(sel, text, playwright.PageTypeOptions{
			Delay: &delayFloat,
		})
	}

	return page.Type(sel, text)
}

func (p *PlaywrightBackend) Press(key string, selector string) error {
	page := p.getCurrentPage()
	if page == nil {
		return fmt.Errorf("browser not launched")
	}

	if selector != "" {
		sel := p.resolveSelector(selector)
		return page.Press(sel, key)
	}

	return page.Keyboard().Press(key)
}

func (p *PlaywrightBackend) Hover(selector string) error {
	page := p.getCurrentPage()
	if page == nil {
		return fmt.Errorf("browser not launched")
	}
	sel := p.resolveSelector(selector)
	return page.Hover(sel)
}

func (p *PlaywrightBackend) Focus(selector string) error {
	page := p.getCurrentPage()
	if page == nil {
		return fmt.Errorf("browser not launched")
	}
	sel := p.resolveSelector(selector)
	return page.Focus(sel)
}

func (p *PlaywrightBackend) Check(selector string) error {
	page := p.getCurrentPage()
	if page == nil {
		return fmt.Errorf("browser not launched")
	}
	sel := p.resolveSelector(selector)
	return page.Check(sel)
}

func (p *PlaywrightBackend) Uncheck(selector string) error {
	page := p.getCurrentPage()
	if page == nil {
		return fmt.Errorf("browser not launched")
	}
	sel := p.resolveSelector(selector)
	return page.Uncheck(sel)
}

func (p *PlaywrightBackend) Select(selector string, values []string) error {
	page := p.getCurrentPage()
	if page == nil {
		return fmt.Errorf("browser not launched")
	}
	sel := p.resolveSelector(selector)
	_, err := page.SelectOption(sel, playwright.SelectOptionValues{Values: &values})
	return err
}

func (p *PlaywrightBackend) DoubleClick(selector string) error {
	page := p.getCurrentPage()
	if page == nil {
		return fmt.Errorf("browser not launched")
	}
	sel := p.resolveSelector(selector)
	return page.Dblclick(sel)
}

func (p *PlaywrightBackend) Clear(selector string) error {
	page := p.getCurrentPage()
	if page == nil {
		return fmt.Errorf("browser not launched")
	}
	sel := p.resolveSelector(selector)
	return page.Fill(sel, "")
}

// Queries

func (p *PlaywrightBackend) GetText(selector string) (string, error) {
	page := p.getCurrentPage()
	if page == nil {
		return "", fmt.Errorf("browser not launched")
	}
	sel := p.resolveSelector(selector)
	return page.TextContent(sel)
}

func (p *PlaywrightBackend) GetAttribute(selector, attr string) (string, error) {
	page := p.getCurrentPage()
	if page == nil {
		return "", fmt.Errorf("browser not launched")
	}
	sel := p.resolveSelector(selector)
	value, err := page.GetAttribute(sel, attr)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (p *PlaywrightBackend) GetHTML(selector string, outer bool) (string, error) {
	page := p.getCurrentPage()
	if page == nil {
		return "", fmt.Errorf("browser not launched")
	}
	sel := p.resolveSelector(selector)

	if outer {
		result, err := page.Evaluate(fmt.Sprintf(`document.querySelector(%q).outerHTML`, sel))
		if err != nil {
			return "", err
		}
		if str, ok := result.(string); ok {
			return str, nil
		}
		return "", fmt.Errorf("unexpected result type")
	}

	return page.InnerHTML(sel)
}

func (p *PlaywrightBackend) GetInputValue(selector string) (string, error) {
	page := p.getCurrentPage()
	if page == nil {
		return "", fmt.Errorf("browser not launched")
	}
	sel := p.resolveSelector(selector)
	return page.InputValue(sel)
}

func (p *PlaywrightBackend) SetValue(selector, value string) error {
	page := p.getCurrentPage()
	if page == nil {
		return fmt.Errorf("browser not launched")
	}
	sel := p.resolveSelector(selector)
	return page.Fill(sel, value)
}

func (p *PlaywrightBackend) IsVisible(selector string) (bool, error) {
	page := p.getCurrentPage()
	if page == nil {
		return false, fmt.Errorf("browser not launched")
	}
	sel := p.resolveSelector(selector)
	return page.IsVisible(sel)
}

func (p *PlaywrightBackend) IsEnabled(selector string) (bool, error) {
	page := p.getCurrentPage()
	if page == nil {
		return false, fmt.Errorf("browser not launched")
	}
	sel := p.resolveSelector(selector)
	return page.IsEnabled(sel)
}

func (p *PlaywrightBackend) IsChecked(selector string) (bool, error) {
	page := p.getCurrentPage()
	if page == nil {
		return false, fmt.Errorf("browser not launched")
	}
	sel := p.resolveSelector(selector)
	return page.IsChecked(sel)
}

func (p *PlaywrightBackend) Count(selector string) (int, error) {
	page := p.getCurrentPage()
	if page == nil {
		return 0, fmt.Errorf("browser not launched")
	}
	sel := p.resolveSelector(selector)
	return page.Locator(sel).Count()
}

func (p *PlaywrightBackend) GetBoundingBox(selector string) (*BoundingBox, error) {
	page := p.getCurrentPage()
	if page == nil {
		return nil, fmt.Errorf("browser not launched")
	}
	sel := p.resolveSelector(selector)
	box, err := page.Locator(sel).BoundingBox()
	if err != nil {
		return nil, err
	}
	if box == nil {
		return nil, fmt.Errorf("element not found")
	}
	return &BoundingBox{
		X:      box.X,
		Y:      box.Y,
		Width:  box.Width,
		Height: box.Height,
	}, nil
}

// Page Info

func (p *PlaywrightBackend) URL() (string, error) {
	page := p.getCurrentPage()
	if page == nil {
		return "", fmt.Errorf("browser not launched")
	}
	return page.URL(), nil
}

func (p *PlaywrightBackend) Title() (string, error) {
	page := p.getCurrentPage()
	if page == nil {
		return "", fmt.Errorf("browser not launched")
	}
	return page.Title()
}

func (p *PlaywrightBackend) Content() (string, error) {
	page := p.getCurrentPage()
	if page == nil {
		return "", fmt.Errorf("browser not launched")
	}
	return page.Content()
}

func (p *PlaywrightBackend) SetContent(html string) error {
	page := p.getCurrentPage()
	if page == nil {
		return fmt.Errorf("browser not launched")
	}
	return page.SetContent(html)
}

// Viewport & Screenshot

func (p *PlaywrightBackend) SetViewport(width, height int) error {
	page := p.getCurrentPage()
	if page == nil {
		return fmt.Errorf("browser not launched")
	}
	return page.SetViewportSize(width, height)
}

func (p *PlaywrightBackend) Screenshot(fullPage bool, selector string, quality int) ([]byte, error) {
	page := p.getCurrentPage()
	if page == nil {
		return nil, fmt.Errorf("browser not launched")
	}

	// Use JPEG format to support quality parameter
	screenshotType := playwright.ScreenshotTypeJpeg
	opts := playwright.PageScreenshotOptions{
		FullPage: &fullPage,
		Type:     screenshotType,
	}

	if quality > 0 {
		opts.Quality = &quality
	}

	if selector != "" {
		sel := p.resolveSelector(selector)
		locator := page.Locator(sel)
		return locator.Screenshot(playwright.LocatorScreenshotOptions{
			Type:    screenshotType,
			Quality: opts.Quality,
		})
	}

	return page.Screenshot(opts)
}

// JavaScript

func (p *PlaywrightBackend) Evaluate(script string) (interface{}, error) {
	page := p.getCurrentPage()
	if page == nil {
		return nil, fmt.Errorf("browser not launched")
	}
	return page.Evaluate(script)
}

// Waiting

func (p *PlaywrightBackend) Wait(selector string, timeout int, state string) error {
	page := p.getCurrentPage()
	if page == nil {
		return fmt.Errorf("browser not launched")
	}

	sel := p.resolveSelector(selector)
	opts := playwright.PageWaitForSelectorOptions{}

	if timeout > 0 {
		timeoutFloat := float64(timeout)
		opts.Timeout = &timeoutFloat
	}

	switch state {
	case "hidden":
		opts.State = playwright.WaitForSelectorStateHidden
	case "detached":
		opts.State = playwright.WaitForSelectorStateDetached
	case "attached":
		opts.State = playwright.WaitForSelectorStateAttached
	default:
		opts.State = playwright.WaitForSelectorStateVisible
	}

	_, err := page.WaitForSelector(sel, opts)
	return err
}

func (p *PlaywrightBackend) WaitForTimeout(ms int) error {
	page := p.getCurrentPage()
	if page == nil {
		return fmt.Errorf("browser not launched")
	}
	page.WaitForTimeout(float64(ms))
	return nil
}

// Scrolling

func (p *PlaywrightBackend) Scroll(direction string, amount int) error {
	page := p.getCurrentPage()
	if page == nil {
		return fmt.Errorf("browser not launched")
	}

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

	_, err := page.Evaluate(fmt.Sprintf(`window.scrollBy(%d, %d)`, dx, dy))
	return err
}

func (p *PlaywrightBackend) ScrollIntoView(selector string) error {
	page := p.getCurrentPage()
	if page == nil {
		return fmt.Errorf("browser not launched")
	}
	sel := p.resolveSelector(selector)
	return page.Locator(sel).ScrollIntoViewIfNeeded()
}

// Tabs

func (p *PlaywrightBackend) NewTab(url string) (int, error) {
	if p.context == nil {
		return 0, fmt.Errorf("browser not launched")
	}

	page, err := p.context.NewPage()
	if err != nil {
		return 0, err
	}

	p.pages = append(p.pages, page)
	p.activeTab = len(p.pages) - 1

	if url != "" && url != "about:blank" {
		_, _, err = p.Navigate(url, "load")
		if err != nil {
			return 0, err
		}
	}

	return p.activeTab, nil
}

func (p *PlaywrightBackend) SwitchTab(index int) error {
	if index < 0 || index >= len(p.pages) {
		return fmt.Errorf("tab index out of range: %d", index)
	}
	p.activeTab = index
	return nil
}

func (p *PlaywrightBackend) CloseTab(index int) error {
	if index < 0 || index >= len(p.pages) {
		return fmt.Errorf("tab index out of range: %d", index)
	}

	if p.pages[index] != nil {
		p.pages[index].Close()
	}

	p.pages = append(p.pages[:index], p.pages[index+1:]...)

	if p.activeTab >= len(p.pages) {
		p.activeTab = len(p.pages) - 1
	}
	if p.activeTab < 0 {
		p.activeTab = 0
	}

	return nil
}

func (p *PlaywrightBackend) ListTabs() ([]TabInfo, error) {
	tabs := make([]TabInfo, len(p.pages))

	for i, page := range p.pages {
		var url, title string
		if page != nil {
			url = page.URL()
			title, _ = page.Title()
		}

		tabs[i] = TabInfo{
			Index:  i,
			URL:    url,
			Title:  title,
			Active: i == p.activeTab,
		}
	}

	return tabs, nil
}

// Snapshot

func (p *PlaywrightBackend) GetSnapshot(opts SnapshotOptions) (*EnhancedSnapshot, error) {
	page := p.getCurrentPage()
	if page == nil {
		return nil, fmt.Errorf("browser not launched")
	}

	// Wait for page to be fully loaded (networkidle ensures all resources loaded)
	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		log.Printf("Warning: WaitForLoadState failed: %v", err)
		// Continue anyway - page might already be loaded
	}

	// Use Playwright's built-in AriaSnapshot API (like TypeScript version)
	// This returns a formatted ARIA tree string
	locator := page.Locator(":root")
	ariaTree, err := locator.AriaSnapshot()
	if err != nil {
		return nil, fmt.Errorf("failed to get ARIA snapshot: %w", err)
	}

	if ariaTree == "" {
		return &EnhancedSnapshot{Tree: "(empty)", Refs: make(RefMap)}, nil
	}

	// Process the ARIA tree to add refs and apply filters
	// This matches the TypeScript processAriaTree function
	snapshot := processAriaTree(ariaTree, opts)

	p.refLock.Lock()
	p.refMap = snapshot.Refs
	p.refLock.Unlock()

	return snapshot, nil
}

// convertToAXNode converts JavaScript result to AXNode tree
func convertToAXNode(result interface{}) *AXNode {
	if result == nil {
		return nil
	}

	m, ok := result.(map[string]interface{})
	if !ok {
		return nil
	}

	node := &AXNode{}

	if role, ok := m["role"].(string); ok {
		node.Role = role
	}
	if name, ok := m["name"].(string); ok {
		node.Name = name
	}

	if children, ok := m["children"].([]interface{}); ok {
		for _, child := range children {
			if childNode := convertToAXNode(child); childNode != nil {
				node.Children = append(node.Children, childNode)
			}
		}
	}

	return node
}

func (p *PlaywrightBackend) GetRefMap() RefMap {
	p.refLock.RLock()
	defer p.refLock.RUnlock()

	result := make(RefMap, len(p.refMap))
	for k, v := range p.refMap {
		result[k] = v
	}
	return result
}

// Storage

func (p *PlaywrightBackend) GetCookies() ([]Cookie, error) {
	if p.context == nil {
		return nil, fmt.Errorf("browser not launched")
	}

	pwCookies, err := p.context.Cookies()
	if err != nil {
		return nil, err
	}

	cookies := make([]Cookie, len(pwCookies))
	for i, c := range pwCookies {
		sameSite := ""
		if c.SameSite != nil {
			sameSite = string(*c.SameSite)
		}
		cookies[i] = Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Expires:  int64(c.Expires),
			HTTPOnly: c.HttpOnly,
			Secure:   c.Secure,
			SameSite: sameSite,
		}
	}

	return cookies, nil
}

// Helper methods

func (p *PlaywrightBackend) getCurrentPage() playwright.Page {
	if len(p.pages) == 0 || p.activeTab >= len(p.pages) {
		return nil
	}
	return p.pages[p.activeTab]
}

func (p *PlaywrightBackend) resolveSelector(selector string) string {
	ref := ParseRef(selector)
	if ref == "" {
		return selector
	}

	p.refLock.RLock()
	defer p.refLock.RUnlock()

	if info, ok := p.refMap[ref]; ok {
		return info.Selector
	}

	return selector
}
