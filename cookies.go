package gocaptcha

import (
	"fmt"
	"strings"
)

type Cookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Cookies []Cookie

func (c Cookies) Add(cookie Cookie) Cookies {
	return append(c, cookie)
}

// String returns the cookies as a string, e.g. "name1=value1; name2=value2"
func (c Cookies) String() string {
	r := strings.Builder{}
	for i, cookie := range c {
		if i > 0 {
			r.WriteString("; ")
		}
		r.WriteString(fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}
	return r.String()
}

// StringAlternative returns the cookies as a string, e.g. "name1:value1; name2:value2"
func (c Cookies) StringAlternative() string {
	r := strings.Builder{}
	for i, cookie := range c {
		if i > 0 {
			r.WriteString("; ")
		}
		r.WriteString(fmt.Sprintf("%s:%s", cookie.Name, cookie.Value))
	}
	return r.String()
}

func (c Cookies) Count() int {
	return len(c)
}

func (c Cookies) Has(name string) bool {
	for _, cookie := range c {
		if cookie.Name == name {
			return true
		}
	}
	return false
}

func (c Cookies) Get(name string) Cookie {
	for _, cookie := range c {
		if cookie.Name == name {
			return cookie
		}
	}
	return Cookie{}
}
