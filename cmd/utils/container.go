package utils

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

type Initializer[T interface{}] func(ctx context.Context) T

// Contains lazily generated singletons of stateless struct used throughout
// the application. Thread-safe.
type DependencyContainer struct {
	singletonInitializers *sync.Map
}

func NewDependencyContainer() *DependencyContainer {
	return &DependencyContainer{
		singletonInitializers: &sync.Map{},
	}
}

func (d *DependencyContainer) ListRegisteredTypes() []reflect.Type {
	typeList := []reflect.Type{}
	d.singletonInitializers.Range(func(key any, value any) bool {
		asType, ok := key.(reflect.Type)
		if !ok {
			panic("key to singletonInitializers was not 'reflect.Type'")
		}
		typeList = append(typeList, asType)
		return true
	})
	return typeList
}

func (d *DependencyContainer) InvokeAll(ctx context.Context) {
	d.singletonInitializers.Range(func(key any, value any) bool {
		reflectValue := reflect.ValueOf(value)
		reflectValue.Call([]reflect.Value{reflect.ValueOf(ctx)})
		return true
	})
}

// These *should* be generic methods of DependencyContainer, but generic methods
// aren't supported in Go... (ノಠ益ಠ)ノ彡┻━┻

func wrapFuncForSingleton[T interface{}](initializer Initializer[T]) Initializer[T] {
	var t T
	var once sync.Once
	return func(ctx context.Context) T {
		once.Do(func() {
			t = initializer(ctx)
		})
		return t
	}
}

func registerDependency[T interface{}](d *DependencyContainer, initializer Initializer[T]) {
	tType := reflect.TypeOf((*T)(nil)).Elem()
	d.singletonInitializers.Store(tType, wrapFuncForSingleton(initializer))
}

func GetDependency[T interface{}](ctx context.Context, d *DependencyContainer) T {
	tType := reflect.TypeOf((*T)(nil)).Elem()

	if initializer, ok := d.singletonInitializers.Load(tType); !ok {
		panic(fmt.Sprintf("no initializer registered for type '%s'", tType))
	} else {
		return initializer.(Initializer[T])(ctx)
	}
}
