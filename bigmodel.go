package bigmodel

import (
	"errors"
	"fmt"
	"reflect"
)

type DataSourceFactory interface {
	Source() interface{}
}

type StaticDataSourceFactory struct {
	source interface{}
}

func (f *StaticDataSourceFactory) Source() interface{} {
	return f.source
}

type LazyDataSourceFactory struct {
	makeSource func() interface{}
}

func (f *LazyDataSourceFactory) Source() interface{} {
	return f.makeSource()
}

type DataSourceFactoryManager struct {
	allowCache      bool
	sourceFactories map[string]DataSourceFactory
}

func NewDataSourceFactoryManager() *DataSourceFactoryManager {
	return &DataSourceFactoryManager{
		sourceFactories: make(map[string]DataSourceFactory),
	}
}

func (m *DataSourceFactoryManager) SetAllowCache(allowCache bool) *DataSourceFactoryManager {
	m.allowCache = allowCache
	return m
}

func (m *DataSourceFactoryManager) WithSource(name string, s interface{}) *DataSourceFactoryManager {
	m.sourceFactories[name] = &StaticDataSourceFactory{source: s}
	return m
}

func (m *DataSourceFactoryManager) WithFactory(name string, f func() interface{}) *DataSourceFactoryManager {
	m.sourceFactories[name] = &LazyDataSourceFactory{makeSource: f}
	return m
}

type (
	GetString func() string
	GetInt    func() int
	GetFloat  func() float64
)

type Getter interface {
	Get(string) interface{}
}

var (
	ErrNotPtrStruct = errors.New("not a pointer to struct")
)

func InitModel(ps interface{}, mgr *DataSourceFactoryManager) error {
	v := reflect.ValueOf(ps)
	if v.Kind() != reflect.Ptr && v.Elem().Kind() != reflect.Struct {
		return ErrNotPtrStruct
	}
	v = v.Elem()
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		ft := f.Type
		if ft.Kind() != reflect.Func || ft.NumIn() != 0 || ft.NumOut() != 1 {
			panic("invalid field " + f.Name)
		}
		sourceName := f.Tag.Get("source")
		if sourceName == "" {
			panic("cannot find source for field " + f.Name)
		}
		fieldName := f.Tag.Get("field")
		if fieldName == "" {
			fieldName = f.Name
		}
		vf := v.Field(i)
		var cachedValue reflect.Value
		hasCached := false
		sourceFactory, factoryFound := mgr.sourceFactories[sourceName]
		if !factoryFound {
			return fmt.Errorf("no factory for source %s", sourceName)
		}
		var source interface{}
		if sdsf, ok := sourceFactory.(*StaticDataSourceFactory); ok {
			source = sdsf.Source()
		}
		fieldFunc := reflect.MakeFunc(ft, func(args []reflect.Value) []reflect.Value {
			rv := make([]reflect.Value, 1)
			if mgr.allowCache && hasCached {
				rv[0] = cachedValue
				return rv
			}
			if source == nil {
				source = sourceFactory.Source()
			}
			switch typedSource := source.(type) {
			case Getter:
				rv[0] = reflect.ValueOf(typedSource.Get(fieldName))
			default:
				sv := reflect.ValueOf(source)
				if sv.Kind() == reflect.Ptr {
					sv = sv.Elem()
				}
				rv[0] = sv.FieldByName(fieldName)
			}
			if mgr.allowCache {
				cachedValue = rv[0]
				hasCached = true
			}
			return rv
		})
		vf.Set(fieldFunc)
	}
	return nil
}
