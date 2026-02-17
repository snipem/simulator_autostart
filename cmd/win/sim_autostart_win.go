package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"simulator_autostart/lib/autostart"
)

func init() {
	// Win32 GUI must run on the main OS thread.
	runtime.LockOSThread()
}

// Win32 API DLLs
var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
)

// user32 functions
var (
	procRegisterClassExW  = user32.NewProc("RegisterClassExW")
	procCreateWindowExW   = user32.NewProc("CreateWindowExW")
	procDefWindowProcW    = user32.NewProc("DefWindowProcW")
	procGetMessageW       = user32.NewProc("GetMessageW")
	procTranslateMessage  = user32.NewProc("TranslateMessage")
	procDispatchMessageW  = user32.NewProc("DispatchMessageW")
	procPostMessageW      = user32.NewProc("PostMessageW")
	procSendMessageW      = user32.NewProc("SendMessageW")
	procShowWindow        = user32.NewProc("ShowWindow")
	procSetForegroundWnd  = user32.NewProc("SetForegroundWindow")
	procMoveWindow        = user32.NewProc("MoveWindow")
	procLoadIconW         = user32.NewProc("LoadIconW")
	procLoadCursorW       = user32.NewProc("LoadCursorW")
	procPostQuitMessage   = user32.NewProc("PostQuitMessage")
	procDestroyWindow     = user32.NewProc("DestroyWindow")
	procCreatePopupMenu   = user32.NewProc("CreatePopupMenu")
	procAppendMenuW       = user32.NewProc("AppendMenuW")
	procTrackPopupMenu    = user32.NewProc("TrackPopupMenu")
	procDestroyMenu       = user32.NewProc("DestroyMenu")
	procGetCursorPos      = user32.NewProc("GetCursorPos")
	procInvalidateRect    = user32.NewProc("InvalidateRect")
	procGetModuleHandleW  = kernel32.NewProc("GetModuleHandleW")
	procLoadLibraryW      = kernel32.NewProc("LoadLibraryW")
	procCloseHandle       = kernel32.NewProc("CloseHandle")
	procGetStockObject    = gdi32.NewProc("GetStockObject")
	procCreateFontW       = gdi32.NewProc("CreateFontW")
	procDeleteObject      = gdi32.NewProc("DeleteObject")
	procShellNotifyIconW  = shell32.NewProc("Shell_NotifyIconW")
	procExtractIconW      = shell32.NewProc("ExtractIconW")
)

var hTrayIcon syscall.Handle

// Win32 constants
const (
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_VISIBLE          = 0x10000000
	WS_CHILD            = 0x40000000
	WS_VSCROLL          = 0x00200000
	WS_EX_CLIENTEDGE    = 0x00000200

	ES_MULTILINE   = 0x0004
	ES_AUTOVSCROLL = 0x0040
	ES_READONLY    = 0x0800

	CW_USEDEFAULT = 0x80000000

	SW_HIDE    = 0
	SW_RESTORE = 9

	WM_CREATE  = 0x0001
	WM_DESTROY = 0x0002
	WM_SIZE    = 0x0005
	WM_CLOSE   = 0x0010
	WM_SETFONT    = 0x0030
	WM_SETREDRAW  = 0x000B
	WM_APP     = 0x8000

	WM_LBUTTONUP = 0x0202
	WM_RBUTTONUP = 0x0205

	WM_COMMAND = 0x0111

	EM_SETSEL          = 0x00B1
	EM_SCROLLCARET     = 0x00B7
	EM_REPLACESEL      = 0x00C2
	EM_SETCHARFORMAT   = 0x0444

	SCF_SELECTION = 0x0001
	CFM_COLOR     = 0x40000000

	IDI_APPLICATION = 32512
	IDC_ARROW       = 32512

	WHITE_BRUSH = 0

	DEFAULT_CHARSET     = 1
	OUT_DEFAULT_PRECIS  = 0
	CLIP_DEFAULT_PRECIS = 0
	DEFAULT_QUALITY     = 0
	FIXED_PITCH         = 1
	FF_DONTCARE         = 0
	LF_FACESIZE         = 32

	NIM_ADD     = 0x00000000
	NIM_DELETE  = 0x00000002
	NIF_MESSAGE = 0x00000001
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004

	MF_STRING       = 0x00000000
	TPM_BOTTOMALIGN = 0x0020
	TPM_LEFTALIGN   = 0x0000

	IDM_EXIT = 1001
	IDM_SHOW = 1000

	WM_APP_TRAY_MESSAGE = WM_APP + 1
	WM_APP_ADD_LOG      = WM_APP + 2
)

