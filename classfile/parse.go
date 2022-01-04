package classfile

import (
	"encoding/binary"
	"fmt"
	"io"
	"unicode/utf16"
)

func Parse(rd io.Reader) (cf *Classfile, err error) {
	defer func() {
		recv := recover()
		if recv != nil {
			if pe, ok := recv.(error); ok {
				err = pe
			} else {
				err = fmt.Errorf("fatal parse error: %v", recv)
			}
		}
	}()
	return parse(newContentReader(rd, binary.BigEndian)), err
}

func parse(rd *contentReader) *Classfile {
	magic := rd.uint32()
	if magic != 0xCAFEBABE {
		panic(fmt.Errorf("invalid magic value: 0x%08X", magic))
	}
	minor := rd.uint16()
	major := rd.uint16()
	constantPool := parseConstantPool(rd)
	accessFlags := rd.uint16()
	thisClass := rd.uint16()
	superClass := rd.uint16()
	n := rd.uint16()
	interfaces := make([]uint16, n)
	for i := range interfaces {
		interfaces[i] = rd.uint16()
	}

	fields := parseMemberInfos(rd, constantPool)
	methods := parseMemberInfos(rd, constantPool)

	attributeTable := parseAttributeTable(rd, constantPool)

	return &Classfile{
		Magic:          magic,
		Minor:          minor,
		Major:          major,
		ConstantPool:   constantPool,
		AccessFlags:    accessFlags,
		ThisClass:      thisClass,
		SuperClass:     superClass,
		Interfaces:     interfaces,
		Fields:         fields,
		Methods:        methods,
		AttributeTable: attributeTable,
	}
}

func parseConstantPool(rd *contentReader) ConstantPool {
	count := int(rd.uint16())
	pool := ConstantPool(make([]ConstantInfo, count))
	for i := 1; i < count; i++ {
		pool[i] = parseConstantInfo(rd, pool)
		// TODO: maybe increment i if info was long or double?
	}
	return pool
}

func parseConstantInfo(rd *contentReader, pool ConstantPool) ConstantInfo {
	tag := ConstantInfoTag(rd.uint8())
	switch tag {
	case ConstantUtf8:
		return parseConstantUtf8Info(rd)
	case ConstantInteger:
		return &ConstantIntegerInfo{
			constantInfoBase{tag},
			int32(rd.uint32()),
		}
	case ConstantFloat:
		return &ConstantFloatInfo{
			constantInfoBase{tag},
			rd.float32(),
		}
	case ConstantLong:
		return &ConstantLongInfo{
			constantInfoBase{tag},
			int64(rd.uint64()),
		}
	case ConstantDouble:
		return &ConstantDoubleInfo{
			constantInfoBase{tag},
			rd.float64(),
		}
	case ConstantClass:
		return &ConstantClassInfo{
			constantInfoBase{tag},
			rd.uint16(),
		}
	case ConstantString:
		return &ConstantStringInfo{
			constantInfoBase{tag},
			rd.uint16(),
		}
	case ConstantFieldref:
		return &ConstantFieldrefInfo{
			constantMemberrefInfo{
				constantInfoBase{tag},
				rd.uint16(),
				rd.uint16(),
			},
		}
	case ConstantMethodref:
		return &ConstantMethodrefInfo{
			constantMemberrefInfo{
				constantInfoBase{tag},
				rd.uint16(),
				rd.uint16(),
			},
		}
	case ConstantInterfaceMethodref:
		return &ConstantInterfaceMethodrefInfo{
			constantMemberrefInfo{
				constantInfoBase{tag},
				rd.uint16(),
				rd.uint16(),
			},
		}
	case ConstantNameAndType:
		return &ConstantNameAndTypeInfo{
			constantInfoBase{tag},
			rd.uint16(),
			rd.uint16(),
		}
	case ConstantMethodHandle:
		return &ConstantMethodHandleInfo{
			constantInfoBase{tag},
			rd.uint8(),
			rd.uint16(),
		}
	case ConstantMethodType:
		return &ConstantMethodTypeInfo{
			constantInfoBase{tag},
			rd.uint16(),
		}
	case ConstantInvokeDynamic:
		return &ConstantInvokeDynamicInfo{
			constantInfoBase{tag},
			rd.uint16(),
			rd.uint16(),
		}
	}
	panic(fmt.Errorf("unknown constant info tag: %v", tag))
}

func parseConstantUtf8Info(rd *contentReader) *ConstantUtf8Info {
	length := rd.uint16()
	bytes := rd.raw(uint(length))
	strVal := decodeMUTF8(bytes)
	return &ConstantUtf8Info{
		constantInfoBase{ConstantUtf8},
		strVal,
	}
}

