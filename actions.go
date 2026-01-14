package agentbrowser

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

// ExecuteCommand executes a command and returns the response.
func ExecuteCommand(cmd Command, browser *BrowserManager) Response {
	id := cmd.GetID()

	switch c := cmd.(type) {
	case *LaunchCommand:
		return handleLaunch(c, browser)
	case *NavigateCommand:
		return handleNavigate(c, browser)
	case *ClickCommand:
		return handleClick(c, browser)
	case *TypeCommand:
		return handleType(c, browser)
	case *FillCommand:
		return handleFill(c, browser)
	case *CheckCommand:
		return handleCheck(c, browser)
	case *UncheckCommand:
		return handleUncheck(c, browser)
	case *PressCommand:
		return handlePress(c, browser)
	case *HoverCommand:
		return handleHover(c, browser)
	case *FocusCommand:
		return handleFocus(c, browser)
	case *ClearCommand:
		return handleClear(c, browser)
	case *SelectCommand:
		return handleSelect(c, browser)
	case *DoubleClickCommand:
		return handleDoubleClick(c, browser)
	case *ScreenshotCommand:
		return handleScreenshot(c, browser)
	case *SnapshotCommand:
		return handleSnapshot(c, browser)
	case *EvaluateCommand:
		return handleEvaluate(c, browser)
	case *WaitCommand:
		return handleWait(c, browser)
	case *ScrollCommand:
		return handleScroll(c, browser)
	case *ScrollIntoViewCommand:
		return handleScrollIntoView(c, browser)
	case *ContentCommand:
		return handleContent(c, browser)
	case *SetContentCommand:
		return handleSetContent(c, browser)
	case *GetTextCommand:
		return handleGetText(c, browser)
	case *GetAttributeCommand:
		return handleGetAttribute(c, browser)
	case *InnerHTMLCommand:
		return handleInnerHTML(c, browser)
	case *InnerTextCommand:
		return handleInnerText(c, browser)
	case *InputValueCommand:
		return handleInputValue(c, browser)
	case *SetValueCommand:
		return handleSetValue(c, browser)
	case *IsVisibleCommand:
		return handleIsVisible(c, browser)
	case *IsEnabledCommand:
		return handleIsEnabled(c, browser)
	case *IsCheckedCommand:
		return handleIsChecked(c, browser)
	case *CountCommand:
		return handleCount(c, browser)
	case *BoundingBoxCommand:
		return handleBoundingBox(c, browser)
	case *URLCommand:
		return handleURL(c, browser)
	case *TitleCommand:
		return handleTitle(c, browser)
	case *BackCommand:
		return handleBack(c, browser)
	case *ForwardCommand:
		return handleForward(c, browser)
	case *ReloadCommand:
		return handleReload(c, browser)
	case *ViewportCommand:
		return handleViewport(c, browser)
	case *TabNewCommand:
		return handleTabNew(c, browser)
	case *TabListCommand:
		return handleTabList(c, browser)
	case *TabSwitchCommand:
		return handleTabSwitch(c, browser)
	case *TabCloseCommand:
		return handleTabClose(c, browser)
	case *CloseCommand:
		return handleClose(c, browser)
	default:
		return ErrorResponse(id, fmt.Sprintf("unsupported action: %s", cmd.GetAction()))
	}
}

func handleLaunch(cmd *LaunchCommand, browser *BrowserManager) Response {
	headless := true
	if cmd.Headless != nil {
		headless = *cmd.Headless
	}

	opts := LaunchOptions{
		Headless:       headless,
		Viewport:       cmd.Viewport,
		ExecutablePath: cmd.ExecutablePath,
		CDPPort:        cmd.CDPPort,
	}

	if err := browser.Launch(opts); err != nil {
		return ErrorResponse(cmd.ID, err.Error())
	}

	return SuccessResponse(cmd.ID, map[string]bool{"launched": true})
}

func handleNavigate(cmd *NavigateCommand, browser *BrowserManager) Response {
	waitUntil := "load"
	if cmd.WaitUntil != "" {
		waitUntil = cmd.WaitUntil
	}

	url, title, err := browser.Navigate(cmd.URL, waitUntil)
	if err != nil {
		return ErrorResponse(cmd.ID, err.Error())
	}

	return SuccessResponse(cmd.ID, NavigateData{URL: url, Title: title})
}

func handleClick(cmd *ClickCommand, browser *BrowserManager) Response {
	if err := browser.Click(cmd.Selector); err != nil {
		return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
	}
	return SuccessResponse(cmd.ID, nil)
}

