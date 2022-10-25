package tl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScopeTypeEnum(t *testing.T) {
	for _, tt := range []struct {
		EScope ScopeTypeEnum
		String string
		OK     bool
	}{
		{
			EScope: ScopeCoreTypes,
			String: "// " + tokScopeCoreTypes,
			OK:     true,
		},
		{
			EScope: ScopeHandshake,
			String: "// " + tokScopeHandshake,
			OK:     true,
		},
		{
			EScope: ScopeSync,
			String: "// " + tokScopeSync,
			OK:     true,
		},
		{
			EScope: ScopeTransport,
			String: "//// // / // / ////" + tokScopeTransport,
			OK:     true,
		},
		{
			EScope: ScopeEmpty,
			String: "// " + tokScopeEnd,
			OK:     true,
		},
		// negatives
		{
			EScope: ScopeEmpty,
			String: tokScopeCoreTypes,
			OK:     false,
		},
		{
			EScope: ScopeEmpty,
			String: "/",
			OK:     false,
		},
		{
			EScope: ScopeEmpty,
			String: "//",
			OK:     false,
		},
		{
			EScope: ScopeEmpty,
			String: "// " + tokScopeHandshake + " //",
			OK:     false,
		},
		{
			EScope: ScopeEmpty,
			String: "msg data:bytes list:flags.0?Vector<long> = Message // System messages",
			OK:     false,
		},
	} {
		t.Run(tt.String, func(t *testing.T) {
			t.Run("ParseScope", func(t *testing.T) {
				parsedScope, ok := ParseScope(tt.String)

				require.Equal(t, ok, tt.OK)
				require.Equal(t, parsedScope.String(), tt.EScope.String())
			})
		})
	}
}
