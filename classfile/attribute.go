package classfile

type AttributeTable struct {
	attributes []Attribute
}

type Attribute interface{} // interface to make it possible to add attributes externally in the future

type UnknownAttribute struct {
	Name    string
	Length  uint32
	Payload []byte
}

type CodeAttribute struct {
	Pool           ConstantPool
	MaxStack       uint16
	MaxLocals      uint16
	Code           []byte
	ExceptionTable ExceptionTable
	Attributes     *AttributeTable
}

type ExceptionTable []ExceptionTableEntry

type ExceptionTableEntry struct {
	StartPc   uint16
	EndPc     uint16
	HandlerPc uint16
	CatchType uint16
}