func handleType(cmd *TypeCommand, browser *BrowserManager) Response {
	if err := browser.Type(cmd.Selector, cmd.Text, cmd.Delay); err != nil {
		return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
	}
	return SuccessResponse(cmd.ID, nil)
}

func handleFill(cmd *FillCommand, browser *BrowserManager) Response {
	if err := browser.Fill(cmd.Selector, cmd.Value); err != nil {
		return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
	}
	return SuccessResponse(cmd.ID, nil)
}

func handleCheck(cmd *CheckCommand, browser *BrowserManager) Response {
	if err := browser.Check(cmd.Selector); err != nil {
		return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
	}
	return SuccessResponse(cmd.ID, nil)
}

func handleUncheck(cmd *UncheckCommand, browser *BrowserManager) Response {
	if err := browser.Uncheck(cmd.Selector); err != nil {
		return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
	}
	return SuccessResponse(cmd.ID, nil)
}

func handlePress(cmd *PressCommand, browser *BrowserManager) Response {
	if err := browser.Press(cmd.Key, cmd.Selector); err != nil {
		return ErrorResponse(cmd.ID, err.Error())
	}
	return SuccessResponse(cmd.ID, nil)
}

func handleHover(cmd *HoverCommand, browser *BrowserManager) Response {
	if err := browser.Hover(cmd.Selector); err != nil {
		return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
	}
	return SuccessResponse(cmd.ID, nil)
}

func handleFocus(cmd *FocusCommand, browser *BrowserManager) Response {
	if err := browser.Focus(cmd.Selector); err != nil {
		return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
	}
	return SuccessResponse(cmd.ID, nil)
}

func handleClear(cmd *ClearCommand, browser *BrowserManager) Response {
	if err := browser.Clear(cmd.Selector); err != nil {
		return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
	}
	return SuccessResponse(cmd.ID, nil)
}

func handleSelect(cmd *SelectCommand, browser *BrowserManager) Response {
	if err := browser.Select(cmd.Selector, cmd.Values); err != nil {
		return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
	}
	return SuccessResponse(cmd.ID, nil)
}

func handleDoubleClick(cmd *DoubleClickCommand, browser *BrowserManager) Response {
	if err := browser.DoubleClick(cmd.Selector); err != nil {
		return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
	}
	return SuccessResponse(cmd.ID, nil)
}

func handleScreenshot(cmd *ScreenshotCommand, browser *BrowserManager) Response {
	quality := 80
	if cmd.Quality > 0 {
		quality = cmd.Quality
	}

	buf, err := browser.Screenshot(cmd.FullPage, cmd.Selector, quality)
	if err != nil {
		return ErrorResponse(cmd.ID, err.Error())
	}

	if cmd.Path != "" {
		if err := os.WriteFile(cmd.Path, buf, 0644); err != nil {
			return ErrorResponse(cmd.ID, fmt.Sprintf("failed to save screenshot: %v", err))
		}
		return SuccessResponse(cmd.ID, ScreenshotData{Path: cmd.Path})
	}

	return SuccessResponse(cmd.ID, ScreenshotData{Base64: base64.StdEncoding.EncodeToString(buf)})
}

func handleSnapshot(cmd *SnapshotCommand, browser *BrowserManager) Response {
	opts := SnapshotOptions{
		Interactive: cmd.Interactive,
		MaxDepth:    cmd.MaxDepth,
		Compact:     cmd.Compact,
		Selector:    cmd.Selector,
	}

	snapshot, err := browser.GetSnapshot(opts)
	if err != nil {
		return ErrorResponse(cmd.ID, err.Error())
	}

	// Convert refs to the expected format
	refsData := make(map[string]RefInfo)
	for k, v := range snapshot.Refs {
		refsData[k] = RefInfo{Role: v.Role, Name: v.Name}
	}

	return SuccessResponse(cmd.ID, SnapshotData{Snapshot: snapshot.Tree, Refs: refsData})
}

func handleEvaluate(cmd *EvaluateCommand, browser *BrowserManager) Response {
	result, err := browser.Evaluate(cmd.Script)
	if err != nil {
		return ErrorResponse(cmd.ID, err.Error())
	}
	return SuccessResponse(cmd.ID, EvaluateData{Result: result})
}

func handleWait(cmd *WaitCommand, browser *BrowserManager) Response {
	if cmd.Selector != "" {
		if err := browser.Wait(cmd.Selector, cmd.Timeout, cmd.State); err != nil {
			return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
		}
	} else if cmd.Timeout > 0 {
		if err := browser.WaitForTimeout(cmd.Timeout); err != nil {
			return ErrorResponse(cmd.ID, err.Error())
		}
	}
	return SuccessResponse(cmd.ID, nil)
}

