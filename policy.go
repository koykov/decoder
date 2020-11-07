package decoder

import (
	"github.com/koykov/policy"
)

var (
	// Suppress go vet warnings.
	_, _ = SetLockPolicy, GetLockPolicy
)

// Set new lock policy.
//
// policy.LockFree allows to skip mutex lock/unlock calls when no one decode ruleset added and/or updated.
// For safety purposes, default lock policy is policy.Locked.
// Recommended flow:
// * Register all your rulesets under default policy.Locked.
// * Set lock-free policy by call decoder.SetLockPolicy(policy.LockFree)
// * Use decoders in lock-free mode.
// * Set lock policy by call decoder.SetLockPolicy(policy.Locked) before add/update rulesets.
// * Make all modifications you need.
// * Set again lock-free policy by call decoder.SetLockPolicy(policy.LockFree)
// * ...
// Caution! It's your responsibility to set proper policy. For example, if you wouldn't set policy.Locked before
// add/update some ruleset, you may catch "concurrent map read and write" panic.
// If you don't sure, please just ignore policies and work under default policy.Locked.
func SetLockPolicy(new policy.Policy) {
	lock.SetPolicy(new)
}

// Get current policy.
func GetLockPolicy() policy.Policy {
	return lock.GetPolicy()
}
