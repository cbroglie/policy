package rqdata

import (
	"strconv"
	"strings"

	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

type requestFunc func(state request.Request) string

// Mapping define the mapping between 'name' of data and the way to extract that data from the Request
// it also defines what will be the empty value returned if the data behind the name is empty.
// it is pretty static, and you should need to instantiate only once
type Mapping struct {
	replacements map[string]requestFunc
	emptyValue   string
}

// Extractor implements a Value(name) (value, valid) function
// which allow to extract data from an existing DNS Request(or state)
type Extractor struct {
	state     request.Request
	requester *Mapping
}

//NewExtractor return a new Extractor based on the Mapping and the Request provided
func NewExtractor(r request.Request, m *Mapping) *Extractor {
	return &Extractor{r, m}
}

// NewMapping build the mapping name -> func to extract data from the Request
func NewMapping(emptyValue string) *Mapping {
	replacements := map[string]requestFunc{
		"type": func(state request.Request) string {
			return state.Type()
		},
		"name": func(state request.Request) string {
			return state.Name()
		},
		"class": func(state request.Request) string {
			return state.Class()
		},
		"proto": func(state request.Request) string {
			return state.Proto()
		},
		"size": func(state request.Request) string {
			return strconv.Itoa(state.Len())
		},
		"client_ip": func(state request.Request) string {
			return addrToRFC3986(state.IP())
		},
		"port": func(state request.Request) string {
			return addrToRFC3986(state.Port())
		},
		">id": func(state request.Request) string {
			return strconv.Itoa(int(state.Req.Id))
		},
		">opcode": func(state request.Request) string {
			return strconv.Itoa(int(state.Req.Opcode))
		},
		">do": func(state request.Request) string {
			return boolToString(state.Do())
		},
		">bufsize": func(state request.Request) string {
			return strconv.Itoa(state.Size())
		},
		"server_ip": func(state request.Request) string {
			return addrToRFC3986(state.LocalIP())
		},
		"server_port": func(state request.Request) string {
			return addrToRFC3986(state.LocalPort())
		},
	}
	return &Mapping{replacements, emptyValue}
}

func (m *Mapping) ValidField(name string) (bool) {
	_, ok := m.replacements[name]
	return ok
}

// Value extract the data that is mapped to this name and return the corresponding value as a string
// if that value is empty then the defaultValue is returned
// Second parameter is a boolean that inform if the name itself is supported in the mapping
func (rd *Extractor) Value(name string) (string, bool) {
	f, ok := rd.requester.replacements[name]
	if ok {
		v := f(rd.state)
		if v != "" {
			return v, true
		}
		return rd.requester.emptyValue, true
	}
	return "", false
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// flagsToString checks all header flags and returns those
// that are set as a string separated with commas
func flagsToString(h dns.MsgHdr) string {
	flags := make([]string, 7)
	i := 0

	if h.Response {
		flags[i] = "qr"
		i++
	}

	if h.Authoritative {
		flags[i] = "aa"
		i++
	}
	if h.Truncated {
		flags[i] = "tc"
		i++
	}
	if h.RecursionDesired {
		flags[i] = "rd"
		i++
	}
	if h.RecursionAvailable {
		flags[i] = "ra"
		i++
	}
	if h.Zero {
		flags[i] = "z"
		i++
	}
	if h.AuthenticatedData {
		flags[i] = "ad"
		i++
	}
	if h.CheckingDisabled {
		flags[i] = "cd"
		i++
	}
	return strings.Join(flags[:i], ",")
}

// addrToRFC3986 will add brackets to the address if it is an IPv6 address.
func addrToRFC3986(addr string) string {
	if strings.Contains(addr, ":") {
		return "[" + addr + "]"
	}
	return addr
}