func handleScroll(cmd *ScrollCommand, browser *BrowserManager) Response {
	amount := 100
	if cmd.Amount > 0 {
		amount = cmd.Amount
	}

	if err := browser.Scroll(cmd.Direction, amount); err != nil {
		return ErrorResponse(cmd.ID, err.Error())
	}
	return SuccessResponse(cmd.ID, nil)
}

func handleScrollIntoView(cmd *ScrollIntoViewCommand, browser *BrowserManager) Response {
	if err := browser.ScrollIntoView(cmd.Selector); err != nil {
		return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
	}
	return SuccessResponse(cmd.ID, nil)
}

func handleContent(cmd *ContentCommand, browser *BrowserManager) Response {
	if cmd.Selector != "" {
		html, err := browser.GetHTML(cmd.Selector, true)
		if err != nil {
			return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
		}
		return SuccessResponse(cmd.ID, ContentData{HTML: html})
	}

	html, err := browser.Content()
	if err != nil {
		return ErrorResponse(cmd.ID, err.Error())
	}
	return SuccessResponse(cmd.ID, ContentData{HTML: html})
}

func handleSetContent(cmd *SetContentCommand, browser *BrowserManager) Response {
	if err := browser.SetContent(cmd.HTML); err != nil {
		return ErrorResponse(cmd.ID, err.Error())
	}
	return SuccessResponse(cmd.ID, nil)
}

func handleGetText(cmd *GetTextCommand, browser *BrowserManager) Response {
	text, err := browser.GetText(cmd.Selector)
	if err != nil {
		return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
	}
	return SuccessResponse(cmd.ID, map[string]string{"text": text})
}

func handleGetAttribute(cmd *GetAttributeCommand, browser *BrowserManager) Response {
	value, err := browser.GetAttribute(cmd.Selector, cmd.Attribute)
	if err != nil {
		return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
	}
	return SuccessResponse(cmd.ID, map[string]string{"value": value})
}

func handleInnerHTML(cmd *InnerHTMLCommand, browser *BrowserManager) Response {
	html, err := browser.GetHTML(cmd.Selector, false)
	if err != nil {
		return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
	}
	return SuccessResponse(cmd.ID, map[string]string{"html": html})
}

func handleInnerText(cmd *InnerTextCommand, browser *BrowserManager) Response {
	text, err := browser.GetText(cmd.Selector)
	if err != nil {
		return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
	}
	return SuccessResponse(cmd.ID, map[string]string{"text": text})
}

func handleInputValue(cmd *InputValueCommand, browser *BrowserManager) Response {
	value, err := browser.GetInputValue(cmd.Selector)
	if err != nil {
		return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
	}
	return SuccessResponse(cmd.ID, map[string]string{"value": value})
}

func handleSetValue(cmd *SetValueCommand, browser *BrowserManager) Response {
	if err := browser.SetValue(cmd.Selector, cmd.Value); err != nil {
		return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
	}
	return SuccessResponse(cmd.ID, nil)
}

func handleIsVisible(cmd *IsVisibleCommand, browser *BrowserManager) Response {
	visible, err := browser.IsVisible(cmd.Selector)
	if err != nil {
		return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
	}
	return SuccessResponse(cmd.ID, map[string]bool{"visible": visible})
}

func handleIsEnabled(cmd *IsEnabledCommand, browser *BrowserManager) Response {
	enabled, err := browser.IsEnabled(cmd.Selector)
	if err != nil {
		return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
	}
	return SuccessResponse(cmd.ID, map[string]bool{"enabled": enabled})
}

func handleIsChecked(cmd *IsCheckedCommand, browser *BrowserManager) Response {
	checked, err := browser.IsChecked(cmd.Selector)
	if err != nil {
		return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
	}
	return SuccessResponse(cmd.ID, map[string]bool{"checked": checked})
}

func handleCount(cmd *CountCommand, browser *BrowserManager) Response {
	count, err := browser.Count(cmd.Selector)
	if err != nil {
		return ErrorResponse(cmd.ID, err.Error())
	}
	return SuccessResponse(cmd.ID, map[string]int{"count": count})
}

func handleBoundingBox(cmd *BoundingBoxCommand, browser *BrowserManager) Response {
	box, err := browser.GetBoundingBox(cmd.Selector)
	if err != nil {
		return ErrorResponse(cmd.ID, toAIFriendlyError(err, cmd.Selector))
	}
	return SuccessResponse(cmd.ID, box)
}

