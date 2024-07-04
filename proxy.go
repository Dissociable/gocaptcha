package gocaptcha

import "fmt"

type Proxy struct {
	Type     string `json:"proxyType"`
	Host     string `json:"proxyAddress"`
	Port     int    `json:"proxyPort"`
	Username string `json:"proxyLogin"`
	Password string `json:"proxyPassword"`
}

func NewProxy(typ string, host string, port int, username string, password string) *Proxy {
	return &Proxy{
		Type:     typ,
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
	}
}

// Map returns the proxy as a map
func (p *Proxy) Map() map[string]any {
	return map[string]any{
		"proxyType":     p.Type,
		"proxyAddress":  p.Host,
		"proxyPort":     p.Port,
		"proxyLogin":    p.Username,
		"proxyPassword": p.Password,
	}
}

func (p *Proxy) String() string {
	return fmt.Sprintf("%s://%s:%s@%s:%d", p.Type, p.Username, p.Password, p.Host, p.Port)
}

func (p *Proxy) IsEmpty() bool {
	return p.Type == "" || p.Host == "" || p.Port == 0
}