// COLORREF values (0x00BBGGRR)
const (
	colorBlack  = 0x00000000
	colorRed    = 0x000000CC
	colorGreen  = 0x00008800
	colorYellow = 0x00008888
	colorBlue   = 0x00CC6600
	colorCyan   = 0x00888800
)

// Win32 structs
type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     syscall.Handle
	HIcon         syscall.Handle
	HCursor       syscall.Handle
	HbrBackground syscall.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       syscall.Handle
}

type MSG struct {
	Hwnd    syscall.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type POINT struct {
	X, Y int32
}

type NOTIFYICONDATA struct {
	CbSize           uint32
	HWnd             syscall.Handle
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            syscall.Handle
	SzTip            [128]uint16
}

type CHARFORMATW struct {
	CbSize          uint32
	DwMask          uint32
	DwEffects       uint32
	YHeight         int32
	YOffset         int32
	CrTextColor     uint32
	BCharSet        byte
	BPitchAndFamily byte
	SzFaceName      [LF_FACESIZE]uint16
}

type logEntry struct {
	text  string
	color uint32
}

// Globals
var (
	hwndMain   syscall.Handle
	hEditLog   syscall.Handle
	hFont      syscall.Handle
	logMutex   sync.Mutex
	pendingLog []logEntry
	engine     *autostart.Engine
)

// Tag-to-color mapping (matches console ANSI colors)
var tagColors = map[string]uint32{
	"[config]": colorBlue,
	"[skip]":   colorYellow,
	"[start]":  colorGreen,
	"[error]":  colorRed,
	"[reload]": colorYellow,
	"[watch]":  colorCyan,
}

func colorForMessage(msg string) uint32 {
	for tag, color := range tagColors {
		if strings.HasPrefix(msg, tag) {
			return color
		}
	}
	return colorBlack
}

func getModuleHandle() syscall.Handle {
	ret, _, _ := procGetModuleHandleW.Call(0)
	return syscall.Handle(ret)
}

func loadJoystickIcon() syscall.Handle {
	joyCpl, _ := syscall.UTF16PtrFromString(`C:\Windows\System32\joy.cpl`)
	ret, _, _ := procExtractIconW.Call(
		uintptr(getModuleHandle()),
		uintptr(unsafe.Pointer(joyCpl)),
		0, // icon index 0 = joystick
	)
	if ret > 1 { // ExtractIcon returns >1 on success, 0 or 1 on failure
		return syscall.Handle(ret)
	}
	return loadIcon(0, makeIntResource(IDI_APPLICATION)) // fallback
}

func loadIcon(instance syscall.Handle, iconName uintptr) syscall.Handle {
	ret, _, _ := procLoadIconW.Call(uintptr(instance), iconName)
	return syscall.Handle(ret)
}

func loadCursor(instance syscall.Handle, cursorName uintptr) syscall.Handle {
	ret, _, _ := procLoadCursorW.Call(uintptr(instance), cursorName)
	return syscall.Handle(ret)
}

func loword(v uint32) uint16 { return uint16(v) }
func hiword(v uint32) uint16 { return uint16(v >> 16) }

func makeIntResource(id uint16) uintptr {
	return uintptr(id)
}

// AddLog appends a log message and signals the UI to update.
func AddLog(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	line := time.Now().Format("15:04:05") + " " + msg
	color := colorForMessage(msg)

	logMutex.Lock()
	pendingLog = append(pendingLog, logEntry{text: line, color: color})
	logMutex.Unlock()

	if hwndMain != 0 {
		procPostMessageW.Call(uintptr(hwndMain), WM_APP_ADD_LOG, 0, 0)
	}
}

func appendColoredText(text string, color uint32) {
	// Move selection to end of text
	procSendMessageW.Call(uintptr(hEditLog), EM_SETSEL, ^uintptr(0), ^uintptr(0))

	// Set text color for the selection
	cf := CHARFORMATW{
		DwMask:      CFM_COLOR,
		CrTextColor: color,
	}
	cf.CbSize = uint32(unsafe.Sizeof(cf))
	procSendMessageW.Call(uintptr(hEditLog), EM_SETCHARFORMAT, SCF_SELECTION, uintptr(unsafe.Pointer(&cf)))

	// Insert the text at the selection (end)
	ptr, _ := syscall.UTF16PtrFromString(text + "\r\n")
	procSendMessageW.Call(uintptr(hEditLog), EM_REPLACESEL, 0, uintptr(unsafe.Pointer(ptr)))
}

func updateLogWindow() {
	logMutex.Lock()
	entries := pendingLog
	pendingLog = nil
	logMutex.Unlock()

	if len(entries) == 0 {
		return
	}

	// Suspend redraws while appending
	procSendMessageW.Call(uintptr(hEditLog), WM_SETREDRAW, 0, 0)

	for _, entry := range entries {
		appendColoredText(entry.text, entry.color)
	}

	// Resume redraws and scroll to end
	procSendMessageW.Call(uintptr(hEditLog), WM_SETREDRAW, 1, 0)
	procInvalidateRect.Call(uintptr(hEditLog), 0, 1)
	procSendMessageW.Call(uintptr(hEditLog), EM_SETSEL, ^uintptr(0), ^uintptr(0))
	procSendMessageW.Call(uintptr(hEditLog), EM_SCROLLCARET, 0, 0)
}

func setTrayIcon(add bool) {
	nid := NOTIFYICONDATA{}
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.HWnd = hwndMain
	nid.UID = 1
	nid.UFlags = NIF_MESSAGE | NIF_ICON | NIF_TIP
	nid.UCallbackMessage = WM_APP_TRAY_MESSAGE
	nid.HIcon = hTrayIcon

	tip, _ := syscall.UTF16FromString("Simulator Autostart")
	copy(nid.SzTip[:], tip)

	var action uint32
	if add {
		action = NIM_ADD
	} else {
		action = NIM_DELETE
	}
	procShellNotifyIconW.Call(uintptr(action), uintptr(unsafe.Pointer(&nid)))
}

func startEngine() {
	engine = autostart.NewEngine(AddLog)
	engine.LoadConfig()
	engine.WatchConfigFile()

	go func() {
		for {
			engine.RunOnce()
			time.Sleep(5 * time.Second)
		}
	}()
}

func showTrayMenu(hwnd syscall.Handle) {
	hMenu, _, _ := procCreatePopupMenu.Call()
	showLabel, _ := syscall.UTF16PtrFromString("Show Log")
	exitLabel, _ := syscall.UTF16PtrFromString("Exit")
	procAppendMenuW.Call(hMenu, MF_STRING, IDM_SHOW, uintptr(unsafe.Pointer(showLabel)))
	procAppendMenuW.Call(hMenu, MF_STRING, IDM_EXIT, uintptr(unsafe.Pointer(exitLabel)))

	var pt POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	// SetForegroundWindow is required before TrackPopupMenu so the menu dismisses properly
	procSetForegroundWnd.Call(uintptr(hwnd))
	procTrackPopupMenu.Call(hMenu, TPM_BOTTOMALIGN|TPM_LEFTALIGN, uintptr(pt.X), uintptr(pt.Y), 0, uintptr(hwnd), 0)
	procDestroyMenu.Call(hMenu)
}

func wndProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_CREATE:
		className, _ := syscall.UTF16PtrFromString("RICHEDIT50W")
		emptyStr, _ := syscall.UTF16PtrFromString("")

		ret, _, _ := procCreateWindowExW.Call(
			WS_EX_CLIENTEDGE,
			uintptr(unsafe.Pointer(className)),
			uintptr(unsafe.Pointer(emptyStr)),
			WS_VISIBLE|WS_CHILD|ES_MULTILINE|ES_AUTOVSCROLL|ES_READONLY|WS_VSCROLL,
			0, 0, 400, 300,
			uintptr(hwnd),
			0,
			uintptr(getModuleHandle()),
			0,
		)
		hEditLog = syscall.Handle(ret)
		if hEditLog == 0 {
			log.Fatal("Failed to create rich edit control")
		}

		fontName, _ := syscall.UTF16PtrFromString("Consolas")
		hf, _, _ := procCreateFontW.Call(
			uintptr(uint32(0xFFFFFFF2)), // -14 as unsigned
			0, 0, 0, 0,
			0, 0, 0,
			DEFAULT_CHARSET,
			OUT_DEFAULT_PRECIS,
			CLIP_DEFAULT_PRECIS,
			DEFAULT_QUALITY,
			FIXED_PITCH|FF_DONTCARE,
			uintptr(unsafe.Pointer(fontName)),
		)
		hFont = syscall.Handle(hf)
		procSendMessageW.Call(uintptr(hEditLog), WM_SETFONT, uintptr(hFont), 0)

		procShowWindow.Call(uintptr(hwnd), SW_HIDE)
		hwndMain = hwnd
		setTrayIcon(true)

		AddLog("simulator_autostart %s started (GUI mode)", autostart.VERSION)
		AddLog("Click the tray icon to show the window.")

		startEngine()

	case WM_SIZE:
		w := loword(uint32(lParam))
		h := hiword(uint32(lParam))
		procMoveWindow.Call(uintptr(hEditLog), 0, 0, uintptr(w), uintptr(h), 1)

	case WM_COMMAND:
		cmdID := loword(uint32(wParam))
		switch cmdID {
		case IDM_SHOW:
			procShowWindow.Call(uintptr(hwnd), SW_RESTORE)
			procSetForegroundWnd.Call(uintptr(hwnd))
		case IDM_EXIT:
			procDestroyWindow.Call(uintptr(hwnd))
		}

	case WM_APP_TRAY_MESSAGE:
		switch loword(uint32(lParam)) {
		case WM_LBUTTONUP:
			procShowWindow.Call(uintptr(hwnd), SW_RESTORE)
			procSetForegroundWnd.Call(uintptr(hwnd))
		case WM_RBUTTONUP:
			showTrayMenu(hwnd)
		}

	case WM_APP_ADD_LOG:
		updateLogWindow()

	case WM_CLOSE:
		procShowWindow.Call(uintptr(hwnd), SW_HIDE)
		return 0

	case WM_DESTROY:
		setTrayIcon(false)
		procDeleteObject.Call(uintptr(hFont))
		procPostQuitMessage.Call(0)
	}

	ret, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return ret
}

