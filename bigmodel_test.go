package bigmodel

import (
	"testing"
)

type HelloModel struct {
	ID       GetInt    `source:"A" field:"UserId"`
	UserName GetString `source:"B"`
}

type DemoSource1 struct {
	UserId int
}

type DemoSource2 struct{}

func (DemoSource2) Get(name string) interface{} {
	if name == "UserName" {
		return "guanming"
	}
	return "x"
}

func TestHello(t *testing.T) {
	var m HelloModel
	InitModel(&m, NewDataSourceFactoryManager().
		WithSource("A", &DemoSource1{UserId: 100}).
		WithFactory("B", func() interface{} { return DemoSource2{} }),
	)
	if m.ID() != 100 {
		t.Errorf("get id error expected %d got %d", 100, m.ID())
	}
	if m.UserName() != "guanming" {
		t.Errorf("get username error expected %s got %s", "guanming", m.UserName())
	}
	if m.ID() != 100 {
		t.Errorf("get id error expected %d got %d", 100, m.ID())
	}
	if m.UserName() != "guanming" {
		t.Errorf("get username error expected %s got %s", "guanming", m.UserName())
	}
	if m.ID() != 100 {
		t.Errorf("get id error expected %d got %d", 100, m.ID())
	}
	if m.UserName() != "guanming" {
		t.Errorf("get username error expected %s got %s", "guanming", m.UserName())
	}
}

func BenchmarkAllowCache(b *testing.B) {
	var m HelloModel
	InitModel(&m, NewDataSourceFactoryManager().
		WithSource("A", &DemoSource1{UserId: 100}).
		WithFactory("B", func() interface{} { return DemoSource2{} }),
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.ID()
		m.UserName()
	}
}

func BenchmarkNotAllowCache(b *testing.B) {
	var m HelloModel
	InitModel(&m, NewDataSourceFactoryManager().
		SetAllowCache(false).
		WithSource("A", &DemoSource1{UserId: 100}).
		WithFactory("B", func() interface{} { return DemoSource2{} }),
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.ID()
		m.UserName()
	}
}
