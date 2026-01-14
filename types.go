// Package agentbrowser provides headless browser automation for AI agents.
package agentbrowser

import "encoding/json"

// BaseCommand contains common fields for all commands.
type BaseCommand struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}

// Viewport represents browser viewport dimensions.
type Viewport struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// LaunchCommand starts a browser instance.
type LaunchCommand struct {
	BaseCommand
	Headless       *bool             `json:"headless,omitempty"`
	Viewport       *Viewport         `json:"viewport,omitempty"`
	Browser        string            `json:"browser,omitempty"` // chromium, firefox, webkit
	Headers        map[string]string `json:"headers,omitempty"`
	ExecutablePath string            `json:"executablePath,omitempty"`
	CDPPort        int               `json:"cdpPort,omitempty"`
	Extensions     []string          `json:"extensions,omitempty"`
}

// NavigateCommand navigates to a URL.
type NavigateCommand struct {
	BaseCommand
	URL       string            `json:"url"`
	WaitUntil string            `json:"waitUntil,omitempty"` // load, domcontentloaded, networkidle
	Headers   map[string]string `json:"headers,omitempty"`
}

// ClickCommand clicks an element.
type ClickCommand struct {
	BaseCommand
	Selector   string `json:"selector"`
	Button     string `json:"button,omitempty"` // left, right, middle
	ClickCount int    `json:"clickCount,omitempty"`
	Delay      int    `json:"delay,omitempty"`
}

// TypeCommand types text into an element.
type TypeCommand struct {
	BaseCommand
	Selector string `json:"selector"`
	Text     string `json:"text"`
	Delay    int    `json:"delay,omitempty"`
	Clear    bool   `json:"clear,omitempty"`
}

// FillCommand clears and fills an input.
type FillCommand struct {
	BaseCommand
	Selector string `json:"selector"`
	Value    string `json:"value"`
}

// CheckCommand checks a checkbox.
type CheckCommand struct {
	BaseCommand
	Selector string `json:"selector"`
}

// UncheckCommand unchecks a checkbox.
type UncheckCommand struct {
	BaseCommand
	Selector string `json:"selector"`
}

// UploadCommand uploads files.
type UploadCommand struct {
	BaseCommand
	Selector string   `json:"selector"`
	Files    []string `json:"files"`
}

// DoubleClickCommand double-clicks an element.
type DoubleClickCommand struct {
	BaseCommand
	Selector string `json:"selector"`
}

// FocusCommand focuses an element.
type FocusCommand struct {
	BaseCommand
	Selector string `json:"selector"`
}

// DragCommand drags from source to target.
type DragCommand struct {
	BaseCommand
	Source string `json:"source"`
	Target string `json:"target"`
}

// FrameCommand switches to an iframe.
type FrameCommand struct {
	BaseCommand
	Selector string `json:"selector,omitempty"`
	Name     string `json:"name,omitempty"`
	URL      string `json:"url,omitempty"`
}

// MainFrameCommand switches back to main frame.
type MainFrameCommand struct {
	BaseCommand
}

// GetByRoleCommand finds element by ARIA role.
type GetByRoleCommand struct {
	BaseCommand
	Role      string `json:"role"`
	Name      string `json:"name,omitempty"`
	SubAction string `json:"subaction"` // click, fill, check, hover
	Value     string `json:"value,omitempty"`
}

// GetByTextCommand finds element by text content.
type GetByTextCommand struct {
	BaseCommand
	Text      string `json:"text"`
	Exact     bool   `json:"exact,omitempty"`
	SubAction string `json:"subaction"` // click, hover
}

// GetByLabelCommand finds element by label.
type GetByLabelCommand struct {
	BaseCommand
	Label     string `json:"label"`
	SubAction string `json:"subaction"` // click, fill, check
	Value     string `json:"value,omitempty"`
}

// GetByPlaceholderCommand finds element by placeholder.
type GetByPlaceholderCommand struct {
	BaseCommand
	Placeholder string `json:"placeholder"`
	SubAction   string `json:"subaction"` // click, fill
	Value       string `json:"value,omitempty"`
}

