package errors

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppendError(t *testing.T) {
	t.Run("nil base error", func(t *testing.T) {
		err1 := errors.New("first error")
		result := AppendError(nil, err1)
		require.Equal(t, err1, result)
	})

	t.Run("nil new error", func(t *testing.T) {
		err1 := errors.New("first error")
		result := AppendError(err1, nil)
		require.Equal(t, err1, result)
	})

	t.Run("both nil", func(t *testing.T) {
		result := AppendError(nil, nil)
		require.Nil(t, result)
	})

	t.Run("both non-nil", func(t *testing.T) {
		err1 := errors.New("first error")
		err2 := errors.New("second error")
		result := AppendError(err1, err2)
		
		require.NotNil(t, result)
		errorStr := result.Error()
		require.True(t, strings.Contains(errorStr, "first error"))
		require.True(t, strings.Contains(errorStr, "second error"))
	})

	t.Run("multiple appends", func(t *testing.T) {
		var err error
		err = AppendError(err, errors.New("error 1"))
		err = AppendError(err, errors.New("error 2"))
		err = AppendError(err, errors.New("error 3"))
		
		require.NotNil(t, err)
		errorStr := err.Error()
		require.True(t, strings.Contains(errorStr, "error 1"))
		require.True(t, strings.Contains(errorStr, "error 2"))
		require.True(t, strings.Contains(errorStr, "error 3"))
	})
}

func TestAppendErrorf(t *testing.T) {
	t.Run("nil base error", func(t *testing.T) {
		result := AppendErrorf(nil, "error %d", 42)
		require.Equal(t, "error 42", result.Error())
	})

	t.Run("non-nil base error", func(t *testing.T) {
		baseErr := errors.New("base error")
		result := AppendErrorf(baseErr, "formatted error %s", "test")
		
		require.NotNil(t, result)
		errorStr := result.Error()
		require.True(t, strings.Contains(errorStr, "base error"))
		require.True(t, strings.Contains(errorStr, "formatted error test"))
	})
}