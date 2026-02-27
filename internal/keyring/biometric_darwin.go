//go:build darwin && cgo

package keyring

/*
#cgo LDFLAGS: -framework CoreFoundation -framework Security

#include <CoreFoundation/CoreFoundation.h>
#include <Security/Security.h>
#include <stdlib.h>
#include <string.h>

// KeychainResult is returned by C helper functions.
typedef struct {
	int    status;    // 0 = success, -1 = not found, >0 = OSStatus error
	char  *data;      // returned data (caller must free)
	int    data_len;
} KeychainResult;

static CFStringRef _toCFString(const char *s) {
	return CFStringCreateWithCString(kCFAllocatorDefault, s, kCFStringEncodingUTF8);
}

// keychain_biometric_available checks if biometric access control can be created.
static int keychain_biometric_available(void) {
	CFErrorRef error = NULL;
	SecAccessControlRef access = SecAccessControlCreateWithFlags(
		kCFAllocatorDefault,
		kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
		kSecAccessControlBiometryAny,
		&error
	);
	if (error != NULL) {
		CFRelease(error);
		return 0;
	}
	if (access != NULL) {
		CFRelease(access);
	}
	return 1;
}

// keychain_set_biometric stores a value with biometric (Touch ID) access control.
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
	CFDictionarySetValue(delQuery, kSecClass,       kSecClassGenericPassword);
	CFDictionarySetValue(delQuery, kSecAttrService,  cfService);
	CFDictionarySetValue(delQuery, kSecAttrAccount,  cfAccount);
	SecItemDelete(delQuery);
	CFRelease(delQuery);

	// Create biometric access control.
	CFErrorRef acError = NULL;
	SecAccessControlRef access = SecAccessControlCreateWithFlags(
		kCFAllocatorDefault,
		kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
		kSecAccessControlBiometryAny,
		&acError);
	if (acError != NULL) {
		result.status = (int)CFErrorGetCode(acError);
		CFRelease(acError);
		CFRelease(cfService);
		CFRelease(cfAccount);
		CFRelease(cfValue);
		return result;
	}

	// Add item with biometric protection.
	CFMutableDictionaryRef query = CFDictionaryCreateMutable(
		kCFAllocatorDefault, 0,
		&kCFTypeDictionaryKeyCallBacks,
		&kCFTypeDictionaryValueCallBacks);
	CFDictionarySetValue(query, kSecClass,            kSecClassGenericPassword);
	CFDictionarySetValue(query, kSecAttrService,       cfService);
	CFDictionarySetValue(query, kSecAttrAccount,       cfAccount);
	CFDictionarySetValue(query, kSecValueData,         cfValue);
	CFDictionarySetValue(query, kSecAttrAccessControl, access);

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
static KeychainResult keychain_get_biometric(const char *service, const char *account) {
	KeychainResult result = {0, NULL, 0};

	CFStringRef cfService = _toCFString(service);
	CFStringRef cfAccount = _toCFString(account);

	CFMutableDictionaryRef query = CFDictionaryCreateMutable(
		kCFAllocatorDefault, 0,
		&kCFTypeDictionaryKeyCallBacks,
		&kCFTypeDictionaryValueCallBacks);
	CFDictionarySetValue(query, kSecClass,       kSecClassGenericPassword);
	CFDictionarySetValue(query, kSecAttrService,  cfService);
	CFDictionarySetValue(query, kSecAttrAccount,  cfAccount);
	CFDictionarySetValue(query, kSecMatchLimit,   kSecMatchLimitOne);
	CFDictionarySetValue(query, kSecReturnData,   kCFBooleanTrue);

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
static int keychain_has_biometric(const char *service, const char *account) {
	CFStringRef cfService = _toCFString(service);
	CFStringRef cfAccount = _toCFString(account);

	CFMutableDictionaryRef query = CFDictionaryCreateMutable(
		kCFAllocatorDefault, 0,
		&kCFTypeDictionaryKeyCallBacks,
		&kCFTypeDictionaryValueCallBacks);
	CFDictionarySetValue(query, kSecClass,            kSecClassGenericPassword);
	CFDictionarySetValue(query, kSecAttrService,       cfService);
	CFDictionarySetValue(query, kSecAttrAccount,       cfAccount);
	CFDictionarySetValue(query, kSecMatchLimit,        kSecMatchLimitOne);
	CFDictionarySetValue(query, kSecReturnAttributes,  kCFBooleanTrue);

	CFTypeRef item = NULL;
	OSStatus status = SecItemCopyMatching(query, &item);

	if (item != NULL) CFRelease(item);
	CFRelease(query);
	CFRelease(cfService);
	CFRelease(cfAccount);

	return (status == errSecSuccess) ? 1 : 0;
}

// keychain_delete_biometric deletes the item.
static int keychain_delete_biometric(const char *service, const char *account) {
	CFStringRef cfService = _toCFString(service);
	CFStringRef cfAccount = _toCFString(account);

	CFMutableDictionaryRef query = CFDictionaryCreateMutable(
		kCFAllocatorDefault, 0,
		&kCFTypeDictionaryKeyCallBacks,
		&kCFTypeDictionaryValueCallBacks);
	CFDictionarySetValue(query, kSecClass,       kSecClassGenericPassword);
	CFDictionarySetValue(query, kSecAttrService,  cfService);
	CFDictionarySetValue(query, kSecAttrAccount,  cfAccount);

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

// BiometricProvider stores secrets in macOS Keychain with Touch ID (biometric)
// protection. Items stored with this provider require biometric authentication
// for retrieval, preventing other processes from reading secrets without user
// presence verification via the Secure Enclave.
type BiometricProvider struct{}

var _ Provider    = (*BiometricProvider)(nil)
var _ KeyChecker  = (*BiometricProvider)(nil)

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
	if result.status != 0 {
		return "", fmt.Errorf("keychain biometric get: OSStatus %d", result.status)
	}

	data := C.GoStringN(result.data, result.data_len)
	C.free(unsafe.Pointer(result.data))
	return data, nil
}

// Set stores a secret in the Keychain with biometric (Touch ID) access control.
// The kSecAccessControlBiometryAny flag ensures that any read of this item
// requires biometric authentication, even from the same UID.
func (p *BiometricProvider) Set(service, key, value string) error {
	cService := C.CString(service)
	defer C.free(unsafe.Pointer(cService))
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	result := C.keychain_set_biometric(cService, cKey, cValue, C.int(len(value)))
	if result.status != 0 {
		return fmt.Errorf("keychain biometric set: OSStatus %d", result.status)
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
	if status != 0 {
		return fmt.Errorf("keychain biometric delete: OSStatus %d", status)
	}
	return nil
}
