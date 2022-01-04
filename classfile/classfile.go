package classfile

type Classfile struct {
	Magic          uint32
	Minor          uint16
	Major          uint16
	ConstantPool   ConstantPool
	AccessFlags    uint16
	ThisClass      uint16
	SuperClass     uint16
	Interfaces     []uint16
	Fields         []*MemberInfo
	Methods        []*MemberInfo
	AttributeTable *AttributeTable
}

type MemberInfo struct {
	AccessFlags     uint16
	NameIndex       uint16
	DescriptorIndex uint16
	AttributeTable  *AttributeTable
}

type ConstantPool []ConstantInfo

type ConstantInfo interface {
	Tag() ConstantInfoTag
}

type ConstantInfoTag uint8

const (
	ConstantUnknown            ConstantInfoTag = 0 // default
	ConstantUtf8                               = 1
	ConstantInteger                            = 3
	ConstantFloat                              = 4
	ConstantLong                               = 5
	ConstantDouble                             = 6
	ConstantClass                              = 7
	ConstantString                             = 8
	ConstantFieldref                           = 9
	ConstantMethodref                          = 10
	ConstantInterfaceMethodref                 = 11
	ConstantNameAndType                        = 12
	ConstantMethodHandle                       = 15
	ConstantMethodType                         = 16
	ConstantInvokeDynamic                      = 18
)

// constantInfoBase is shared by all constant info objects.
// Currently, this only holds a tag and thus implements the
// ConstantInfo interface. Embed this into the constant info
// structs.
type constantInfoBase struct {
	tag ConstantInfoTag
}

func (c constantInfoBase) Tag() ConstantInfoTag {
	return c.tag
}

// contant pool structs
type (
	ConstantUtf8Info struct {
		constantInfoBase
		Value string
	}

	ConstantIntegerInfo struct {
		constantInfoBase
		Value int32
	}

	ConstantFloatInfo struct {
		constantInfoBase
		Value float32
	}

	ConstantLongInfo struct {
		constantInfoBase
		Value int64
	}

	ConstantDoubleInfo struct {
		constantInfoBase
		Value float64
	}

	ConstantClassInfo struct {
		constantInfoBase
		NameIndex uint16
	}

	ConstantStringInfo struct {
		constantInfoBase
		StringIndex uint16
	}

	ConstantFieldrefInfo struct {
		constantMemberrefInfo
	}

	ConstantMethodrefInfo struct {
		constantMemberrefInfo
	}

	ConstantInterfaceMethodrefInfo struct {
		constantMemberrefInfo
	}

	constantMemberrefInfo struct {
		constantInfoBase
		ClassIndex       uint16
		NameAndTypeIndex uint16
	}

	ConstantNameAndTypeInfo struct {
		constantInfoBase
		NameIndex       uint16
		DescriptorIndex uint16
	}

	ConstantMethodHandleInfo struct {
		constantInfoBase
		ReferenceKind  uint8
		ReferenceIndex uint16
	}

	ConstantMethodTypeInfo struct {
		constantInfoBase
		DescriptorIndex uint16
	}

	ConstantInvokeDynamicInfo struct {
		constantInfoBase
		BootstrapMethodAttrIndex uint16
		NameAndTypeIndex         uint16
	}
)
