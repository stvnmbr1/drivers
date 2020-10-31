package tasmota

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/reef-pi/hal"
)

const addressParam = "Address"

type Tasmota struct {
	state   bool
	command *cmd
	meta    hal.Metadata
}

func newTasmota(addr string, meta hal.Metadata) *Tasmota {
	return &Tasmota{
		meta: meta,
		command: &cmd{
			addr: addr,
			cf: func(proto, addr string, t time.Duration) (Conn, error) {
				return net.DialTimeout(proto, addr, t)
			},
		},
	}
}

func (p *Tasmota) SetFactory(cf ConnectionFactory) {
	p.command.cf = cf
}
func (p *Tasmota) On() error {
	cmd := new(CmdRelayState)
	cmd.System.RelayState.State = 1
	if _, err := p.command.Execute(cmd, false); err != nil {
		return err
	}
	p.state = true
	return nil
}

func (p *Tasmota) Off() error {
	cmd := new(CmdRelayState)
	cmd.System.RelayState.State = 0
	if _, err := p.command.Execute(cmd, false); err != nil {
		return err
	}
	p.state = false
	return nil
}

func (p *Tasmota) Info() (*Sysinfo, error) {
	buf, err := p.command.Execute(new(Plug), true)
	if err != nil {
		return nil, err
	}
	var d Plug
	if err := json.Unmarshal(buf, &d); err != nil {
		return nil, err
	}
	return &d.System.Sysinfo, nil
}

func (p *Tasmota) Metadata() hal.Metadata {
	return p.meta
}

func (p *Tasmota) Name() string {
	return p.meta.Name
}

func (p *Tasmota) Number() int {
	return 0
}
func (p *Tasmota) DigitalOutputPins() []hal.DigitalOutputPin {
	return []hal.DigitalOutputPin{p}
}

func (p *Tasmota) DigitalOutputPin(i int) (hal.DigitalOutputPin, error) {
	if i != 0 {
		return nil, fmt.Errorf("invalid pin: %d", i)
	}
	return p, nil
}

func (p *Tasmota) Write(state bool) error {
	if state {
		return p.On()
	}
	return p.Off()
}

func (p *Tasmota) LastState() bool {
	return p.state
}

func (p *Tasmota) Close() error {
	return nil
}
func (p *Tasmota) Pins(cap hal.Capability) ([]hal.Pin, error) {
	switch cap {
	case hal.DigitalOutput:
		return []hal.Pin{p}, nil
	default:
		return nil, fmt.Errorf("unsupported capability:%s", cap.String())
	}
}

type tasmotaFactory struct {
	meta       hal.Metadata
	parameters []hal.ConfigParameter
}

var factorytasmota *tasmotaFactory
var tasmotaonce sync.Once

// HS103Factory returns a singleton HS103 Driver factory
func TASMOTAFactory() hal.DriverFactory {

	tasmotaonce.Do(func() {
		factorytasmota = &tasmotaFactory{
			meta: hal.Metadata{
				Name:        "Tasmota",
				Description: "Tasmota",
				Capabilities: []hal.Capability{
					hal.DigitalOutput,
				},
			},
			parameters: []hal.ConfigParameter{
				{
					Name:    addressParam,
					Type:    hal.String,
					Order:   0,
					Default: "192.168.1.11:9999",
				},
			},
		}
	})

	return factorytasmota
}

func (f *tasmotaFactory) Metadata() hal.Metadata {
	return f.meta
}

func (f *tasmotaFactory) GetParameters() []hal.ConfigParameter {
	return f.parameters
}

func (f *tasmotaFactory) ValidateParameters(parameters map[string]interface{}) (bool, map[string][]string) {

	var failures = make(map[string][]string)

	if v, ok := parameters[addressParam]; ok {
		_, ok := v.(string)
		if !ok {
			failure := fmt.Sprint(addressParam, " is not a string. ", v, " was received.")
			failures[addressParam] = append(failures[addressParam], failure)
		}
	} else {
		failure := fmt.Sprint(addressParam, " is a required parameter, but was not received.")
		failures[addressParam] = append(failures[addressParam], failure)
	}

	return len(failures) == 0, failures
}

func (f *tasmotaFactory) NewDriver(parameters map[string]interface{}, _ interface{}) (hal.Driver, error) {
	if valid, failures := f.ValidateParameters(parameters); !valid {
		return nil, errors.New(hal.ToErrorString(failures))
	}

	addr := parameters[addressParam].(string)

	return newTasmota(addr, f.meta), nil
}
