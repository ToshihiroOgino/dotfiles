//go:build darwin

package imectl

/*
#cgo LDFLAGS: -framework Carbon -framework CoreFoundation
#include <Carbon/Carbon.h>
#include <stdio.h>
#include <stdlib.h>
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

// selectBy selects the enabled input source whose property key equals val.
static int selectBy(CFStringRef key, CFStringRef val) {
	const void *keys[] = { key };
	const void *vals[] = { val };
	CFDictionaryRef filter = CFDictionaryCreate(NULL, keys, vals, 1,
		&kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);
	CFArrayRef list = TISCreateInputSourceList(filter, false);
	CFRelease(filter);
	if (list == NULL) {
		return -1;
	}
	int rc = -1;
	if (CFArrayGetCount(list) > 0) {
		TISInputSourceRef src = (TISInputSourceRef)CFArrayGetValueAtIndex(list, 0);
		rc = TISSelectInputSource(src) == noErr ? 0 : -1;
	}
	CFRelease(list);
	return rc;
}

// imeSelect selects the ASCII-capable layout (ja=0) or the Japanese hiragana
// mode (ja=1). Returns 0 on success.
static int imeSelect(int ja) {
	if (ja) {
		return selectBy(kTISPropertyInputModeID, CFSTR("com.apple.inputmethod.Japanese"));
	}
	TISInputSourceRef src = TISCopyCurrentASCIICapableKeyboardInputSource();
	if (src == NULL) {
		return -1;
	}
	OSStatus st = TISSelectInputSource(src);
	CFRelease(src);
	return st == noErr ? 0 : -1;
}

// imeToken writes an identifier of the current input source: "mode:<id>" for
// an input method mode, "id:<id>" for a plain layout. Returns 0 on success.
static int imeToken(char *out, size_t n) {
	CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0, true);
	TISInputSourceRef src = TISCopyCurrentKeyboardInputSource();
	if (src == NULL) {
		return -1;
	}
	const char *prefix = "mode:";
	CFStringRef s = TISGetInputSourceProperty(src, kTISPropertyInputModeID);
	if (s == NULL) {
		prefix = "id:";
		s = TISGetInputSourceProperty(src, kTISPropertyInputSourceID);
	}
	char id[256];
	int rc = -1;
	if (s != NULL && CFStringGetCString(s, id, sizeof id, kCFStringEncodingUTF8)) {
		rc = snprintf(out, n, "%s%s", prefix, id) < (int)n ? 0 : -1;
	}
	CFRelease(src);
	return rc;
}

// imeRestore selects the input source named by a token from imeToken.
static int imeRestore(const char *token) {
	CFStringRef key;
	if (strncmp(token, "mode:", 5) == 0) {
		key = kTISPropertyInputModeID;
		token += 5;
	} else if (strncmp(token, "id:", 3) == 0) {
		key = kTISPropertyInputSourceID;
		token += 3;
	} else {
		return -1;
	}
	CFStringRef val = CFStringCreateWithCString(NULL, token, kCFStringEncodingUTF8);
	if (val == NULL) {
		return -1;
	}
	int rc = selectBy(key, val);
	CFRelease(val);
	return rc;
}
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

func init() {
	// TIS must be called from the main thread; keep the main goroutine on it.
	runtime.LockOSThread()
}

func State() (string, error) {
	switch C.imeMode() {
	case 1:
		return "あ", nil
	case 0:
		return "A", nil
	}
	return "?", nil
}

func Set(native bool) error {
	ja := C.int(0)
	if native {
		ja = 1
	}
	if C.imeSelect(ja) != 0 {
		return fmt.Errorf("failed to select input source")
	}
	return nil
}

// Save returns an opaque token for the current input source.
func Save() (string, error) {
	var buf [256]C.char
	if C.imeToken(&buf[0], C.size_t(len(buf))) != 0 {
		return "", fmt.Errorf("failed to read input source")
	}
	return C.GoString(&buf[0]), nil
}

// Restore switches back to the input source a Save token names.
func Restore(token string) error {
	cs := C.CString(token)
	defer C.free(unsafe.Pointer(cs))
	if C.imeRestore(cs) != 0 {
		return fmt.Errorf("failed to restore input source %q", token)
	}
	return nil
}
