//go:build darwin && cgo

package keyring

/*
#cgo LDFLAGS: -framework CoreFoundation -framework Security

#include <CoreFoundation/CoreFoundation.h>
#include <Security/Security.h>
#include <stdlib.h>
#include <string.h>

// secure_free zeroes out memory before freeing to prevent plaintext lingering
// in freed heap pages (memory dumps, core dumps). Uses volatile pointer to
// prevent compiler from optimizing away the zeroing.
static void secure_free(char *ptr, int len) {
	if (ptr) {
		volatile char *vp = (volatile char *)ptr;
		for (int i = 0; i < len; i++) vp[i] = 0;
		free(ptr);
	}
}

// KeychainResult is returned by C helper functions.
typedef struct {
	int    status;    // 0 = success, -1 = not found, >0 = OSStatus error
	char  *data;      // returned data (caller must free)
	int    data_len;
} KeychainResult;

static CFStringRef _toCFString(const char *s) {
	return CFStringCreateWithCString(kCFAllocatorDefault, s, kCFStringEncodingUTF8);
}

// keychain_biometric_available checks if biometric access control can be created
// and verifies that the login Keychain actually accepts biometric-protected items.
// This probe does NOT trigger a Touch ID prompt (writes bypass ACL evaluation).
static int keychain_biometric_available(void) {
	// 1. Check if biometric access control flags are supported.
	CFErrorRef error = NULL;
	SecAccessControlRef access = SecAccessControlCreateWithFlags(
		kCFAllocatorDefault,
		kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly,
		kSecAccessControlBiometryCurrentSet,
		&error
	);
	if (error != NULL) {
		CFRelease(error);
		return 0;
	}

	// 2. Probe the login Keychain with a real SecItemAdd to verify entitlements.
	CFStringRef svc  = CFSTR("lango-probe");
	CFStringRef acct = CFSTR("biometric-check");
	CFDataRef   val  = CFDataCreate(kCFAllocatorDefault, (const UInt8 *)"p", 1);

	// Clean up any leftover probe item.
	CFMutableDictionaryRef del = CFDictionaryCreateMutable(
		kCFAllocatorDefault, 0,
		&kCFTypeDictionaryKeyCallBacks,
		&kCFTypeDictionaryValueCallBacks);
	CFDictionarySetValue(del, kSecClass,                    kSecClassGenericPassword);
	CFDictionarySetValue(del, kSecAttrService,              svc);
	CFDictionarySetValue(del, kSecAttrAccount,              acct);
	CFDictionarySetValue(del, kSecUseDataProtectionKeychain, kCFBooleanFalse);
	SecItemDelete(del);
	CFRelease(del);

	// Attempt to add a probe item with biometric ACL.
	CFMutableDictionaryRef add = CFDictionaryCreateMutable(
		kCFAllocatorDefault, 0,
		&kCFTypeDictionaryKeyCallBacks,
		&kCFTypeDictionaryValueCallBacks);
	CFDictionarySetValue(add, kSecClass,                    kSecClassGenericPassword);
	CFDictionarySetValue(add, kSecAttrService,              svc);
	CFDictionarySetValue(add, kSecAttrAccount,              acct);
	CFDictionarySetValue(add, kSecValueData,                val);
	CFDictionarySetValue(add, kSecAttrAccessControl,        access);
	CFDictionarySetValue(add, kSecUseDataProtectionKeychain, kCFBooleanFalse);

	OSStatus status = SecItemAdd(add, NULL);

	// Clean up probe item on success.
	if (status == errSecSuccess) {
		CFMutableDictionaryRef cleanup = CFDictionaryCreateMutable(
			kCFAllocatorDefault, 0,
			&kCFTypeDictionaryKeyCallBacks,
			&kCFTypeDictionaryValueCallBacks);
		CFDictionarySetValue(cleanup, kSecClass,                    kSecClassGenericPassword);
		CFDictionarySetValue(cleanup, kSecAttrService,              svc);
		CFDictionarySetValue(cleanup, kSecAttrAccount,              acct);
		CFDictionarySetValue(cleanup, kSecUseDataProtectionKeychain, kCFBooleanFalse);
		SecItemDelete(cleanup);
		CFRelease(cleanup);
	}

	CFRelease(add);
	CFRelease(access);
	CFRelease(val);

	return (status == errSecSuccess) ? 1 : 0;
}

// keychain_set_biometric stores a value with biometric (Touch ID) access control.
// Uses the login Keychain (kSecUseDataProtectionKeychain = false) so that
// ad-hoc signed binaries work without keychain-access-groups entitlement.
static KeychainResult keychain_set_biometric(const char *service, const char *account,
                                              const char *value, int value_len) {
	KeychainResult result = {0, NULL, 0};

	CFStringRef cfService = _toCFString(service);
	CFStringRef cfAccount = _toCFString(account);
	CFDataRef   cfValue   = CFDataCreate(kCFAllocatorDefault, (const UInt8 *)value, value_len);

	// Delete any existing item first (ignore errors).
	CFMutableDictionaryRef delQuery = CFDictionaryCreateMutable(
		kCFAllocatorDefault, 0,
		&kCFTypeDictionaryKeyCallBacks,
		&kCFTypeDictionaryValueCallBacks);
	CFDictionarySetValue(delQuery, kSecClass,                    kSecClassGenericPassword);
	CFDictionarySetValue(delQuery, kSecAttrService,              cfService);
	CFDictionarySetValue(delQuery, kSecAttrAccount,              cfAccount);
	CFDictionarySetValue(delQuery, kSecUseDataProtectionKeychain, kCFBooleanFalse);
	SecItemDelete(delQuery);
	CFRelease(delQuery);

	// Create biometric access control targeting login Keychain.
	// BiometryCurrentSet: invalidates item when biometric enrollment changes.
	CFErrorRef acError = NULL;
	SecAccessControlRef access = SecAccessControlCreateWithFlags(
		kCFAllocatorDefault,
		kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly,
		kSecAccessControlBiometryCurrentSet,
		&acError);
	if (acError != NULL) {
		result.status = (int)CFErrorGetCode(acError);
		CFRelease(acError);
		CFRelease(cfService);
		CFRelease(cfAccount);
		CFRelease(cfValue);
		return result;
	}

	// Add item with biometric protection to login Keychain.
	CFMutableDictionaryRef query = CFDictionaryCreateMutable(
		kCFAllocatorDefault, 0,
		&kCFTypeDictionaryKeyCallBacks,
		&kCFTypeDictionaryValueCallBacks);
	CFDictionarySetValue(query, kSecClass,                    kSecClassGenericPassword);
	CFDictionarySetValue(query, kSecAttrService,              cfService);
	CFDictionarySetValue(query, kSecAttrAccount,              cfAccount);
	CFDictionarySetValue(query, kSecValueData,                cfValue);
	CFDictionarySetValue(query, kSecAttrAccessControl,        access);
	CFDictionarySetValue(query, kSecUseDataProtectionKeychain, kCFBooleanFalse);

	OSStatus status = SecItemAdd(query, NULL);
	result.status = (int)status;

	CFRelease(query);
	CFRelease(access);
	CFRelease(cfService);
	CFRelease(cfAccount);
	CFRelease(cfValue);
	return result;
}

// keychain_get_biometric retrieves a value; triggers Touch ID prompt.
// Targets the login Keychain explicitly.
static KeychainResult keychain_get_biometric(const char *service, const char *account) {
	KeychainResult result = {0, NULL, 0};

	CFStringRef cfService = _toCFString(service);
	CFStringRef cfAccount = _toCFString(account);

	CFMutableDictionaryRef query = CFDictionaryCreateMutable(
		kCFAllocatorDefault, 0,
		&kCFTypeDictionaryKeyCallBacks,
		&kCFTypeDictionaryValueCallBacks);
	CFDictionarySetValue(query, kSecClass,                    kSecClassGenericPassword);
	CFDictionarySetValue(query, kSecAttrService,              cfService);
	CFDictionarySetValue(query, kSecAttrAccount,              cfAccount);
	CFDictionarySetValue(query, kSecMatchLimit,               kSecMatchLimitOne);
	CFDictionarySetValue(query, kSecReturnData,               kCFBooleanTrue);
	CFDictionarySetValue(query, kSecUseDataProtectionKeychain, kCFBooleanFalse);

	CFTypeRef item = NULL;
	OSStatus status = SecItemCopyMatching(query, &item);

	if (status == errSecItemNotFound) {
		result.status = -1;
	} else if (status != errSecSuccess) {
		result.status = (int)status;
	} else {
		CFDataRef data = (CFDataRef)item;
		CFIndex   len  = CFDataGetLength(data);
		result.data     = (char *)malloc(len);
		memcpy(result.data, CFDataGetBytePtr(data), len);
		result.data_len = (int)len;
		CFRelease(item);
	}

	CFRelease(query);
	CFRelease(cfService);
	CFRelease(cfAccount);
	return result;
}

// keychain_has_biometric checks if an item exists WITHOUT triggering Touch ID.
// Queries for attributes only (not data), so biometric ACL is not enforced.
// Targets the login Keychain explicitly.
static int keychain_has_biometric(const char *service, const char *account) {
	CFStringRef cfService = _toCFString(service);
	CFStringRef cfAccount = _toCFString(account);

	CFMutableDictionaryRef query = CFDictionaryCreateMutable(
		kCFAllocatorDefault, 0,
		&kCFTypeDictionaryKeyCallBacks,
		&kCFTypeDictionaryValueCallBacks);
	CFDictionarySetValue(query, kSecClass,                    kSecClassGenericPassword);
	CFDictionarySetValue(query, kSecAttrService,              cfService);
	CFDictionarySetValue(query, kSecAttrAccount,              cfAccount);
	CFDictionarySetValue(query, kSecMatchLimit,               kSecMatchLimitOne);
	CFDictionarySetValue(query, kSecReturnAttributes,         kCFBooleanTrue);
	CFDictionarySetValue(query, kSecUseDataProtectionKeychain, kCFBooleanFalse);

	CFTypeRef item = NULL;
	OSStatus status = SecItemCopyMatching(query, &item);

	if (item != NULL) CFRelease(item);
	CFRelease(query);
	CFRelease(cfService);
	CFRelease(cfAccount);

	return (status == errSecSuccess) ? 1 : 0;
}

// keychain_delete_biometric deletes the item from the login Keychain.
static int keychain_delete_biometric(const char *service, const char *account) {
	CFStringRef cfService = _toCFString(service);
	CFStringRef cfAccount = _toCFString(account);

	CFMutableDictionaryRef query = CFDictionaryCreateMutable(
		kCFAllocatorDefault, 0,
		&kCFTypeDictionaryKeyCallBacks,
		&kCFTypeDictionaryValueCallBacks);
	CFDictionarySetValue(query, kSecClass,                    kSecClassGenericPassword);
	CFDictionarySetValue(query, kSecAttrService,              cfService);
	CFDictionarySetValue(query, kSecAttrAccount,              cfAccount);
	CFDictionarySetValue(query, kSecUseDataProtectionKeychain, kCFBooleanFalse);

	OSStatus status = SecItemDelete(query);

	CFRelease(query);
	CFRelease(cfService);
	CFRelease(cfAccount);

	if (status == errSecItemNotFound) {
		return -1;
	}
	return (int)status;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// BiometricProvider stores secrets in the macOS login Keychain with Touch ID
// (biometric) protection via kSecAccessControlBiometryCurrentSet. Items require
// biometric authentication for retrieval. Uses the login Keychain instead of
// the Data Protection Keychain so that ad-hoc signed binaries work without
// keychain-access-groups entitlement. Biometric enrollment changes invalidate
// stored items (BiometryCurrentSet), providing stronger security than BiometryAny.
type BiometricProvider struct{}

var _ Provider = (*BiometricProvider)(nil)
var _ KeyChecker = (*BiometricProvider)(nil)

// NewBiometricProvider creates a new BiometricProvider.
// Returns ErrBiometricNotAvailable if Touch ID hardware is not available.
func NewBiometricProvider() (*BiometricProvider, error) {
	if C.keychain_biometric_available() == 0 {
		return nil, ErrBiometricNotAvailable
	}
	return &BiometricProvider{}, nil
}

// Get retrieves a secret from the biometric-protected Keychain.
// This triggers a Touch ID prompt on the user's device.
func (p *BiometricProvider) Get(service, key string) (string, error) {
	cService := C.CString(service)
	defer C.free(unsafe.Pointer(cService))
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))

	result := C.keychain_get_biometric(cService, cKey)
	if result.status == -1 {
		return "", ErrNotFound
	}
	if int(result.status) == -34018 {
		return "", fmt.Errorf("keychain biometric get: %w", ErrEntitlement)
	}
	if result.status != 0 {
		return "", fmt.Errorf("keychain biometric get: OSStatus %d (%s)", result.status, osStatusDescription(int(result.status)))
	}

	// Copy into Go []byte first so we can zero it after extracting the string.
	data := C.GoBytes(unsafe.Pointer(result.data), result.data_len)
	C.secure_free(result.data, result.data_len) // zero C heap before freeing
	pass := string(data)
	for i := range data {
		data[i] = 0 // zero the Go []byte copy
	}
	return pass, nil
}

// Set stores a secret in the login Keychain with biometric (Touch ID) access control.
// The kSecAccessControlBiometryCurrentSet flag ensures that any read of this item
// requires biometric authentication, and the item is invalidated if biometric
// enrollment changes (fingerprints added/removed).
func (p *BiometricProvider) Set(service, key, value string) error {
	cService := C.CString(service)
	defer C.free(unsafe.Pointer(cService))
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	cValue := C.CString(value)
	defer func() {
		C.memset(unsafe.Pointer(cValue), 0, C.size_t(len(value)+1)) // zero before free
		C.free(unsafe.Pointer(cValue))
	}()

	result := C.keychain_set_biometric(cService, cKey, cValue, C.int(len(value)))
	if int(result.status) == -34018 {
		return fmt.Errorf("keychain biometric set: %w", ErrEntitlement)
	}
	if result.status != 0 {
		return fmt.Errorf("keychain biometric set: OSStatus %d (%s)", result.status, osStatusDescription(int(result.status)))
	}
	return nil
}

// HasKey checks if a key exists in the biometric-protected Keychain WITHOUT
// triggering a Touch ID prompt. Queries for item attributes only, not data.
func (p *BiometricProvider) HasKey(service, key string) bool {
	cService := C.CString(service)
	defer C.free(unsafe.Pointer(cService))
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))

	return C.keychain_has_biometric(cService, cKey) == 1
}

// Delete removes a secret from the biometric-protected Keychain.
func (p *BiometricProvider) Delete(service, key string) error {
	cService := C.CString(service)
	defer C.free(unsafe.Pointer(cService))
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))

	status := C.keychain_delete_biometric(cService, cKey)
	if status == -1 {
		return ErrNotFound
	}
	if int(status) == -34018 {
		return fmt.Errorf("keychain biometric delete: %w", ErrEntitlement)
	}
	if status != 0 {
		return fmt.Errorf("keychain biometric delete: OSStatus %d (%s)", status, osStatusDescription(int(status)))
	}
	return nil
}

// osStatusDescription returns a human-readable description for common Security
// framework OSStatus error codes. This helps diagnose Keychain issues without
// requiring the developer to look up Apple documentation.
func osStatusDescription(code int) string {
	switch code {
	case -34018:
		return "errSecMissingEntitlement: binary needs Apple Developer signing"
	case -25308:
		return "errSecInteractionNotAllowed: cannot present Touch ID UI"
	case -128:
		return "errSecUserCanceled: user cancelled biometric prompt"
	case -25293:
		return "errSecAuthFailed: authentication failed or biometric enrollment changed"
	case -25300:
		return "errSecItemNotFound: item not found"
	case -25291:
		return "errSecInvalidOwnerEdit: device passcode may not be set"
	default:
		return "unknown"
	}
}