func handleURL(cmd *URLCommand, browser *BrowserManager) Response {
	url, err := browser.URL()
	if err != nil {
		return ErrorResponse(cmd.ID, err.Error())
	}
	return SuccessResponse(cmd.ID, map[string]string{"url": url})
}

func handleTitle(cmd *TitleCommand, browser *BrowserManager) Response {
	title, err := browser.Title()
	if err != nil {
		return ErrorResponse(cmd.ID, err.Error())
	}
	return SuccessResponse(cmd.ID, map[string]string{"title": title})
}

func handleBack(cmd *BackCommand, browser *BrowserManager) Response {
	if err := browser.Back(); err != nil {
		return ErrorResponse(cmd.ID, err.Error())
	}
	return SuccessResponse(cmd.ID, nil)
}

func handleForward(cmd *ForwardCommand, browser *BrowserManager) Response {
	if err := browser.Forward(); err != nil {
		return ErrorResponse(cmd.ID, err.Error())
	}
	return SuccessResponse(cmd.ID, nil)
}

func handleReload(cmd *ReloadCommand, browser *BrowserManager) Response {
	if err := browser.Reload(); err != nil {
		return ErrorResponse(cmd.ID, err.Error())
	}
	return SuccessResponse(cmd.ID, nil)
}

func handleViewport(cmd *ViewportCommand, browser *BrowserManager) Response {
	if err := browser.SetViewport(cmd.Width, cmd.Height); err != nil {
		return ErrorResponse(cmd.ID, err.Error())
	}
	return SuccessResponse(cmd.ID, nil)
}

func handleTabNew(cmd *TabNewCommand, browser *BrowserManager) Response {
	index, err := browser.NewTab(cmd.URL)
	if err != nil {
		return ErrorResponse(cmd.ID, err.Error())
	}
	tabs, _ := browser.ListTabs()
	return SuccessResponse(cmd.ID, TabNewData{Index: index, Total: len(tabs)})
}

func handleTabList(cmd *TabListCommand, browser *BrowserManager) Response {
	tabs, err := browser.ListTabs()
	if err != nil {
		return ErrorResponse(cmd.ID, err.Error())
	}

	active := 0
	for i, t := range tabs {
		if t.Active {
			active = i
			break
		}
	}

	return SuccessResponse(cmd.ID, TabListData{Tabs: tabs, Active: active})
}

func handleTabSwitch(cmd *TabSwitchCommand, browser *BrowserManager) Response {
	if err := browser.SwitchTab(cmd.Index); err != nil {
		return ErrorResponse(cmd.ID, err.Error())
	}

	url, _ := browser.URL()
	title, _ := browser.Title()

	return SuccessResponse(cmd.ID, TabSwitchData{Index: cmd.Index, URL: url, Title: title})
}

func handleTabClose(cmd *TabCloseCommand, browser *BrowserManager) Response {
	// Get active tab index from ListTabs
	tabs, _ := browser.ListTabs()
	index := 0
	for i, t := range tabs {
		if t.Active {
			index = i
			break
		}
	}

	if cmd.Index != nil {
		index = *cmd.Index
	}

	if err := browser.CloseTab(index); err != nil {
		return ErrorResponse(cmd.ID, err.Error())
	}

	tabs, _ = browser.ListTabs()
	return SuccessResponse(cmd.ID, TabCloseData{Closed: index, Remaining: len(tabs)})
}

func handleClose(cmd *CloseCommand, browser *BrowserManager) Response {
	if err := browser.Close(); err != nil {
		return ErrorResponse(cmd.ID, err.Error())
	}
	return SuccessResponse(cmd.ID, map[string]bool{"closed": true})
}

// toAIFriendlyError converts chromedp errors to user-friendly messages.
func toAIFriendlyError(err error, selector string) string {
	errStr := err.Error()

	// Check for common error patterns
	if contains(errStr, "timeout") {
		return fmt.Sprintf("Timeout waiting for element: %s. Try using 'snapshot' to see available elements.", selector)
	}
	if contains(errStr, "not found") || contains(errStr, "no node") {
		return fmt.Sprintf("Element not found: %s. Use 'snapshot' to find correct ref or selector.", selector)
	}
	if contains(errStr, "not visible") {
		return fmt.Sprintf("Element not visible: %s. It may be hidden or off-screen.", selector)
	}
	if contains(errStr, "not interactable") || contains(errStr, "not clickable") {
		return fmt.Sprintf("Element not interactable: %s. It may be covered by another element.", selector)
	}

	return errStr
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// SerializeCommand serializes a command to JSON.
func SerializeCommand(cmd Command) ([]byte, error) {
	return json.Marshal(cmd)
}
