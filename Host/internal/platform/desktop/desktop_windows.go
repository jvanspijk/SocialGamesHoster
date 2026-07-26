//go:build windows

package desktop

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strconv"
	"sync"
	"syscall"
	"unsafe"

	"github.com/gogpu/systray"
	"golang.org/x/sys/windows"
)

const (
	cfUnicodeText = 13
	gmemMoveable  = 0x0002
)

var (
	user32               = windows.NewLazySystemDLL("user32.dll")
	kernel32             = windows.NewLazySystemDLL("kernel32.dll")
	ntdll                = windows.NewLazySystemDLL("ntdll.dll")
	procOpenClipboard    = user32.NewProc("OpenClipboard")
	procCloseClipboard   = user32.NewProc("CloseClipboard")
	procEmptyClipboard   = user32.NewProc("EmptyClipboard")
	procSetClipboardData = user32.NewProc("SetClipboardData")
	procMessageBoxW      = user32.NewProc("MessageBoxW")
	procPostQuitMessage  = user32.NewProc("PostQuitMessage")
	procGlobalAlloc      = kernel32.NewProc("GlobalAlloc")
	procGlobalLock       = kernel32.NewProc("GlobalLock")
	procGlobalUnlock     = kernel32.NewProc("GlobalUnlock")
	procGlobalFree       = kernel32.NewProc("GlobalFree")
	procRtlMoveMemory    = ntdll.NewProc("RtlMoveMemory")
)

func Supported() bool {
	return true
}

func AcquireSingleInstance(dataDir string, installerManaged bool) (bool, func(), error) {
	first, releaseData, err := acquireMutex(`Local\SocialGamesHoster-` + instanceID(dataDir))
	if err != nil || !first {
		return first, releaseData, err
	}
	if !installerManaged {
		return true, releaseData, nil
	}
	first, releaseInstaller, err := acquireMutex(`Local\SocialGamesHoster`)
	if err != nil || !first {
		releaseData()
		return first, releaseInstaller, err
	}
	return true, func() {
		releaseInstaller()
		releaseData()
	}, nil
}

func acquireMutex(value string) (bool, func(), error) {
	name, err := windows.UTF16PtrFromString(value)
	if err != nil {
		return false, func() {}, err
	}
	handle, err := windows.CreateMutex(nil, false, name)
	if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		if handle != 0 {
			_ = windows.CloseHandle(handle)
		}
		return false, func() {}, nil
	}
	if err != nil {
		return false, func() {}, err
	}
	return true, func() { _ = windows.CloseHandle(handle) }, nil
}

func OpenURL(url string) error {
	verb, _ := windows.UTF16PtrFromString("open")
	target, err := windows.UTF16PtrFromString(url)
	if err != nil {
		return err
	}
	return windows.ShellExecute(0, verb, target, nil, nil, windows.SW_SHOWNORMAL)
}

func UpdateFirewallPort(port int) error {
	verb, _ := windows.UTF16PtrFromString("runas")
	program, _ := windows.UTF16PtrFromString("netsh.exe")
	arguments, err := windows.UTF16PtrFromString(
		`advfirewall firewall set rule name="Social Games Hoster" new protocol=TCP localport=` +
			strconv.Itoa(port) + ` profile=private action=allow`,
	)
	if err != nil {
		return err
	}
	return windows.ShellExecute(0, verb, program, arguments, nil, windows.SW_HIDE)
}

func ShowError(title, message string) {
	titlePointer, _ := windows.UTF16PtrFromString(title)
	messagePointer, _ := windows.UTF16PtrFromString(message)
	procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(messagePointer)),
		uintptr(unsafe.Pointer(titlePointer)),
		0x00000010,
	)
}

