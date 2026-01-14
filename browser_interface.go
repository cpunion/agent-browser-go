package agentbrowser

// BrowserBackend defines the interface all browser implementations must satisfy.
type BrowserBackend interface {
	// Lifecycle
	Launch(opts LaunchOptions) error
	Close() error
	IsLaunched() bool

	// Navigation
	Navigate(url string, waitUntil string) (string, string, error)
	Back() error
	Forward() error
	Reload() error

	// Interaction
	Click(selector string) error
	Fill(selector, value string) error
	Type(selector, text string, delay int) error
	Press(key string, selector string) error
	Hover(selector string) error
	Focus(selector string) error
	Check(selector string) error
	Uncheck(selector string) error
	Select(selector string, values []string) error
	DoubleClick(selector string) error
	Clear(selector string) error

	// Queries
	GetText(selector string) (string, error)
	GetAttribute(selector, attr string) (string, error)
	GetHTML(selector string, outer bool) (string, error)
	GetInputValue(selector string) (string, error)
	SetValue(selector, value string) error
	IsVisible(selector string) (bool, error)
	IsEnabled(selector string) (bool, error)
	IsChecked(selector string) (bool, error)
	Count(selector string) (int, error)
	GetBoundingBox(selector string) (*BoundingBox, error)

	// Page Info
	URL() (string, error)
	Title() (string, error)
	Content() (string, error)
	SetContent(html string) error

	// Viewport & Screenshot
	SetViewport(width, height int) error
	Screenshot(fullPage bool, selector string, quality int) ([]byte, error)

	// JavaScript
	Evaluate(script string) (interface{}, error)

	// Waiting
	Wait(selector string, timeout int, state string) error
	WaitForTimeout(ms int) error

	// Scrolling
	Scroll(direction string, amount int) error
	ScrollIntoView(selector string) error

	// Tabs
	NewTab(url string) (int, error)
	SwitchTab(index int) error
	CloseTab(index int) error
	ListTabs() ([]TabInfo, error)

	// Snapshot
	GetSnapshot(opts SnapshotOptions) (*EnhancedSnapshot, error)
	GetRefMap() RefMap

	// Storage
	GetCookies() ([]Cookie, error)
}

// BackendType specifies which browser backend to use.
type BackendType string

const (
	BackendChromedp   BackendType = "chromedp"
	BackendPlaywright BackendType = "playwright"
)
