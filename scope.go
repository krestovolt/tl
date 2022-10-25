package tl

import "strings"

type ScopeTypeEnum uint8

const (
	ScopeEmpty ScopeTypeEnum = iota
	ScopeHandshake
	ScopeTransport
	ScopeSync
	ScopeCoreTypes
)

func (st ScopeTypeEnum) String() string {
	switch st {
	case ScopeHandshake:
		return "handshake"
	case ScopeTransport:
		return "transport"
	case ScopeSync:
		return "sync"
	case ScopeCoreTypes:
		return "core_types"
	case ScopeEmpty:
		fallthrough
	default:
		return ""
	}
}

// ParseScope parses a given line into a ScopeTypeEnum enums
func ParseScope(line string) (ScopeTypeEnum, bool) {
	if !strings.HasPrefix(line, "//") {
		// Might have type definition, should ignore this line
		return ScopeEmpty, false
	}

	if strings.HasSuffix(line, tokScopeHandshake) {
		return ScopeHandshake, true
	}
	if strings.HasSuffix(line, tokScopeTransport) {
		return ScopeTransport, true
	}
	if strings.HasSuffix(line, tokScopeSync) {
		return ScopeSync, true
	}
	if strings.HasSuffix(line, tokScopeCoreTypes) {
		return ScopeCoreTypes, true
	}
	if strings.HasSuffix(line, tokScopeEnd) {
		// End of scope
		return ScopeEmpty, true
	}

	return ScopeEmpty, false
}
