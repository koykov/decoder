package decoder

import "sync/atomic"

const (
	PolicyLock = iota
	PolicyLockFree
)

var (
	// Package level mode.
	mode uint32 = PolicyLock
	// Suppress go vet warning.
	_ = SetLockPolicy
)

// Set new lock policy.
//
// PolicyLockFree allows to skip mutex lock/unlock calls no one decode ruleset added and/or updated.
// For safety purposes, default lock policy is PolicyLock.
// Recommended flow:
// * Register all your rulesets under default PolicyLock.
// * Set lock-free policy by call decoder.SetLockPolicy(decoder.PolicyLockFree)
// * Use decoders in lock-free mode.
// * Set lock policy by call decoder.SetLockPolicy(decoder.PolicyLock) before add/update rulesets.
// * Make all modifications you need.
// * Set again lock-free policy by call decoder.SetLockPolicy(decoder.PolicyLockFree)
// * ...
// Caution! It's your responsibility to set proper policy. For example, if you wouldn't set PolicyLock before add/update
// some ruleset, you may catch "concurrent map read and write" panic.
// If you don't sure, please just ignore policies and work under default PolicyLock.
func SetLockPolicy(new uint32) error {
	if new != PolicyLock && new != PolicyLockFree {
		return ErrUnknownPolicy
	}
	atomic.StoreUint32(&mode, new)
	return nil
}

// Get current policy.
func GetLockPolicy() uint32 {
	return atomic.LoadUint32(&mode)
}
