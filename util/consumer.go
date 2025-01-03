package util

type ConsumeFunc[T any] func(item T) (err error)

type ConsumeBatchFunc[T any] func(items []T) (count uint32, err error)
