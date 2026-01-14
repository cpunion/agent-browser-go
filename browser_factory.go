package agentbrowser

// NewBrowser creates a browser backend based on the specified type.
func NewBrowser(backendType BackendType) BrowserBackend {
	switch backendType {
	case BackendPlaywright:
		return NewPlaywrightBackend()
	case BackendChromedp:
		fallthrough
	default:
		return NewChromeDPBackend()
	}
}

// BrowserManager wraps a backend for backward compatibility.
type BrowserManager struct {
	backend BrowserBackend
}

// NewBrowserManager creates a new browser manager with chromedp backend (default).
func NewBrowserManager() *BrowserManager {
	return NewBrowserManagerWithBackend(BackendChromedp)
}

// NewBrowserManagerWithBackend creates a browser manager with the specified backend.
func NewBrowserManagerWithBackend(backendType BackendType) *BrowserManager {
	return &BrowserManager{
		backend: NewBrowser(backendType),
	}
}

// Lifecycle methods - delegate to backend

func (m *BrowserManager) Launch(opts LaunchOptions) error {
	return m.backend.Launch(opts)
}

func (m *BrowserManager) Close() error {
	return m.backend.Close()
}

func (m *BrowserManager) IsLaunched() bool {
	return m.backend.IsLaunched()
}

// Navigation methods

func (m *BrowserManager) Navigate(url string, waitUntil string) (string, string, error) {
	return m.backend.Navigate(url, waitUntil)
}

func (m *BrowserManager) Back() error {
	return m.backend.Back()
}

func (m *BrowserManager) Forward() error {
	return m.backend.Forward()
}

func (m *BrowserManager) Reload() error {
	return m.backend.Reload()
}

// Interaction methods

func (m *BrowserManager) Click(selector string) error {
	return m.backend.Click(selector)
}

func (m *BrowserManager) Fill(selector, value string) error {
	return m.backend.Fill(selector, value)
}

func (m *BrowserManager) Type(selector, text string, delay int) error {
	return m.backend.Type(selector, text, delay)
}

func (m *BrowserManager) Press(key string, selector string) error {
	return m.backend.Press(key, selector)
}

func (m *BrowserManager) Hover(selector string) error {
	return m.backend.Hover(selector)
}

func (m *BrowserManager) Focus(selector string) error {
	return m.backend.Focus(selector)
}

func (m *BrowserManager) Check(selector string) error {
	return m.backend.Check(selector)
}

func (m *BrowserManager) Uncheck(selector string) error {
	return m.backend.Uncheck(selector)
}

func (m *BrowserManager) Select(selector string, values []string) error {
	return m.backend.Select(selector, values)
}

func (m *BrowserManager) DoubleClick(selector string) error {
	return m.backend.DoubleClick(selector)
}

func (m *BrowserManager) Clear(selector string) error {
	return m.backend.Clear(selector)
}

// Query methods

func (m *BrowserManager) GetText(selector string) (string, error) {
	return m.backend.GetText(selector)
}

func (m *BrowserManager) GetAttribute(selector, attr string) (string, error) {
	return m.backend.GetAttribute(selector, attr)
}

func (m *BrowserManager) GetHTML(selector string, outer bool) (string, error) {
	return m.backend.GetHTML(selector, outer)
}

func (m *BrowserManager) GetInputValue(selector string) (string, error) {
	return m.backend.GetInputValue(selector)
}

func (m *BrowserManager) SetValue(selector, value string) error {
	return m.backend.SetValue(selector, value)
}

func (m *BrowserManager) IsVisible(selector string) (bool, error) {
	return m.backend.IsVisible(selector)
}

func (m *BrowserManager) IsEnabled(selector string) (bool, error) {
	return m.backend.IsEnabled(selector)
}

func (m *BrowserManager) IsChecked(selector string) (bool, error) {
	return m.backend.IsChecked(selector)
}

func (m *BrowserManager) Count(selector string) (int, error) {
	return m.backend.Count(selector)
}

func (m *BrowserManager) GetBoundingBox(selector string) (*BoundingBox, error) {
	return m.backend.GetBoundingBox(selector)
}

// Page info methods

func (m *BrowserManager) URL() (string, error) {
	return m.backend.URL()
}

func (m *BrowserManager) Title() (string, error) {
	return m.backend.Title()
}

func (m *BrowserManager) Content() (string, error) {
	return m.backend.Content()
}

func (m *BrowserManager) SetContent(html string) error {
	return m.backend.SetContent(html)
}

// Viewport & Screenshot

func (m *BrowserManager) SetViewport(width, height int) error {
	return m.backend.SetViewport(width, height)
}

func (m *BrowserManager) Screenshot(fullPage bool, selector string, quality int) ([]byte, error) {
	return m.backend.Screenshot(fullPage, selector, quality)
}

// JavaScript

func (m *BrowserManager) Evaluate(script string) (interface{}, error) {
	return m.backend.Evaluate(script)
}

// Waiting

func (m *BrowserManager) Wait(selector string, timeout int, state string) error {
	return m.backend.Wait(selector, timeout, state)
}

func (m *BrowserManager) WaitForTimeout(ms int) error {
	return m.backend.WaitForTimeout(ms)
}

// Scrolling

func (m *BrowserManager) Scroll(direction string, amount int) error {
	return m.backend.Scroll(direction, amount)
}

func (m *BrowserManager) ScrollIntoView(selector string) error {
	return m.backend.ScrollIntoView(selector)
}

// Tabs

func (m *BrowserManager) NewTab(url string) (int, error) {
	return m.backend.NewTab(url)
}

func (m *BrowserManager) SwitchTab(index int) error {
	return m.backend.SwitchTab(index)
}

func (m *BrowserManager) CloseTab(index int) error {
	return m.backend.CloseTab(index)
}

func (m *BrowserManager) ListTabs() ([]TabInfo, error) {
	return m.backend.ListTabs()
}

// Snapshot

func (m *BrowserManager) GetSnapshot(opts SnapshotOptions) (*EnhancedSnapshot, error) {
	return m.backend.GetSnapshot(opts)
}

func (m *BrowserManager) GetRefMap() RefMap {
	return m.backend.GetRefMap()
}

// Storage

func (m *BrowserManager) GetCookies() ([]Cookie, error) {
	return m.backend.GetCookies()
}
