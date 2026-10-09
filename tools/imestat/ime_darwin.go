//go:build darwin

package main

/*
#cgo LDFLAGS: -framework Carbon -framework CoreFoundation
#include <Carbon/Carbon.h>
#include <string.h>

// imeMode returns 1 for a native Japanese mode, 0 for alphanumeric, -1 on failure.
static int imeMode(void) {
	// Let the run loop apply pending input source changes; without it a
	// long-running process keeps seeing the value from its first call.
	CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0, true);

	TISInputSourceRef src = TISCopyCurrentKeyboardInputSource();
	if (src == NULL) {
		return -1;
	}
	int result = 0;
	CFStringRef mode = TISGetInputSourceProperty(src, kTISPropertyInputModeID);
	if (mode != NULL) {
		char buf[256];
		if (!CFStringGetCString(mode, buf, sizeof buf, kCFStringEncodingUTF8)) {
			CFRelease(src);
			return -1;
		}
		// Input methods expose a mode per state, e.g. com.apple.inputmethod.Japanese
		// (hiragana) vs com.apple.inputmethod.Roman. Plain keyboard layouts have none.
		size_t n = strlen(buf), suffix = strlen(".Roman");
		result = !(n >= suffix && strcmp(buf + n - suffix, ".Roman") == 0);
	}
	CFRelease(src);
	return result;
}
*/
import "C"

import "runtime"

func init() {
	// TIS must be called from the main thread; keep the main goroutine on it.
	runtime.LockOSThread()
}

func imeState() (string, error) {
	switch C.imeMode() {
	case 1:
		return "あ", nil
	case 0:
		return "A", nil
	}
	return "?", nil
}