func main() {
	if autostart.IsAnotherInstanceRunning() {
		return
	}

	// Load RichEdit 5.0 library
	libName, _ := syscall.UTF16PtrFromString("Msftedit.dll")
	procLoadLibraryW.Call(uintptr(unsafe.Pointer(libName)))

	hInstance := getModuleHandle()
	hTrayIcon = loadJoystickIcon()
	className, _ := syscall.UTF16PtrFromString("SimAutostart")

	stockBrush, _, _ := procGetStockObject.Call(WHITE_BRUSH)

	wc := WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEX{})),
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HInstance:     hInstance,
		HIcon:         hTrayIcon,
		HCursor:       loadCursor(0, makeIntResource(IDC_ARROW)),
		HbrBackground: syscall.Handle(stockBrush),
		LpszClassName: className,
	}

	ret, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if ret == 0 {
		log.Fatal("Failed to register window class")
	}

	windowTitle, _ := syscall.UTF16PtrFromString("Simulator Autostart")
	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowTitle)),
		WS_OVERLAPPEDWINDOW,
		CW_USEDEFAULT, CW_USEDEFAULT,
		800, 600,
		0, 0,
		uintptr(hInstance),
		0,
	)
	if hwnd == 0 {
		log.Fatal("Failed to create window")
	}

	// Message loop
	var msg MSG
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}
