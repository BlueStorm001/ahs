package module

type Config struct {
	Host []Host
}

type Host struct {
	Status        int                 `json:"status"`
	Domain        string              `json:"domain"`
	Port          int                 `json:"port"`
	Timeout       int                 `json:"timeout"`
	ContentType   string              `json:"contentType,omitempty"`
	ErrorResponse string              `json:"errorResponse,omitempty"`
	Routing       []string            `json:"routing,omitempty"`
	UseHeader     []string            `json:"useHeader"`
	HeaderMap     map[string]struct{} `json:"-"`
	LbsModeRemark string              `json:"lbsModeRemark,omitempty"`
	LbsMode       int                 `json:"lbsMode,omitempty"`
	Lbs           []Lbs               `json:"lbs"`
}

type Lbs struct {
	Listen    string     `json:"listen"`
	Key       string     `json:"key,omitempty"`
	Port      int        `json:"port"`
	Heartbeat *Heartbeat `json:"heartbeat"`
	IdleTime  *IdleTime  `json:"idleTime"` //空闲时间
	Count     int
}

type Heartbeat struct {
	Content  string `json:"content,omitempty"`
	Interval int    `json:"interval"`
}

type IdleTime struct {
	MaxTimeout int `json:"maxTimeout"`
}
