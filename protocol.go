package agentbrowser

import (
	"encoding/json"
	"fmt"
)

// ParseCommand parses a JSON command into the appropriate typed command.
func ParseCommand(data []byte) (Command, error) {
	var base BaseCommand
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, fmt.Errorf("failed to parse command: %w", err)
	}

	if base.ID == "" {
		return nil, fmt.Errorf("command missing id")
	}
	if base.Action == "" {
		return nil, fmt.Errorf("command missing action")
	}

	var cmd Command
	var err error

	switch base.Action {
	case "launch":
		var c LaunchCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "navigate":
		var c NavigateCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "click":
		var c ClickCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "type":
		var c TypeCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "fill":
		var c FillCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "check":
		var c CheckCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "uncheck":
		var c UncheckCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "upload":
		var c UploadCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "dblclick":
		var c DoubleClickCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "focus":
		var c FocusCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "drag":
		var c DragCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "frame":
		var c FrameCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "mainframe":
		var c MainFrameCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "getbyrole":
		var c GetByRoleCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "getbytext":
		var c GetByTextCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "getbylabel":
		var c GetByLabelCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "getbyplaceholder":
		var c GetByPlaceholderCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "getbyalttext":
		var c GetByAltTextCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "getbytitle":
		var c GetByTitleCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "getbytestid":
		var c GetByTestIdCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "nth":
		var c NthCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "cookies_get":
		var c CookiesGetCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "cookies_set":
		var c CookiesSetCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "cookies_clear":
		var c CookiesClearCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "storage_get":
		var c StorageGetCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "storage_set":
		var c StorageSetCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "storage_clear":
		var c StorageClearCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "dialog":
		var c DialogCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "pdf":
		var c PdfCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "route":
		var c RouteCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "unroute":
		var c UnrouteCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "requests":
		var c RequestsCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "download":
		var c DownloadCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "geolocation":
		var c GeolocationCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "permissions":
		var c PermissionsCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "viewport":
		var c ViewportCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "useragent":
		var c UserAgentCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "device":
		var c DeviceCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "back":
		var c BackCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "forward":
		var c ForwardCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "reload":
		var c ReloadCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "url":
		var c URLCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "title":
		var c TitleCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "getattribute":
		var c GetAttributeCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "gettext":
		var c GetTextCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "isvisible":
		var c IsVisibleCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "isenabled":
		var c IsEnabledCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "ischecked":
		var c IsCheckedCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "count":
		var c CountCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "boundingbox":
		var c BoundingBoxCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "press":
		var c PressCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "screenshot":
		var c ScreenshotCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "snapshot":
		var c SnapshotCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "evaluate":
		var c EvaluateCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "wait":
		var c WaitCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "waitforurl":
		var c WaitForURLCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "waitforloadstate":
		var c WaitForLoadStateCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "waitforfunction":
		var c WaitForFunctionCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "scroll":
		var c ScrollCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "scrollintoview":
		var c ScrollIntoViewCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "select":
		var c SelectCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "multiselect":
		var c MultiSelectCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "hover":
		var c HoverCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "content":
		var c ContentCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "setcontent":
		var c SetContentCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "close":
		var c CloseCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "tab_new":
		var c TabNewCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "tab_list":
		var c TabListCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "tab_switch":
		var c TabSwitchCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "tab_close":
		var c TabCloseCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "window_new":
		var c WindowNewCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "mousemove":
		var c MouseMoveCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "mousedown":
		var c MouseDownCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "mouseup":
		var c MouseUpCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "wheel":
		var c WheelCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "keydown":
		var c KeyDownCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "keyup":
		var c KeyUpCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "inserttext":
		var c InsertTextCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "keyboard":
		var c KeyboardCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "timezone":
		var c TimezoneCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "locale":
		var c LocaleCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "credentials":
		var c HTTPCredentialsCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "offline":
		var c OfflineCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "headers":
		var c HeadersCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "emulatemedia":
		var c EmulateMediaCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "tap":
		var c TapCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "highlight":
		var c HighlightCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "clear":
		var c ClearCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "selectall":
		var c SelectAllCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "innertext":
		var c InnerTextCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "innerhtml":
		var c InnerHTMLCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "inputvalue":
		var c InputValueCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "setvalue":
		var c SetValueCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "dispatch":
		var c DispatchEventCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "addscript":
		var c AddScriptCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "addstyle":
		var c AddStyleCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "addinitscript":
		var c AddInitScriptCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "trace_start":
		var c TraceStartCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "trace_stop":
		var c TraceStopCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "console":
		var c ConsoleCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "errors":
		var c ErrorsCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "state_save":
		var c StateSaveCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "state_load":
		var c StateLoadCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "bringtofront":
		var c BringToFrontCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "pause":
		var c PauseCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "screencast_start":
		var c ScreencastStartCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "screencast_stop":
		var c ScreencastStopCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "input_mouse":
		var c InputMouseCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "input_keyboard":
		var c InputKeyboardCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "input_touch":
		var c InputTouchCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	case "clipboard":
		var c ClipboardCommand
		err = json.Unmarshal(data, &c)
		cmd = &c
	default:
		return nil, fmt.Errorf("unknown action: %s", base.Action)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse %s command: %w", base.Action, err)
	}

	return cmd, nil
}

// SuccessResponse creates a success response.
func SuccessResponse(id string, data interface{}) Response {
	var rawData json.RawMessage
	if data != nil {
		var err error
		rawData, err = json.Marshal(data)
		if err != nil {
			return ErrorResponse(id, fmt.Sprintf("failed to marshal response data: %v", err))
		}
	}
	return Response{
		ID:      id,
		Success: true,
		Data:    rawData,
	}
}

// ErrorResponse creates an error response.
func ErrorResponse(id string, errMsg string) Response {
	return Response{
		ID:      id,
		Success: false,
		Error:   errMsg,
	}
}

// SerializeResponse serializes a response to JSON.
func SerializeResponse(resp Response) ([]byte, error) {
	return json.Marshal(resp)
}
