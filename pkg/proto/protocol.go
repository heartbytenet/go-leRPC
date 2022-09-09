package proto

type ExecuteCommand struct {
	Token     string                 `json:"tk"`
	ID        string                 `json:"id"`
	Namespace string                 `json:"ns"`
	Method    string                 `json:"mt"`
	Params    map[string]interface{} `json:"pm"`
}

type ExecuteResult struct {
	ID      string                 `json:"id,omitempty"`
	Success bool                   `json:"ok"`
	Payload map[string]interface{} `json:"pl"`
	Error   string                 `json:"er,omitempty"`
}

func (cmd *ExecuteCommand) SetNamespace(value string) *ExecuteCommand {
	cmd.Namespace = value
	return cmd
}

func (cmd *ExecuteCommand) SetMethod(value string) *ExecuteCommand {
	cmd.Method = value
	return cmd
}

func (cmd *ExecuteCommand) SetToken(value string) *ExecuteCommand {
	cmd.Token = value
	return cmd
}

func (cmd *ExecuteCommand) SetParams(value map[string]interface{}) *ExecuteCommand {
	cmd.Params = value
	return cmd
}

func (cmd *ExecuteCommand) SetParam(key string, value interface{}) *ExecuteCommand {
	if cmd.Params == nil {
		cmd.Params = map[string]interface{}{}
	}
	cmd.Params[key] = value
	return cmd
}

func (res *ExecuteResult) ToPayload(value map[string]interface{}) *ExecuteResult {
	res.Success = true
	res.Payload = value
	return res
}

func (res *ExecuteResult) ToError(value string) *ExecuteResult {
	res.Success = false
	res.Payload = nil
	res.Error = value
	return res
}
