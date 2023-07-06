package utils

import (
	"errors"
	"reflect"
)

func AnyMatch(slice interface{}, condition func(x interface{}) bool) (bool, error) {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return false, errors.New("first arg. is not a slice")
	}

	for i := 0; i < v.Len(); i++ {
		if condition(v.Index(i).Interface()) {
			return true, nil
		}
	}
	return false, nil
}

func AllMatch(slice interface{}, condition func(x interface{}) bool) (bool, error) {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return false, errors.New("first arg. is not a slice")
	}

	for i := 0; i < v.Len(); i++ {
		if !condition(v.Index(i).Interface()) {
			return false, nil
		}
	}
	return true, nil
}

func Filter[T comparable](
	slice []T,
	condition func(x T) bool,
) []T {
	var newSlice []T
	for _, v := range slice {
		if condition(v) {
			newSlice = append(newSlice, v)
		}
	}
	return newSlice
}
