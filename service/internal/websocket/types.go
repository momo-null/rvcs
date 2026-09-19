package websocket

// MessageType 消息类型
type MessageType string

const (
	MessageTypeAuth         MessageType = "auth"
	MessageTypeControl      MessageType = "control"
	MessageTypeResponse     MessageType = "response"
	MessageTypeEvent        MessageType = "event"
	MessageTypePing         MessageType = "ping"
	MessageTypePong         MessageType = "pong"
	MessageTypeConfigUpdate MessageType = "config_update"
)

// Message 消息结构
type Message struct {
	MsgID      string                 `json:"msg_id"`
	Type       MessageType            `json:"type"`
	Command    string                 `json:"command,omitempty"`
	Status     string                 `json:"status,omitempty"`
	State      string                 `json:"state,omitempty"` // Android端摄像头状态
	Timestamp  int64                  `json:"timestamp"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Error      *ErrorInfo             `json:"error,omitempty"`
	Signature  string                 `json:"signature,omitempty"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	DeviceID   string `json:"device_id"`
	Status     string `json:"status"`
	IPAddress  string `json:"ip_address"`
	UserAgent  string `json:"user_agent"`
	ConnectedAt int64 `json:"connected_at"`
}