// decodeMUTF8 was borrowed from
// https://github.com/ianynchen/glass/blob/396a8585c72094d66e8d1d96657c635412a36be7/classfile/constant_info.go
// The panics are covered by the Parse method.
func decodeMUTF8(bytearr []byte) string {
	utflen := len(bytearr)
	chararr := make([]uint16, utflen)

	var c, char2, char3 uint16
	count := 0
	chararr_count := 0

	for count < utflen {
		c = uint16(bytearr[count])
		if c > 127 {
			break
		}
		count++
		chararr[chararr_count] = c
		chararr_count++
	}

	for count < utflen {
		c = uint16(bytearr[count])
		switch c >> 4 {
		case 0, 1, 2, 3, 4, 5, 6, 7:
			/* 0xxxxxxx*/
			count++
			chararr[chararr_count] = c
			chararr_count++
		case 12, 13:
			/* 110x xxxx   10xx xxxx*/
			count += 2
			if count > utflen {
				panic("malformed input: partial character at end")
			}
			char2 = uint16(bytearr[count-1])
			if char2&0xC0 != 0x80 {
				panic(fmt.Errorf("malformed input around byte %v", count))
			}
			chararr[chararr_count] = c&0x1F<<6 | char2&0x3F
			chararr_count++
		case 14:
			/* 1110 xxxx  10xx xxxx  10xx xxxx*/
			count += 3
			if count > utflen {
				panic("malformed input: partial character at end")
			}
			char2 = uint16(bytearr[count-2])
			char3 = uint16(bytearr[count-1])
			if char2&0xC0 != 0x80 || char3&0xC0 != 0x80 {
				panic(fmt.Errorf("malformed input around byte %v", (count - 1)))
			}
			chararr[chararr_count] = c&0x0F<<12 | char2&0x3F<<6 | char3&0x3F<<0
			chararr_count++
		default:
			/* 10xx xxxx,  1111 xxxx */
			panic(fmt.Errorf("malformed input around byte %v", count))
		}
	}
	// The number of chars produced may be less than utflen
	chararr = chararr[0:chararr_count]
	runes := utf16.Decode(chararr)
	return string(runes)
}

func parseMemberInfos(rd *contentReader, pool ConstantPool) []*MemberInfo {
	count := rd.uint16()
	members := make([]*MemberInfo, count)
	for i := range members {
		members[i] = parseMemberInfo(rd, pool)
	}
	return members
}

func parseMemberInfo(rd *contentReader, pool ConstantPool) *MemberInfo {
	return &MemberInfo{
		AccessFlags:     rd.uint16(),
		NameIndex:       rd.uint16(),
		DescriptorIndex: rd.uint16(),
		AttributeTable:  parseAttributeTable(rd, pool),
	}
}

func parseAttributeTable(rd *contentReader, pool ConstantPool) *AttributeTable {
	attributeCount := rd.uint16()
	table := &AttributeTable{
		attributes: make([]Attribute, attributeCount),
	}

	for i := range table.attributes {
		table.attributes[i] = parseAttribute(rd, pool)
	}

	return table
}

func parseAttribute(rd *contentReader, pool ConstantPool) Attribute {
	attributeNameIndex := rd.uint16()
	attributeLength := rd.uint32()
	attributeName := pool[attributeNameIndex].(*ConstantUtf8Info).Value

	switch attributeName {
	case "BootstrapMethods":
	case "Code":
		return parseCodeAttribute(rd, pool)
	case "ConstantValue":
	case "Deprecated":
	case "EnclosingMethod":
	case "Exceptions":
	case "InnerClasses":
	case "LineNumberTable":
	case "LocalVariableTable":
	case "LocalVariableTypeTable":
	case "MethodParameters":
	case "RuntimeInvisibleAnnotations":
	case "RuntimeInvisibleParameterAnnotation":
	case "RuntimeInvisibleHypeAnnotations":
	case "RuntimeVisibleAnnotations":
	case "RuntimeVisibleParameterAnnotations":
	case "RuntimeVisibleTypeAnnotations":
	case "Signature":
	case "SourceFile":
	case "SourceDebugExtension":
	case "StackMapTable":
	case "Synthetic":
	}

	return UnknownAttribute{
		Name:    attributeName,
		Length:  attributeLength,
		Payload: rd.raw(uint(attributeLength)),
	}
}

func parseCodeAttribute(rd *contentReader, pool ConstantPool) *CodeAttribute {
	maxStack := rd.uint16()
	maxLocals := rd.uint16()
	codeLength := rd.uint32()
	code := rd.raw(uint(codeLength))
	exceptionTable := parseExceptionTable(rd)
	attributeTable := parseAttributeTable(rd, pool)

	return &CodeAttribute{
		Pool:           pool,
		MaxStack:       maxStack,
		MaxLocals:      maxLocals,
		Code:           code,
		ExceptionTable: exceptionTable,
		Attributes:     attributeTable,
	}
}

func parseExceptionTable(rd *contentReader) ExceptionTable {
	length := rd.uint16()
	table := make(ExceptionTable, length)
	for i := range table {
		entry := ExceptionTableEntry{
			StartPc:   rd.uint16(),
			EndPc:     rd.uint16(),
			HandlerPc: rd.uint16(),
			CatchType: rd.uint16(),
		}
		table[i] = entry
	}
	return table
}
