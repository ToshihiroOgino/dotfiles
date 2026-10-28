//go:build windows

package imectl

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	wmImeControl         = 0x0283 // WM_IME_CONTROL
	imcGetConversionMode = 0x0001 // IMC_GETCONVERSIONMODE
	imcGetOpenStatus     = 0x0005 // IMC_GETOPENSTATUS
	imcSetConversionMode = 0x0002 // IMC_SETCONVERSIONMODE
	imcSetOpenStatus     = 0x0006 // IMC_SETOPENSTATUS
	imeCmodeFullShape    = 0x0008 // IME_CMODE_FULLSHAPE
	imeCmodeNative       = 0x0001 // IME_CMODE_NATIVE (kana input)
	smtoAbortIfHung      = 0x0002 // SMTO_ABORTIFHUNG
	timeoutMs            = 50
)

var (
	user32  = syscall.NewLazyDLL("user32.dll")
	imm32   = syscall.NewLazyDLL("imm32.dll")
	getFg   = user32.NewProc("GetForegroundWindow")
	sendMsg = user32.NewProc("SendMessageTimeoutW")
	imeWnd  = imm32.NewProc("ImmGetDefaultIMEWnd")
)

func State() (string, error) {
	ime := foregroundIME()
	if ime == 0 {
		return "?", nil
	}
	open, ok1 := sendIMC(ime, imcGetOpenStatus, 0)
	mode, ok2 := sendIMC(ime, imcGetConversionMode, 0)
	if !ok1 || !ok2 {
		return "?", nil
	}
	// A modern MS-IME can report "open" while in half-width alphanumeric mode,
	// so also require the native (kana) conversion mode.
	if open != 0 && mode&imeCmodeNative != 0 {
		return "あ", nil
	}
	return "A", nil
}

// Set closes the IME for alphanumeric input, or opens it in hiragana
// (native + full-width) mode.
func Set(native bool) error {
	ime := foregroundIME()
	if ime == 0 {
		return fmt.Errorf("no IME window")
	}
	open, mode := uintptr(0), uintptr(0)
	if native {
		open, mode = 1, imeCmodeNative|imeCmodeFullShape
	}
	_, ok1 := sendIMC(ime, imcSetOpenStatus, open)
	ok2 := true
	if native {
		_, ok2 = sendIMC(ime, imcSetConversionMode, mode)
	}
	if !ok1 || !ok2 {
		return fmt.Errorf("IME did not respond")
	}
	return nil
}

// sendIMC sends an IMC_* request to the IME window. ok is false on timeout or failure.
func sendIMC(hwnd, cmd, arg uintptr) (result uintptr, ok bool) {
	r, _, _ := sendMsg.Call(hwnd, wmImeControl, cmd, arg,
		smtoAbortIfHung, timeoutMs, uintptr(unsafe.Pointer(&result)))
	return result, r != 0
}

// foregroundIME returns the IME window of the foreground window, or 0.
func foregroundIME() uintptr {
	fg, _, _ := getFg.Call()
	if fg == 0 {
		return 0
	}
	ime, _, _ := imeWnd.Call(fg)
	return ime
}

// Save returns the open status and conversion mode as "open:mode".
func Save() (string, error) {
	ime := foregroundIME()
	if ime == 0 {
		return "", fmt.Errorf("no IME window")
	}
	open, ok1 := sendIMC(ime, imcGetOpenStatus, 0)
	mode, ok2 := sendIMC(ime, imcGetConversionMode, 0)
	if !ok1 || !ok2 {
		return "", fmt.Errorf("IME did not respond")
	}
	return fmt.Sprintf("%d:%d", open, mode), nil
}

// Restore puts back the open status and conversion mode of a Save token.
func Restore(token string) error {
	var open, mode uintptr
	if _, err := fmt.Sscanf(token, "%d:%d", &open, &mode); err != nil {
		return fmt.Errorf("bad token %q", token)
	}
	ime := foregroundIME()
	if ime == 0 {
		return fmt.Errorf("no IME window")
	}
	_, ok1 := sendIMC(ime, imcSetOpenStatus, open)
	ok2 := true
	if open != 0 {
		_, ok2 = sendIMC(ime, imcSetConversionMode, mode)
	}
	if !ok1 || !ok2 {
		return fmt.Errorf("IME did not respond")
	}
	return nil
}
