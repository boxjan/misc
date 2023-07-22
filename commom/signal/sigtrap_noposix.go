//go:build windows || plan9 || nacl || js
// +build windows plan9 nacl js

package signal

func trapSignalsPosix() {}
