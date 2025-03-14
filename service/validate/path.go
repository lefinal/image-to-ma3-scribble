package validate

import (
	"bytes"
	"fmt"
	"strconv"
)

// Path represents the path from some root to a particular field.
type Path struct {
	// name of this field or empty if this is an index.
	name string
	// index is a subscript of the previous element if name is empty.
	index string
	// parent is nil if this is the root element.
	parent *Path
}

// NewPath creates a root Path object.
func NewPath(name string, moreNames ...string) *Path {
	r := &Path{name: name, parent: nil} //nolint:exhaustruct
	for _, anotherName := range moreNames {
		r = &Path{name: anotherName, parent: r} //nolint:exhaustruct
	}
	return r
}

// Root returns the root element of this Path.
func (p *Path) Root() *Path {
	for ; p.parent != nil; p = p.parent {
		// Do nothing.
	}
	return p
}

// Child creates a new Path that is a child of the method receiver.
func (p *Path) Child(name string, moreNames ...string) *Path {
	r := NewPath(name, moreNames...)
	r.Root().parent = p
	return r
}

// Index indicates that the previous Path is to be subscripted by an int.
// This sets the same underlying value as Key.
func (p *Path) Index(index int) *Path {
	//nolint:exhaustruct
	return &Path{index: strconv.Itoa(index), parent: p}
}

// Key indicates that the previous Path is to be subscripted by a string.
// This sets the same underlying value as Index.
func (p *Path) Key(key string) *Path {
	//nolint:exhaustruct
	return &Path{index: key, parent: p}
}

// String produces a string representation of the Path.
func (p *Path) String() string {
	if p == nil {
		return "<nil>"
	}
	// make a slice to iterate
	elems := make([]*Path, 0)
	for ; p != nil; p = p.parent {
		elems = append(elems, p)
	}

	// iterate, but it has to be backwards
	buf := bytes.NewBuffer(nil)
	for i := range elems {
		p := elems[len(elems)-1-i]
		if p.parent != nil && len(p.name) > 0 {
			// This is either the root or it is a subscript.
			buf.WriteString(".")
		}
		if len(p.name) > 0 {
			buf.WriteString(p.name)
		} else {
			_, _ = fmt.Fprintf(buf, "[%s]", p.index)
		}
	}
	return buf.String()
}
