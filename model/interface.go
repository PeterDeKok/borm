package model

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"reflect"
	"sync"
	"time"
)

type Interface interface {
	Id() uuid.UUID
	CreatedAt() time.Time
	UpdatedAt() time.Time
	DeletedAt() time.Time

	Collection() CollectionInterface
	Exists() bool

	Marshal() ([]byte, error)
	Unmarshal(b []byte) error
	Save() error

	sync.Locker
}

type CollectionInterface interface {
	Load() error
	Create(i Interface) error
	Save(i Interface) error
}

func CheckInterface(i Interface) (reflect.Value, reflect.Value, error) {
	iv, err := getInterfaceValue(i)

	if err != nil {
		return iv, reflect.Value{}, fmt.Errorf("invalid model type: %s: expected pointer to named struct", err)
	}

	fv, err := getEmbeddedModelField(iv)

	if err != nil {
		return iv, fv, fmt.Errorf("invalid model type: %s: it can not embed new Model: %s", iv.Type(), err)
	}

	return iv, fv, nil
}

func getInterfaceValue(i Interface) (reflect.Value, error) {
	iv := reflect.ValueOf(i)

	if iv.Kind() != reflect.Ptr {
		return iv, fmt.Errorf("%s (%s)", iv.Type(), iv.Kind().String())
	}

	iv = iv.Elem()

	// The pointer should point to a named struct, anonymous structs are not allowed
	if iv.Kind() != reflect.Struct || iv.Type().Name() == "" {
		return iv, errors.New(fmt.Sprintf("%s (%s to %s)", iv.Type(), reflect.Ptr, iv.Kind()))
	}

	return iv, nil
}

func getEmbeddedModelField(iv reflect.Value) (reflect.Value, error) {
	mt := reflect.TypeOf((*Model)(nil)).Elem()

	fv := iv.FieldByName(mt.Name())

	if !fv.IsValid() {
		return fv, errors.New("field missing")
	}

	if fv.Kind() == reflect.Ptr {
		if !reflect.PtrTo(mt).AssignableTo(fv.Type()) {
			return fv, errors.New("field type invalid")
		}

		fv.Set(reflect.New(mt))

		fv = fv.Elem()
	}

	if !fv.IsZero() || !fv.CanSet() {
		return fv, errors.New("field already initialized")
	}

	if !mt.AssignableTo(fv.Type()) {
		return fv, errors.New("field type invalid")
	}

	return fv, nil
}
