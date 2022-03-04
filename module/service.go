package module

type Api interface {
	Run(name, port string) error
	Handler(func(conn HttpConn, data []byte))
	Shutdown() error
	Send(data []byte) error
	Heartbeat(data string, interval int)
	CheckIdleTime(interval int)
}

type Service struct {
	Api  Api
	Port string
}
