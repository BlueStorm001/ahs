package transfer

import (
	"ahs_server/module"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

var (
	config       *module.Config
	configDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	configTime   = time.Now()
)

func localConfiguration(force bool) (*module.Config, bool) {
	file := configDir + "/config.json"
	fileInfo, err := os.Stat(file)
	if err != nil {
		return nil, false
	}
	var currentTime = fileInfo.ModTime()
	if !force {
		if currentTime.Sub(configTime).Seconds() == 0 {
			return config, false
		}
	}
	configTime = currentTime
	if b, e := ioutil.ReadFile(file); e == nil {
		var c *module.Config
		if e = json.Unmarshal(b, &c); e == nil {
			return c, true
		}
	}
	return nil, false
}

func (w *web) routing(request *module.Request) bool {
	for _, path := range w.host.Routing {
		if path == request.Header.Path {
			return true
		}
	}
	return false
}

func (w *web) lbFor(request *module.Request) (err error) {
	for _, lb := range w.host.Lbs {
		if service, ok := apis[lb.Listen]; ok {
			var data []byte
			if data, err = json.Marshal(request); err == nil {
				err = service.Api.Send(data)
				if err == nil {
					lb.Count++
					return err
				}
			}
		}
	}
	return errors.New("data was not delivered")
}
