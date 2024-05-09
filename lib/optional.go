package lib

import "fmt"

type Optional[T any] struct {
	Value T
	Has   bool
}

func NewOptional[T any](value T) Optional[T] {
	return Optional[T]{Value: value, Has: true}
}

func (o Optional[T]) IsPresent() bool {
	return o.Has
}

func (o Optional[T]) Get() T {
	return o.Value
}

func (o Optional[T]) GetOrDefault(defaultValue T) T {
	if o.IsPresent() {
		return o.Get()
	}
	return defaultValue
}

func (o Optional[T]) OrElseGet(supplier func() T) T {
	if o.IsPresent() {
		return o.Get()
	}
	return supplier()
}

func (o Optional[T]) OrElse(other T) T {
	return o.GetOrDefault(other)
}

func (o Optional[T]) String() string {
	if o.Has {
		return fmt.Sprint(o.Value)
	}
	return ""
}
