package internal_test

import (
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/AndreySenov/meridian/v3/internal"
)

func TestNoCopy(t *testing.T) {
	t.Run("Only a pointer to NoCopy is a sync.Locker", func(t *testing.T) {
		locker := reflect.TypeFor[sync.Locker]()
		require.True(t, reflect.TypeFor[*internal.NoCopy]().Implements(locker))
		require.False(t, reflect.TypeFor[internal.NoCopy]().Implements(locker))
	})
}
