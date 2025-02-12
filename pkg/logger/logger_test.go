package logger

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {

	filePath := "test.log"
	if _, err := os.Stat(filePath); err == nil {
		// Файл существует, удаляем его
		err := os.Remove(filePath)
		assert.NoError(t, err)
	}

	InitLogger("debug", "test.log")
	assert.NotNil(t, instance)

	lg := Log()
	assert.NotNil(t, lg)

	//Log().Info("Test")

}
