package proto

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExecuteCommand_SetNamespace(t *testing.T) {
	cmd := ExecuteCommand{}
	assert.Equal(t, cmd.Namespace, "")

	cmd.SetNamespace("test1234")
	assert.Equal(t, cmd.Namespace, "test1234")

	cmd.SetNamespace("")
	assert.Equal(t, cmd.Namespace, "")
}

func TestExecuteCommand_SetMethod(t *testing.T) {
	cmd := ExecuteCommand{}
	assert.Equal(t, cmd.Method, "")

	cmd.SetMethod("test1234")
	assert.Equal(t, cmd.Method, "test1234")

	cmd.SetMethod("")
	assert.Equal(t, cmd.Method, "")
}

func TestExecuteCommand_SetParams(t *testing.T) {
	cmd := ExecuteCommand{}
	assert.Equal(t, len(cmd.Params), 0)

	cmd.SetParams(map[string]interface{}{"hello": "world"})
	assert.Equal(t, cmd.Params, map[string]interface{}{"hello": "world"})

	cmd.SetParams(nil)
	assert.Equal(t, len(cmd.Params), 0)
}
