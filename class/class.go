package class

import (
	"fmt"
	"io"

	"github.com/tsatke/jt/classfile"
)

type Class struct {
	cf *classfile.Classfile
}

func ParseClass(rd io.Reader) (*Class, error) {
	cf, err := classfile.Parse(rd)
	if err != nil {
		return nil, fmt.Errorf("parse classfile: %w", err)
	}
	return &Class{cf}, nil
}

func (c Class) Name() string {
	thisClassInfo := c.cf.ConstantPool[c.cf.ThisClass].(*classfile.ConstantClassInfo)
	return c.cf.ConstantPool[thisClassInfo.NameIndex].(*classfile.ConstantUtf8Info).Value
}

func (c Class) Version() (int, int) {
	return int(c.cf.Major), int(c.cf.Minor)
}

func (c Class) SuperclassName() string {
	if c.cf.SuperClass == 0 {
		// class is java/lang/Object
		return ""
	}
	return c.cf.ConstantPool[c.cf.ConstantPool[c.cf.SuperClass].(*classfile.ConstantClassInfo).NameIndex].(*classfile.ConstantUtf8Info).Value
}

func (c Class) Methods() []Method {
	methods := make([]Method, len(c.cf.Methods))
	for i := range methods {
		methods[i] = Method{member{
			info: c.cf.Methods[i],
			cf:   c.cf,
		}}
	}
	return methods
}
