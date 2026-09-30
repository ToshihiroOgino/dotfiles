//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	wmImeControl         = 0x0283 // WM_IME_CONTROL
	imcGetConversionMode = 0x0001 // IMC_GETCONVERSIONMODE
	imcGetOpenStatus     = 0x0005 // IMC_GETOPENSTATUS
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

// query sends an IMC_* request to the IME window. ok is false on timeout or failure.
func query(hwnd, cmd uintptr) (result uintptr, ok bool) {
	r, _, _ := sendMsg.Call(hwnd, wmImeControl, cmd, 0,
		smtoAbortIfHung, timeoutMs, uintptr(unsafe.Pointer(&result)))
	return result, r != 0
}

func imeState() (string, error) {
	fg, _, _ := getFg.Call()
	if fg == 0 {
		return "?", nil
	}
	ime, _, _ := imeWnd.Call(fg)
	if ime == 0 {
		return "?", nil
	}
	open, ok1 := query(ime, imcGetOpenStatus)
	mode, ok2 := query(ime, imcGetConversionMode)
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
