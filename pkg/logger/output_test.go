package logger_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"finance-app-go-demo/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOutput_WithoutFileReturnsWriterAsIs(t *testing.T) {
	var stdout bytes.Buffer

	out, closeOut := logger.NewOutput(&stdout, logger.FileConfig{})

	assert.Same(t, &stdout, out)
	assert.NoError(t, closeOut())
}

func TestNewOutput_WritesToBothStdoutAndFile(t *testing.T) {
	var stdout bytes.Buffer
	path := filepath.Join(t.TempDir(), "logs", "app.log")

	out, closeOut := logger.NewOutput(&stdout, logger.FileConfig{Path: path, MaxSizeMB: 1, MaxBackups: 1, MaxAgeDays: 1})

	_, err := out.Write([]byte("{\"msg\":\"hello\"}\n"))
	require.NoError(t, err)
	require.NoError(t, closeOut())

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "{\"msg\":\"hello\"}\n", string(data))
	assert.Equal(t, "{\"msg\":\"hello\"}\n", stdout.String())
}
