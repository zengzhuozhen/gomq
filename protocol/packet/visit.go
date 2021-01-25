package packet

// k8s visitor implement

type Visitor interface {
	Visit(VisitorFunc) error
}

type VisitorFunc func(ControlPacket) error

type VisitorList []Visitor

func (v VisitorList) Visit(fn VisitorFunc) error {
	for i := range v {
		if err := v[i].Visit(fn); err != nil {
			return err
		}
	}
	return nil
}

type DecoratedVisitor struct {
	visitor    Visitor
	decorators []VisitorFunc
}

func (v DecoratedVisitor) Visit(fn VisitorFunc) error {
	return v.visitor.Visit(func(packet ControlPacket) error {
		for i := range v.decorators {
			if err := v.decorators[i](packet); err != nil {
				return err
			}
		}
		return fn(packet)
	})
}

func NewDecoratedVisitor(v Visitor, fn ...VisitorFunc) Visitor {
	if len(fn) == 0 {
		return v
	}
	return DecoratedVisitor{v, fn}
}

type FilteredVisitor struct {
	visitor Visitor
	filters []VisitorFunc
}

func (v FilteredVisitor) Visit(fn VisitorFunc) error {
	return v.visitor.Visit(func(packet ControlPacket) error {
		for _, filter := range v.filters {
			err := filter(packet)
			if err != nil {
				return err
			}
		}
		return fn(packet)
	})
}

func NewFilteredVisitor(v Visitor, fn ...VisitorFunc) Visitor {
	if len(fn) == 0 {
		return v
	}
	return FilteredVisitor{v, fn}
}