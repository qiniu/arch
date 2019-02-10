package assembler

import (
	"testing"

	"github.com/qiniu/arch/devices/console"
	"github.com/qiniu/arch/devices/keyboard"
	"github.com/qiniu/arch/drivers"
	"github.com/qiniu/arch/von"

	con "github.com/qiniu/arch/drivers/console"
	kb "github.com/qiniu/arch/drivers/keyboard"
)

type shortMem struct {
	b []byte
}

func newShortMem(b []byte) *shortMem {
	code := make([]byte, von.PageSize)
	copy(code, b)
	return &shortMem{code}
}

func (p *shortMem) OnPageMiss(ipage int64) ([]byte, error) {
	if ipage > 0 {
		panic("out of range")
	}
	return p.b, nil
}

var (
	theKeyboard *keyboard.Device
)

func run(b *Builder) *von.CPU {
	b.Halt()
	code := b.Bytes()
	mem := von.NewMemory(newShortMem(code))
	cpu := von.NewCPU(mem)
	devKeyboard := theKeyboard
	if devKeyboard == nil {
		devKeyboard = keyboard.New()
	}
	cpu.AddDevice(drivers.KEYBOARD, devKeyboard)
	cpu.AddDevice(drivers.CONSOLE, console.New())
	cpu.Run(0)
	return cpu
}

func notEq(ret interface{}, v string) bool {
	return string(ret.([]byte)) != v
}

func TestInt(t *testing.T) {
	asm := New(nil)
	asm.PushInt(2).
		PushInt(3).
		Mul()
	ret := run(asm).Top(1)
	if ret != int64(6) {
		t.Fatal("TestInt:", ret)
	}
}

func TestString(t *testing.T) {
	asm := New(nil)
	asm.PushString("Hello, ").
		PushString("World").
		Concat()
	ret := run(asm).Top(1)
	if notEq(ret, "Hello, World") {
		t.Fatal("TestString:", ret)
	}
}

func TestJZ_true(t *testing.T) {
	asm := New(nil)
	asm.PushString("Hello").
		PushString("World").
		LessThanString().
		JZ("else").
		PushString("true").
		Halt().
		Label("else").
		PushString("false")
	ret := run(asm).Top(1)
	if notEq(ret, "true") {
		t.Fatal("TestJZ_true:", ret)
	}
}

func TestJZ_false(t *testing.T) {
	asm := New(nil)
	asm.PushInt(3).
		PushInt(2).
		LessThanInt().
		JZ("else").
		PushString("true").
		Halt().
		Label("else").
		PushString("false")
	ret := run(asm).Top(1)
	if notEq(ret, "false") {
		t.Fatal("TestJZ_false:", ret)
	}
}

func TestProc(t *testing.T) {
	asm := New(nil)
	asm.PushInt(0).
		PushInt(2).
		PushInt(3).
		Call("sub").
		Halt().
		Label("sub").
		PushArg(-2).
		PushArg(-1).
		Sub().
		SetArg(-3).
		Ret(2)
	ret := run(asm).Top(1)
	if ret != int64(-1) {
		t.Fatal("TestProc:", ret)
	}
}

func TestKeyboardAndConsole(t *testing.T) {
	theKeyboard = keyboard.New()
	theKeyboard.KeyDown(kb.KeyShift).
		KeyPress(kb.KeyH).
		KeyUp(kb.KeyShift).
		KeyPress(kb.KeyE).
		KeyPress(kb.KeyL).
		KeyPress(kb.KeyL).
		KeyPress(kb.KeyO)

	asm := New(nil)
	asm.PushString("").
		PushInt(128).
		Alloc(). // var2 = make([]byte, 128)
		PushArg(2).
		Read(drivers.KEYBOARD). // var3: nread int64
		PushInt(0).             // var4: i int64
		PushInt(0).             // var5: c byte
		PushInt(0).             // var6: t byte
		PushInt(0).             // var7: shift bool
		Label("loop").
		PushArg(4).
		PushArg(3).
		LessThanInt(). // i < nread?
		JZ("done").
		PushArg(2).
		PushArg(4).
		PushInt(1).
		Add().
		Index().
		SetArg(5). // c = var2[i+1]
		PushArg(2).
		PushArg(4).
		Index().
		SetArg(6). // t = var2[i]
		PushArg(5).
		PushInt(int64(kb.KeyShift)).
		EqualInt(). // c == kb.KeyShift?
		JZ("normalkey").
		PushArg(6).
		PushInt(kb.KEYDOWN).
		EqualInt().
		SetArg(7). // shift = (t == kb.KEYDOWN)
		Jmp("continue").
		Label("normalkey").
		PushArg(6).
		PushInt(kb.KEYDOWN).
		EqualInt(). // t == kb.KEYDOWN?
		JZ("continue").
		PushArg(1).
		PushArg(7).
		JZ("lowercase").
		PushInt('A').
		Jmp("lcend").
		Label("lowercase").
		PushInt('a').
		Label("lcend"). // (shift ? 'A' : 'a')
		PushArg(5).
		Add().
		PushInt(int64(kb.KeyA)).
		Sub().
		String().
		Concat().
		SetArg(1). // var1 += string((shift ? 'A' : 'a') + c - kb.KeyA)
		Label("continue").
		PushArg(4).
		PushInt(2).
		Add().
		SetArg(4). // i += 2
		Jmp("loop").
		Label("done").
		PushString(string(con.PUTS)).
		PushArg(1).
		Concat().
		PushString("\n").
		Concat().
		Write(drivers.CONSOLE). // 将 var1 通过 console 设备输出：PUTS var1+"\n"
		PushArg(1)
	ret := run(asm).Top(1)
	if notEq(ret, "Hello") {
		t.Fatal("TestKeyboardAndConsole:", ret)
	}
}

/*
	var1 := ""
	var2 := make([]byte, 128)
	nread := Read(drivers.KEYBOARD, var2)
	shift := false
	for i := 0; i < nread; i += 2 {
		c := var2[i+1]
		t := var2[i]
		if c == kb.KeyShift {
			shift = (t == kb.KEYDOWN)
		} else if t == kb.KEYDOWN {
			var1 += string((shift ? 'A' : 'a') + c - kb.KeyA)
		}
	}
*/