// GetByAltTextCommand finds element by alt text.
type GetByAltTextCommand struct {
	BaseCommand
	Text      string `json:"text"`
	Exact     bool   `json:"exact,omitempty"`
	SubAction string `json:"subaction"` // click, hover
}

// GetByTitleCommand finds element by title attribute.
type GetByTitleCommand struct {
	BaseCommand
	Text      string `json:"text"`
	Exact     bool   `json:"exact,omitempty"`
	SubAction string `json:"subaction"` // click, hover
}

// GetByTestIdCommand finds element by data-testid.
type GetByTestIdCommand struct {
	BaseCommand
	TestID    string `json:"testId"`
	SubAction string `json:"subaction"` // click, fill, check, hover
	Value     string `json:"value,omitempty"`
}

// NthCommand selects nth matching element.
type NthCommand struct {
	BaseCommand
	Selector  string `json:"selector"`
	Index     int    `json:"index"` // 0-based, -1 for last
	SubAction string `json:"subaction"`
	Value     string `json:"value,omitempty"`
}

// Cookie represents a browser cookie.
type Cookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	URL      string `json:"url,omitempty"`
	Domain   string `json:"domain,omitempty"`
	Path     string `json:"path,omitempty"`
	Expires  int64  `json:"expires,omitempty"`
	HTTPOnly bool   `json:"httpOnly,omitempty"`
	Secure   bool   `json:"secure,omitempty"`
	SameSite string `json:"sameSite,omitempty"` // Strict, Lax, None
}

// CookiesGetCommand gets cookies.
type CookiesGetCommand struct {
	BaseCommand
	URLs []string `json:"urls,omitempty"`
}

// CookiesSetCommand sets cookies.
type CookiesSetCommand struct {
	BaseCommand
	Cookies []Cookie `json:"cookies"`
}

// CookiesClearCommand clears all cookies.
type CookiesClearCommand struct {
	BaseCommand
}

// StorageGetCommand gets storage value.
type StorageGetCommand struct {
	BaseCommand
	Key  string `json:"key,omitempty"`
	Type string `json:"type"` // local, session
}

// StorageSetCommand sets storage value.
type StorageSetCommand struct {
	BaseCommand
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"` // local, session
}

// StorageClearCommand clears storage.
type StorageClearCommand struct {
	BaseCommand
	Type string `json:"type"` // local, session
}

// DialogCommand handles dialogs.
type DialogCommand struct {
	BaseCommand
	Response   string `json:"response"` // accept, dismiss
	PromptText string `json:"promptText,omitempty"`
}

// PdfCommand saves page as PDF.
type PdfCommand struct {
	BaseCommand
	Path   string `json:"path"`
	Format string `json:"format,omitempty"` // Letter, Legal, A4, etc.
}

// RouteCommand intercepts network requests.
type RouteCommand struct {
	BaseCommand
	URL      string         `json:"url"`
	Response *RouteResponse `json:"response,omitempty"`
	Abort    bool           `json:"abort,omitempty"`
}

