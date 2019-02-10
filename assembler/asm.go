package assembler

import (
	"bytes"
	"encoding/binary"
	"strings"

	"github.com/qiniu/arch/von"
)

type Builder struct {
	data        *bytes.Buffer
	undefLabels map[string]*labelRef
	defLabels   map[string]int64
}

type labelRef struct {
	data []int
}

func New(buf *bytes.Buffer) *Builder {
	if buf == nil {
		buf = new(bytes.Buffer)
	}
	return &Builder{
		data:        buf,
		undefLabels: make(map[string]*labelRef),
		defLabels:   make(map[string]int64),
	}
}

func (p *Builder) Bytes() []byte {
	n := len(p.undefLabels)
	if n > 0 {
		labels := make([]string, 0, n)
		for label, _ := range p.undefLabels {
			labels = append(labels, label)
		}
		panic("Undefined labels: " + strings.Join(labels, ", "))
	}
	return p.data.Bytes()
}

func (p *Builder) Add() *Builder {
	p.writeU16(von.ADD)
	return p
}

func (p *Builder) Sub() *Builder {
	p.writeU16(von.SUB)
	return p
}

func (p *Builder) Mul() *Builder {
	p.writeU16(von.MUL)
	return p
}

func (p *Builder) Div() *Builder {
	p.writeU16(von.DIV)
	return p
}

func (p *Builder) Mod() *Builder {
	p.writeU16(von.MOD)
	return p
}

func (p *Builder) Neg() *Builder {
	p.writeU16(von.NEG)
	return p
}

func (p *Builder) LessThanInt() *Builder {
	p.writeU16(von.LTI)
	return p
}

func (p *Builder) LessThanString() *Builder {
	p.writeU16(von.LTS)
	return p
}

func (p *Builder) EqualInt() *Builder {
	p.writeU16(von.EQI)
	return p
}

func (p *Builder) EqualString() *Builder {
	p.writeU16(von.EQS)
	return p
}

func (p *Builder) Not() *Builder {
	p.writeU16(von.NOT)
	return p
}

func (p *Builder) Concat() *Builder {
	p.writeU16(von.CONCAT)
	return p
}

func (p *Builder) Index() *Builder {
	p.writeU16(von.INDEX)
	return p
}

func (p *Builder) String() *Builder {
	p.writeU16(von.STRING)
	return p
}

func (p *Builder) Alloc() *Builder {
	p.writeU16(von.ALLOC)
	return p
}

func (p *Builder) Read(port uint16) *Builder {
	p.writeU16(von.READ)
	p.writeU16(port)
	return p
}

func (p *Builder) Write(port uint16) *Builder {
	p.writeU16(von.WRITE)
	p.writeU16(port)
	return p
}

func (p *Builder) PushInt(v int64) *Builder {
	p.writeU16(von.PUSHI)
	p.writeI64(v)
	return p
}

func (p *Builder) PushString(v string) *Builder {
	p.writeU16(von.PUSHS)
	p.writeU16(uint16(len(v)))
	p.data.WriteString(v)
	return p
}

func (p *Builder) PushArg(index int16) *Builder {
	p.writeU16(von.PUSHA)
	p.writeU16(uint16(index))
	return p
}

func (p *Builder) SetArg(index int16) *Builder {
	p.writeU16(von.SETA)
	p.writeU16(uint16(index))
	return p
}

func (p *Builder) Ret(narg int) *Builder {
	p.writeU16(von.RET)
	p.writeU16(uint16(narg))
	return p
}

func (p *Builder) Jmp(name string) *Builder {
	return p.goLabel(von.JMP, name)
}

func (p *Builder) JZ(name string) *Builder {
	return p.goLabel(von.JZ, name)
}

func (p *Builder) Call(name string) *Builder {
	return p.goLabel(von.CALL, name)
}

func (p *Builder) Label(name string) *Builder {
	if _, ok := p.defLabels[name]; ok {
		panic("Redefine label: " + name)
	}
	pc := p.data.Len()
	if lref, ok := p.undefLabels[name]; ok {
		b := p.data.Bytes()
		for _, off := range lref.data {
			binary.LittleEndian.PutUint64(b[off:], uint64(pc-(off-2)))
		}
		delete(p.undefLabels, name)
	}
	p.defLabels[name] = int64(pc)
	return p
}

func (p *Builder) goLabel(op uint16, name string) *Builder {
	p.writeU16(op)
	if pc, ok := p.defLabels[name]; ok {
		base := int64(p.data.Len() - 2)
		p.writeI64(pc - base)
	} else {
		lref := p.reqUndefLabel(name)
		lref.data = append(lref.data, p.data.Len())
		p.writeI64(0)
	}
	return p
}

func (p *Builder) reqUndefLabel(name string) *labelRef {
	if lref, ok := p.undefLabels[name]; ok {
		return lref
	}
	lref := new(labelRef)
	p.undefLabels[name] = lref
	return lref
}

func (p *Builder) Halt() *Builder {
	p.writeU16(von.HALT)
	return p
}

func (p *Builder) writeU16(v uint16) {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], v)
	p.data.Write(buf[:])
}

func (p *Builder) writeI64(v int64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(v))
	p.data.Write(buf[:])
}
