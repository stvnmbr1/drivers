package tasmota

import (
	"github.com/reef-pi/hal"
	"testing"
)

func TestTasmota(t *testing.T) {
	p := newTasmota("127.0.0.1", hal.Metadata{})
	nop := NewNop()
	nop.Buffer([]byte(`{}`))
	p.SetFactory(nop.Factory)
	if err := p.On(); err != nil {
		t.Error(err)
	}
	nop.Buffer([]byte(`{}`))
	if err := p.Off(); err != nil {
		t.Error(err)
	}

	f := tasmotaFactory()

	params := map[string]interface{}{
		"Address": "http://192.168.0.60",
	}

	d, err := f.NewDriver(params, nil)

	if err != nil {
		t.Error(err)
	}
	if d.Metadata().Name == "" {
		t.Error("HAL metadata should not have empty name")
	}

	d1 := d.(hal.DigitalOutputDriver)

	if len(d1.DigitalOutputPins()) != 1 {
		t.Error("Expected exactly one output pin")
	}
	pin, err := d1.DigitalOutputPin(0)
	if err != nil {
		t.Error(err)
	}
	if pin.LastState() != false {
		t.Error("Expected initial state to be false")
	}
}