func CopyText(value string) error {
	utf16, err := syscall.UTF16FromString(value)
	if err != nil {
		return err
	}
	if result, _, callErr := procOpenClipboard.Call(0); result == 0 {
		return callErr
	}
	defer procCloseClipboard.Call()
	if result, _, callErr := procEmptyClipboard.Call(); result == 0 {
		return callErr
	}
	size := uintptr(len(utf16) * 2)
	memory, _, callErr := procGlobalAlloc.Call(gmemMoveable, size)
	if memory == 0 {
		return callErr
	}
	locked, _, callErr := procGlobalLock.Call(memory)
	if locked == 0 {
		procGlobalFree.Call(memory)
		return callErr
	}
	procRtlMoveMemory.Call(locked, uintptr(unsafe.Pointer(&utf16[0])), size)
	procGlobalUnlock.Call(memory)
	if result, _, callErr := procSetClipboardData.Call(cfUnicodeText, memory); result == 0 {
		procGlobalFree.Call(memory)
		return callErr
	}
	return nil
}

func Run(actions Actions) error {
	tray := systray.New()
	var operation sync.Mutex
	menu := systray.NewMenu()
	menu.Add("Open Dashboard", func() { _ = OpenURL(actions.DashboardURL()) })
	menu.Add("Open Player Join Page", func() { _ = OpenURL(actions.JoinURL()) })
	menu.Add("Copy Join Link", func() {
		if err := CopyText(actions.JoinURL()); err != nil {
			tray.ShowNotification("Social Games Hoster", "The join link could not be copied.")
			return
		}
		tray.ShowNotification("Social Games Hoster", "Join link copied.")
	})
	menu.Add("Show QR Code", func() { _ = OpenURL(actions.JoinURL() + "?showQr=1") })
	menu.AddSeparator()
	menu.Add("Start / Stop Hosting", func() {
		go serialized(&operation, func() {
			var err error
			if actions.IsHosting() {
				err = actions.StopHosting()
				if err == nil {
					tray.ShowNotification("Social Games Hoster", "Hosting stopped.")
				}
			} else {
				err = actions.StartHosting()
				if err == nil {
					tray.ShowNotification("Social Games Hoster", "Hosting started.")
				}
			}
			if err != nil {
				tray.ShowNotification("Social Games Hoster", "The hosting state could not be changed.")
			}
		})
	})
	menu.Add("Start / Show Diagnostics", func() {
		if !actions.DiagnosticsActive {
			tray.ShowNotification("Social Games Hoster", "Use the Diagnostic Mode shortcut to enable detailed diagnostics.")
		}
		_ = OpenURL(actions.DiagnosticsURL())
	})
	menu.Add("Create Backup", func() {
		go serialized(&operation, func() {
			name, err := actions.CreateBackup()
			if err != nil {
				tray.ShowNotification("Social Games Hoster", "The backup could not be created.")
				return
			}
			tray.ShowNotification("Social Games Hoster", "Backup created: "+name)
		})
	})
	menu.AddSeparator()
	menu.Add("Exit", func() {
		tray.Remove()
		procPostQuitMessage.Call(0)
		actions.Exit()
	})

	tray.SetIcon(trayIcon()).
		SetTooltip("Social Games Hoster").
		SetMenu(menu).
		OnClick(func() { _ = OpenURL(actions.DashboardURL()) }).
		OnDoubleClick(func() { _ = OpenURL(actions.DashboardURL()) }).
		Show()
	return tray.Run()
}

func serialized(lock *sync.Mutex, operation func()) {
	lock.Lock()
	defer lock.Unlock()
	operation()
}

func trayIcon() []byte {
	canvas := image.NewRGBA(image.Rect(0, 0, 32, 32))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.RGBA{R: 248, G: 236, B: 198, A: 255}}, image.Point{}, draw.Src)
	ink := color.RGBA{R: 100, G: 27, B: 31, A: 255}
	for y := 5; y < 27; y++ {
		for x := 5; x < 27; x++ {
			dx, dy := x-16, y-16
			if dx*dx+dy*dy >= 95 && dx*dx+dy*dy <= 125 {
				canvas.Set(x, y, ink)
			}
		}
	}
	for y := 9; y < 23; y++ {
		x := 10 + (y-9)*11/13
		canvas.Set(x, y, ink)
		canvas.Set(x+1, y, ink)
		canvas.Set(31-x, y, ink)
	}
	var output bytes.Buffer
	_ = png.Encode(&output, canvas)
	return output.Bytes()
}
