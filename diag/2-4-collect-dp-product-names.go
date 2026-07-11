//go:build windows

package diag

import (
	"golang.org/x/sys/windows/registry"
)

func isProductPasswordManager() bool {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, DPProductsPM, registry.READ|registry.WOW64_32KEY)
	if err != nil {
		return false
	}
	defer key.Close()
	return true
}

func getProductDPPath() string {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, DPKey, registry.READ|registry.WOW64_64KEY)
	if err != nil {
		return ""
	}
	defer key.Close()
	val, _, err := key.GetStringValue("DigitalPersonaPath")
	if err != nil {
		return ""
	}
	return val
}

func getOTSTemplatePath() string {
	key, err := registry.OpenKey(registry.CURRENT_USER, DPGroupPolicyKey, registry.READ)
	if err != nil {
		return ""
	}
	defer key.Close()
	val, _, err := key.GetStringValue("TemplatePath")
	if err != nil {
		return ""
	}
	return val
}
