package jt

import "github.com/tsatke/jt/classfile"

type Method struct {
	member
}

type Field struct {
	member
}

type member struct {
	info *classfile.MemberInfo
	cf   *classfile.Classfile
}

func (m member) Name() string {
	return m.cf.ConstantPool[m.info.NameIndex].(*classfile.ConstantUtf8Info).Value
}
