package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"peterdekok.nl/gotools/logger"
	"reflect"
	"sync"
	"time"
)

type model struct {
	Id        uuid.UUID
	*timestamps
}

type Model struct {
	m *model
	i Interface
	c CollectionInterface

	name string
	log  *logrus.Entry

	sync.Mutex
}

type timestamps struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

type marshaller struct {
	Model    interface{}
	Instance interface{}
}

var (
	log logger.Logger
)

func init() {
	log = logger.New("borm.model")
}

func Embed(i Interface, c CollectionInterface) (Interface, error) {
	iv, fv, err := CheckInterface(i)

	if err != nil {
		log.WithError(err).Error("Failed to embed model")

		return nil, fmt.Errorf("failed to embed model: %s", err)
	}

	id := uuid.New()

	name := iv.Type().Name()

	m := &Model{
		m: &model{
			Id:        id,
			timestamps: &timestamps{
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
				DeletedAt: time.Time{},
			},
		},
		i: i,
		c: c,

		name: name,
		log:  log.WithField("id", id).WithField("model", name),
	}

	mv := reflect.ValueOf(m)

	fv.Set(mv.Elem())

	return i, nil
}

func Unmarshal(b []byte, i Interface, c CollectionInterface) (Interface, error) {
	if _, err := Embed(i, c); err != nil {
		return nil, err
	}

	return i, i.Unmarshal(b)
}

func (m *Model) Marshal() ([]byte, error) {
	if m == nil || m.m == nil {
		return nil, errors.New("failed to marshal, nil receiver")
	}

	return json.Marshal(&marshaller{
		Model:    m.m,
		Instance: m.i,
	})
}

func (m *Model) Unmarshal(b []byte) error {
	if m == nil || m.m == nil {
		return errors.New("failed to unmarshal, nil receiver")
	}

	return json.Unmarshal(b, &marshaller{
		Model:    m.m,
		Instance: m.i,
	})
}

func (m *Model) Id() uuid.UUID {
	if m == nil || m.m == nil {
		return uuid.UUID{}
	}

	return m.m.Id
}

func (m *Model) CreatedAt() time.Time {
	if m == nil || m.m == nil {
		return time.Time{}
	}

	return m.m.CreatedAt
}

func (m *Model) UpdatedAt() time.Time {
	if m == nil || m.m == nil {
		return time.Time{}
	}

	return m.m.UpdatedAt
}

func (m *Model) DeletedAt() time.Time {
	if m == nil || m.m == nil {
		return time.Time{}
	}

	return m.m.DeletedAt
}

func (m *Model) Collection() CollectionInterface {
	if m == nil {
		return nil
	}

	return m.c
}

func (m *Model) Exists() bool {
	if m == nil || m.m == nil {
		return false
	}

	return !m.m.CreatedAt.IsZero()
}

func (m *Model) Save() error {
	if m == nil || m.m == nil {
		err := errors.New("model not initialized")

		log.WithError(err).Error("Failed to save model")

		return fmt.Errorf("failed to save model: %s", err)
	}

	m.Lock()
	defer m.Unlock()

	backup := m.m.BackupTimestamps()

	m.m.UpdatedAt = time.Now()

	if !m.Exists() {
		m.m.CreatedAt = m.m.UpdatedAt
	}

	if err := m.c.Save(m); err != nil {
		m.m.RestoreTimestamps(backup)

		m.log.WithError(err).Error("Failed to save model")

		return fmt.Errorf("failed to save model: %s", err)
	}

	return nil
}

func (t *timestamps) BackupTimestamps() timestamps {
	return timestamps{
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
		DeletedAt: t.DeletedAt,
	}
}

func (t *timestamps) RestoreTimestamps(backup timestamps) {
	t.CreatedAt = backup.CreatedAt
	t.UpdatedAt = backup.UpdatedAt
	t.DeletedAt = backup.DeletedAt
}