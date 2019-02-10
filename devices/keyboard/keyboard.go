package keyboard

import (
	"github.com/qiniu/arch/drivers/keyboard"
	"github.com/qiniu/arch/utils/deque"
)

type Device struct {
	data *deque.Deque
}

type keyEvent struct {
	evType byte
	evData byte
}

func New() *Device {
	return &Device{
		data: deque.New(),
	}
}

func (p *Device) Read(b []byte) (int, error) {
	n := len(b) / 2
	for i := 0; i < n; i++ {
		if v, ok := p.data.PopFront(); ok {
			o := v.(keyEvent)
			b[i<<1] = o.evType
			b[(i<<1)+1] = o.evData
		} else {
			return i * 2, nil
		}
	}
	return n * 2, nil
}

func (p *Device) Write(b []byte) (int, error) {
	panic("can't write to keyboard")
}

func (p *Device) KeyDown(key keyboard.Key) *Device {
	p.data.PushBack(keyEvent{
		evType: keyboard.KEYDOWN,
		evData: byte(key),
	})
	return p
}

func (p *Device) KeyUp(key keyboard.Key) *Device {
	p.data.PushBack(keyEvent{
		evType: keyboard.KEYUP,
		evData: byte(key),
	})
	return p
}

func (p *Device) KeyPress(key keyboard.Key) *Device {
	return p.KeyDown(key).KeyUp(key)
}