// RouteResponse defines mock response.
type RouteResponse struct {
	Status      int               `json:"status,omitempty"`
	Body        string            `json:"body,omitempty"`
	ContentType string            `json:"contentType,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// UnrouteCommand removes route.
type UnrouteCommand struct {
	BaseCommand
	URL string `json:"url,omitempty"`
}

// RequestsCommand gets tracked requests.
type RequestsCommand struct {
	BaseCommand
	Filter string `json:"filter,omitempty"`
	Clear  bool   `json:"clear,omitempty"`
}

// DownloadCommand triggers download.
type DownloadCommand struct {
	BaseCommand
	Selector string `json:"selector"`
	Path     string `json:"path"`
}

// GeolocationCommand sets geolocation.
type GeolocationCommand struct {
	BaseCommand
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Accuracy  float64 `json:"accuracy,omitempty"`
}

// PermissionsCommand grants/denies permissions.
type PermissionsCommand struct {
	BaseCommand
	Permissions []string `json:"permissions"`
	Grant       bool     `json:"grant"`
}

// ViewportCommand sets viewport size.
type ViewportCommand struct {
	BaseCommand
	Width  int `json:"width"`
	Height int `json:"height"`
}

// UserAgentCommand sets user agent.
type UserAgentCommand struct {
	BaseCommand
	UserAgent string `json:"userAgent"`
}

// DeviceCommand emulates a device.
type DeviceCommand struct {
	BaseCommand
	Device string `json:"device"`
}

// BackCommand navigates back.
type BackCommand struct {
	BaseCommand
}

// ForwardCommand navigates forward.
type ForwardCommand struct {
	BaseCommand
}

// ReloadCommand reloads the page.
type ReloadCommand struct {
	BaseCommand
}

// URLCommand gets current URL.
type URLCommand struct {
	BaseCommand
}

// TitleCommand gets page title.
type TitleCommand struct {
	BaseCommand
}

// GetAttributeCommand gets element attribute.
type GetAttributeCommand struct {
	BaseCommand
	Selector  string `json:"selector"`
	Attribute string `json:"attribute"`
}

// GetTextCommand gets element text.
type GetTextCommand struct {
	BaseCommand
	Selector string `json:"selector"`
}

// IsVisibleCommand checks visibility.
type IsVisibleCommand struct {
	BaseCommand
	Selector string `json:"selector"`
}

// IsEnabledCommand checks if enabled.
type IsEnabledCommand struct {
	BaseCommand
	Selector string `json:"selector"`
}

// IsCheckedCommand checks if checked.
type IsCheckedCommand struct {
	BaseCommand
	Selector string `json:"selector"`
}

// CountCommand counts matching elements.
type CountCommand struct {
	BaseCommand
	Selector string `json:"selector"`
}

// BoundingBoxCommand gets element bounds.
type BoundingBoxCommand struct {
	BaseCommand
	Selector string `json:"selector"`
}

// PressCommand presses a key.
type PressCommand struct {
	BaseCommand
	Key      string `json:"key"`
	Selector string `json:"selector,omitempty"`
}

// ScreenshotCommand takes a screenshot.
type ScreenshotCommand struct {
	BaseCommand
	Path     string `json:"path,omitempty"`
	FullPage bool   `json:"fullPage,omitempty"`
	Selector string `json:"selector,omitempty"`
	Format   string `json:"format,omitempty"` // png, jpeg
	Quality  int    `json:"quality,omitempty"`
}

// SnapshotCommand gets accessibility tree.
type SnapshotCommand struct {
	BaseCommand
	Interactive bool   `json:"interactive,omitempty"`
	MaxDepth    int    `json:"maxDepth,omitempty"`
	Compact     bool   `json:"compact,omitempty"`
	Selector    string `json:"selector,omitempty"`
}

// EvaluateCommand runs JavaScript.
type EvaluateCommand struct {
	BaseCommand
	Script string        `json:"script"`
	Args   []interface{} `json:"args,omitempty"`
}

// WaitCommand waits for condition.
type WaitCommand struct {
	BaseCommand
	Selector string `json:"selector,omitempty"`
	Timeout  int    `json:"timeout,omitempty"`
	State    string `json:"state,omitempty"` // attached, detached, visible, hidden
}

// WaitForURLCommand waits for URL pattern.
type WaitForURLCommand struct {
	BaseCommand
	URL     string `json:"url"`
	Timeout int    `json:"timeout,omitempty"`
}

// WaitForLoadStateCommand waits for load state.
type WaitForLoadStateCommand struct {
	BaseCommand
	State   string `json:"state"` // load, domcontentloaded, networkidle
	Timeout int    `json:"timeout,omitempty"`
}

// WaitForFunctionCommand waits for JS condition.
type WaitForFunctionCommand struct {
	BaseCommand
	Expression string `json:"expression"`
	Timeout    int    `json:"timeout,omitempty"`
}

// ScrollCommand scrolls the page.
type ScrollCommand struct {
	BaseCommand
	Selector  string `json:"selector,omitempty"`
	X         int    `json:"x,omitempty"`
	Y         int    `json:"y,omitempty"`
	Direction string `json:"direction,omitempty"` // up, down, left, right
	Amount    int    `json:"amount,omitempty"`
}

// ScrollIntoViewCommand scrolls element into view.
type ScrollIntoViewCommand struct {
	BaseCommand
	Selector string `json:"selector"`
}

// SelectCommand selects dropdown option.
type SelectCommand struct {
	BaseCommand
	Selector string   `json:"selector"`
	Values   []string `json:"values"`
}

// MultiSelectCommand selects multiple options.
type MultiSelectCommand struct {
	BaseCommand
	Selector string   `json:"selector"`
	Values   []string `json:"values"`
}

// HoverCommand hovers over element.
type HoverCommand struct {
	BaseCommand
	Selector string `json:"selector"`
}

// ContentCommand gets page HTML.
type ContentCommand struct {
	BaseCommand
	Selector string `json:"selector,omitempty"`
}

// SetContentCommand sets page HTML.
type SetContentCommand struct {
	BaseCommand
	HTML string `json:"html"`
}

// CloseCommand closes the browser.
type CloseCommand struct {
	BaseCommand
}

// TabNewCommand opens a new tab.
type TabNewCommand struct {
	BaseCommand
	URL string `json:"url,omitempty"`
}

// TabListCommand lists all tabs.
type TabListCommand struct {
	BaseCommand
}

// TabSwitchCommand switches to a tab.
type TabSwitchCommand struct {
	BaseCommand
	Index int `json:"index"`
}

// TabCloseCommand closes a tab.
type TabCloseCommand struct {
	BaseCommand
	Index *int `json:"index,omitempty"`
}

// WindowNewCommand opens a new window.
type WindowNewCommand struct {
	BaseCommand
	Viewport *Viewport `json:"viewport,omitempty"`
}

// MouseMoveCommand moves the mouse.
type MouseMoveCommand struct {
	BaseCommand
	X int `json:"x"`
	Y int `json:"y"`
}

// MouseDownCommand presses mouse button.
type MouseDownCommand struct {
	BaseCommand
	Button string `json:"button,omitempty"` // left, right, middle
}

// MouseUpCommand releases mouse button.
type MouseUpCommand struct {
	BaseCommand
	Button string `json:"button,omitempty"`
}

// WheelCommand scrolls with mouse wheel.
type WheelCommand struct {
	BaseCommand
	DeltaX   int    `json:"deltaX,omitempty"`
	DeltaY   int    `json:"deltaY,omitempty"`
	Selector string `json:"selector,omitempty"`
}

// KeyDownCommand holds a key down.
type KeyDownCommand struct {
	BaseCommand
	Key string `json:"key"`
}

// KeyUpCommand releases a key.
type KeyUpCommand struct {
	BaseCommand
	Key string `json:"key"`
}

// InsertTextCommand inserts text without key events.
type InsertTextCommand struct {
	BaseCommand
	Text string `json:"text"`
}

// KeyboardCommand presses key combo.
type KeyboardCommand struct {
	BaseCommand
	Keys string `json:"keys"` // e.g., "Control+a"
}

// TimezoneCommand sets timezone.
type TimezoneCommand struct {
	BaseCommand
	Timezone string `json:"timezone"`
}

// LocaleCommand sets locale.
type LocaleCommand struct {
	BaseCommand
	Locale string `json:"locale"`
}

// HTTPCredentialsCommand sets HTTP auth.
type HTTPCredentialsCommand struct {
	BaseCommand
	Username string `json:"username"`
	Password string `json:"password"`
}

// OfflineCommand toggles offline mode.
type OfflineCommand struct {
	BaseCommand
	Offline bool `json:"offline"`
}

// HeadersCommand sets extra HTTP headers.
type HeadersCommand struct {
	BaseCommand
	Headers map[string]string `json:"headers"`
}

// EmulateMediaCommand emulates media features.
type EmulateMediaCommand struct {
	BaseCommand
	Media         string `json:"media,omitempty"`         // screen, print
	ColorScheme   string `json:"colorScheme,omitempty"`   // light, dark
	ReducedMotion string `json:"reducedMotion,omitempty"` // reduce, no-preference
	ForcedColors  string `json:"forcedColors,omitempty"`  // active, none
}

// TapCommand taps (touch) an element.
type TapCommand struct {
	BaseCommand
	Selector string `json:"selector"`
}

// HighlightCommand highlights an element.
type HighlightCommand struct {
	BaseCommand
	Selector string `json:"selector"`
}

// ClearCommand clears an input.
type ClearCommand struct {
	BaseCommand
	Selector string `json:"selector"`
}

// SelectAllCommand selects all text.
type SelectAllCommand struct {
	BaseCommand
	Selector string `json:"selector"`
}

// InnerTextCommand gets inner text.
type InnerTextCommand struct {
	BaseCommand
	Selector string `json:"selector"`
}

// InnerHTMLCommand gets inner HTML.
type InnerHTMLCommand struct {
	BaseCommand
	Selector string `json:"selector"`
}

// InputValueCommand gets input value.
type InputValueCommand struct {
	BaseCommand
	Selector string `json:"selector"`
}

// SetValueCommand sets input value directly.
type SetValueCommand struct {
	BaseCommand
	Selector string `json:"selector"`
	Value    string `json:"value"`
}

// DispatchEventCommand dispatches a DOM event.
type DispatchEventCommand struct {
	BaseCommand
	Selector  string                 `json:"selector"`
	Event     string                 `json:"event"`
	EventInit map[string]interface{} `json:"eventInit,omitempty"`
}

// AddScriptCommand adds a script tag.
type AddScriptCommand struct {
	BaseCommand
	Content string `json:"content,omitempty"`
	URL     string `json:"url,omitempty"`
}

// AddStyleCommand adds a style tag.
type AddStyleCommand struct {
	BaseCommand
	Content string `json:"content,omitempty"`
	URL     string `json:"url,omitempty"`
}

// AddInitScriptCommand adds init script.
type AddInitScriptCommand struct {
	BaseCommand
	Script string `json:"script"`
}

// TraceStartCommand starts tracing.
type TraceStartCommand struct {
	BaseCommand
	Screenshots bool `json:"screenshots,omitempty"`
	Snapshots   bool `json:"snapshots,omitempty"`
}

// TraceStopCommand stops tracing.
type TraceStopCommand struct {
	BaseCommand
	Path string `json:"path"`
}

// ConsoleCommand gets console messages.
type ConsoleCommand struct {
	BaseCommand
	Clear bool `json:"clear,omitempty"`
}

// ErrorsCommand gets page errors.
type ErrorsCommand struct {
	BaseCommand
	Clear bool `json:"clear,omitempty"`
}

// StateSaveCommand saves auth state.
type StateSaveCommand struct {
	BaseCommand
	Path string `json:"path"`
}

// StateLoadCommand loads auth state.
type StateLoadCommand struct {
	BaseCommand
	Path string `json:"path"`
}

// BringToFrontCommand brings page to front.
type BringToFrontCommand struct {
	BaseCommand
}

// PauseCommand pauses execution (debug).
type PauseCommand struct {
	BaseCommand
}

// ScreencastStartCommand starts screencast.
type ScreencastStartCommand struct {
	BaseCommand
	Format        string `json:"format,omitempty"` // jpeg, png
	Quality       int    `json:"quality,omitempty"`
	MaxWidth      int    `json:"maxWidth,omitempty"`
	MaxHeight     int    `json:"maxHeight,omitempty"`
	EveryNthFrame int    `json:"everyNthFrame,omitempty"`
}

// ScreencastStopCommand stops screencast.
type ScreencastStopCommand struct {
	BaseCommand
}

// InputMouseCommand injects mouse event.
type InputMouseCommand struct {
	BaseCommand
	Type       string `json:"type"` // mousePressed, mouseReleased, mouseMoved, mouseWheel
	X          int    `json:"x"`
	Y          int    `json:"y"`
	Button     string `json:"button,omitempty"`
	ClickCount int    `json:"clickCount,omitempty"`
	DeltaX     int    `json:"deltaX,omitempty"`
	DeltaY     int    `json:"deltaY,omitempty"`
	Modifiers  int    `json:"modifiers,omitempty"`
}

// InputKeyboardCommand injects keyboard event.
type InputKeyboardCommand struct {
	BaseCommand
	Type      string `json:"type"` // keyDown, keyUp, char
	Key       string `json:"key,omitempty"`
	Code      string `json:"code,omitempty"`
	Text      string `json:"text,omitempty"`
	Modifiers int    `json:"modifiers,omitempty"`
}

// TouchPoint represents a touch point.
type TouchPoint struct {
	X  int `json:"x"`
	Y  int `json:"y"`
	ID int `json:"id,omitempty"`
}

// InputTouchCommand injects touch event.
type InputTouchCommand struct {
	BaseCommand
	Type        string       `json:"type"` // touchStart, touchEnd, touchMove, touchCancel
	TouchPoints []TouchPoint `json:"touchPoints"`
	Modifiers   int          `json:"modifiers,omitempty"`
}

// ClipboardCommand manages clipboard.
type ClipboardCommand struct {
	BaseCommand
	Operation string `json:"operation"` // copy, paste, read
	Text      string `json:"text,omitempty"`
}

// Command is a union type for all commands.
type Command interface {
	GetID() string
	GetAction() string
}

// GetID returns the command ID.
func (c BaseCommand) GetID() string { return c.ID }

// GetAction returns the command action.
func (c BaseCommand) GetAction() string { return c.Action }

// Response types

// Response is the base response interface.
type Response struct {
	ID      string          `json:"id"`
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// NavigateData is the response for navigate.
type NavigateData struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

// ScreenshotData is the response for screenshot.
type ScreenshotData struct {
	Path   string `json:"path,omitempty"`
	Base64 string `json:"base64,omitempty"`
}

// SnapshotData is the response for snapshot.
type SnapshotData struct {
	Snapshot string             `json:"snapshot"`
	Refs     map[string]RefInfo `json:"refs,omitempty"`
}

// RefInfo describes a ref in the snapshot.
type RefInfo struct {
	Role string `json:"role"`
	Name string `json:"name,omitempty"`
}

// EvaluateData is the response for evaluate.
type EvaluateData struct {
	Result interface{} `json:"result"`
}

// ContentData is the response for content.
type ContentData struct {
	HTML string `json:"html"`
}

// TabInfo describes a tab.
type TabInfo struct {
	Index  int    `json:"index"`
	URL    string `json:"url"`
	Title  string `json:"title"`
	Active bool   `json:"active"`
}

// TabListData is the response for tab list.
type TabListData struct {
	Tabs   []TabInfo `json:"tabs"`
	Active int       `json:"active"`
}

// TabNewData is the response for new tab.
type TabNewData struct {
	Index int `json:"index"`
	Total int `json:"total"`
}

// TabSwitchData is the response for tab switch.
type TabSwitchData struct {
	Index int    `json:"index"`
	URL   string `json:"url"`
	Title string `json:"title"`
}

// TabCloseData is the response for tab close.
type TabCloseData struct {
	Closed    int `json:"closed"`
	Remaining int `json:"remaining"`
}

// BoundingBox describes element bounds.
type BoundingBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// TrackedRequest describes a tracked network request.
type TrackedRequest struct {
	URL          string            `json:"url"`
	Method       string            `json:"method"`
	Headers      map[string]string `json:"headers"`
	Timestamp    int64             `json:"timestamp"`
	ResourceType string            `json:"resourceType"`
}

// ConsoleMessage describes a console message.
type ConsoleMessage struct {
	Type      string `json:"type"`
	Text      string `json:"text"`
	Timestamp int64  `json:"timestamp"`
}

// PageError describes a page error.
type PageError struct {
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// ScreencastFrame describes a screencast frame.
type ScreencastFrame struct {
	Data     string             `json:"data"` // base64
	Metadata ScreencastMetadata `json:"metadata"`
}

// ScreencastMetadata describes frame metadata.
type ScreencastMetadata struct {
	OffsetTop       int     `json:"offsetTop"`
	PageScaleFactor float64 `json:"pageScaleFactor"`
	DeviceWidth     int     `json:"deviceWidth"`
	DeviceHeight    int     `json:"deviceHeight"`
	ScrollOffsetX   int     `json:"scrollOffsetX"`
	ScrollOffsetY   int     `json:"scrollOffsetY"`
	Timestamp       float64 `json:"timestamp,omitempty"`
}
