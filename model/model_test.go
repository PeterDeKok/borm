package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	. "peterdekok.nl/gotools/test"
	"testing"
	"time"
)

type TestModelStruct struct {
	Model
	FieldA string
	FieldB int
}
type TestModelPtr struct{ *Model }

type TestModelInterfaceNoModel struct{}

func (m TestModelInterfaceNoModel) Id() uuid.UUID                   { return uuid.New() }
func (m TestModelInterfaceNoModel) CreatedAt() time.Time            { return time.Time{} }
func (m TestModelInterfaceNoModel) UpdatedAt() time.Time            { return time.Time{} }
func (m TestModelInterfaceNoModel) DeletedAt() time.Time            { return time.Time{} }
func (m TestModelInterfaceNoModel) Exists() bool                    { return true }
func (m TestModelInterfaceNoModel) Collection() CollectionInterface { return nil }
func (m TestModelInterfaceNoModel) Marshal() ([]byte, error)        { return nil, nil }
func (m TestModelInterfaceNoModel) Unmarshal(b []byte) error        { return nil }
func (m TestModelInterfaceNoModel) Save() error                     { return errors.New("error") }
func (m TestModelInterfaceNoModel) Lock()                           {}
func (m TestModelInterfaceNoModel) Unlock()                         {}

type TestModelPtrToInt int

func (m *TestModelPtrToInt) Id() uuid.UUID                   { return uuid.New() }
func (m *TestModelPtrToInt) CreatedAt() time.Time            { return time.Time{} }
func (m *TestModelPtrToInt) UpdatedAt() time.Time            { return time.Time{} }
func (m *TestModelPtrToInt) DeletedAt() time.Time            { return time.Time{} }
func (m *TestModelPtrToInt) Exists() bool                    { return true }
func (m *TestModelPtrToInt) Collection() CollectionInterface { return nil }
func (m *TestModelPtrToInt) Marshal() ([]byte, error)        { return nil, nil }
func (m *TestModelPtrToInt) Unmarshal(_ []byte) error        { return nil }
func (m *TestModelPtrToInt) Save() error                     { return errors.New("error") }
func (m *TestModelPtrToInt) Lock()                           {}
func (m *TestModelPtrToInt) Unlock()                         {}

type TestModelStructWrongTypeEmbed struct{ Model int }

func (m *TestModelStructWrongTypeEmbed) Id() uuid.UUID                   { return uuid.New() }
func (m *TestModelStructWrongTypeEmbed) CreatedAt() time.Time            { return time.Time{} }
func (m *TestModelStructWrongTypeEmbed) UpdatedAt() time.Time            { return time.Time{} }
func (m *TestModelStructWrongTypeEmbed) DeletedAt() time.Time            { return time.Time{} }
func (m *TestModelStructWrongTypeEmbed) Exists() bool                    { return true }
func (m *TestModelStructWrongTypeEmbed) Collection() CollectionInterface { return nil }
func (m *TestModelStructWrongTypeEmbed) Marshal() ([]byte, error)        { return nil, nil }
func (m *TestModelStructWrongTypeEmbed) Unmarshal(_ []byte) error        { return nil }
func (m *TestModelStructWrongTypeEmbed) Save() error                     { return errors.New("error") }
func (m *TestModelStructWrongTypeEmbed) Lock()                           {}
func (m *TestModelStructWrongTypeEmbed) Unlock()                         {}

type TestModelPtrWrongTypeEmbed struct{ Model *int }

func (m *TestModelPtrWrongTypeEmbed) Id() uuid.UUID                   { return uuid.New() }
func (m *TestModelPtrWrongTypeEmbed) CreatedAt() time.Time            { return time.Time{} }
func (m *TestModelPtrWrongTypeEmbed) UpdatedAt() time.Time            { return time.Time{} }
func (m *TestModelPtrWrongTypeEmbed) DeletedAt() time.Time            { return time.Time{} }
func (m *TestModelPtrWrongTypeEmbed) Exists() bool                    { return true }
func (m *TestModelPtrWrongTypeEmbed) Collection() CollectionInterface { return nil }
func (m *TestModelPtrWrongTypeEmbed) Marshal() ([]byte, error)        { return nil, nil }
func (m *TestModelPtrWrongTypeEmbed) Unmarshal(_ []byte) error        { return nil }
func (m *TestModelPtrWrongTypeEmbed) Save() error                     { return errors.New("error") }
func (m *TestModelPtrWrongTypeEmbed) Lock()                           {}
func (m *TestModelPtrWrongTypeEmbed) Unlock()                         {}

type TestModelCollection struct{}

func (m *TestModelCollection) Load() error              { return nil }
func (m *TestModelCollection) Create(_ Interface) error { return nil }
func (m *TestModelCollection) Save(_ Interface) error   { return nil }

type TestModelCollectionError struct{}

func (m *TestModelCollectionError) Load() error              { return errors.New("error load") }
func (m *TestModelCollectionError) Create(_ Interface) error { return errors.New("error create") }
func (m *TestModelCollectionError) Save(_ Interface) error   { return errors.New("error save") }

func TestCheckInterface(t *testing.T) {
	var err error

	_, _, err = CheckInterface(TestModelInterfaceNoModel{})

	ExpectedError(t, err, "invalid model type: model.TestModelInterfaceNoModel (struct): expected pointer to named struct")

	_, _, err = CheckInterface(new(TestModelPtrToInt))

	ExpectedError(t, err, "invalid model type: model.TestModelPtrToInt (ptr to int): expected pointer to named struct")

	_, _, err = CheckInterface(&struct{ Model }{})

	ExpectedError(t, err, "invalid model type: struct { model.Model } (ptr to struct): expected pointer to named struct")

	_, _, err = CheckInterface(&TestModelInterfaceNoModel{})

	ExpectedError(t, err, "invalid model type: model.TestModelInterfaceNoModel: it can not embed new Model: field missing")

	_, _, err = CheckInterface(&TestModelStruct{Model: Model{name: "already initialized"}})

	ExpectedError(t, err, "invalid model type: model.TestModelStruct: it can not embed new Model: field already initialized")

	_, _, err = CheckInterface(&TestModelStructWrongTypeEmbed{})

	ExpectedError(t, err, "invalid model type: model.TestModelStructWrongTypeEmbed: it can not embed new Model: field type invalid")

	_, _, err = CheckInterface(&TestModelPtrWrongTypeEmbed{})

	ExpectedError(t, err, "invalid model type: model.TestModelPtrWrongTypeEmbed: it can not embed new Model: field type invalid")

	_, _, err = CheckInterface(&TestModelStruct{})

	ExpectedNoError(t, err)

	_, _, err = CheckInterface(&TestModelPtr{Model: &Model{}})

	ExpectedNoError(t, err)

	_, _, err = CheckInterface(&TestModelPtr{})

	ExpectedNoError(t, err)
}

func TestEmbed(t *testing.T) {
	var err error

	c := &TestModelCollection{}

	_, err = Embed(new(TestModelPtrToInt), c)

	ExpectedError(t, err, "failed to embed model: invalid model type: model.TestModelPtrToInt (ptr to int): expected pointer to named struct")

	m := &TestModelStruct{}

	i, err := Embed(m, c)

	ExpectedNoError(t, err)

	//noinspection GoVetCopyLock
	ExpectedNoZeroValue(t, m.Model)

	if i != m {
		t.Error("Embed should be fluent")
	}

	ExpectedNoZeroValue(t, i.Id())
	ExpectedZeroValue(t, i.CreatedAt())
	ExpectedZeroValue(t, i.UpdatedAt())
	ExpectedZeroValue(t, i.DeletedAt())
	ExpectedEqual(t, m.name, "TestModelStruct")
}

func TestModel_Collection(t *testing.T) {
	var i Interface = &TestModelPtr{}

	// nil receiver
	ExpectedZeroValue(t, i.Collection())

	c := &TestModelCollection{}

	i, err := Embed(&TestModelStruct{}, c)

	ExpectedNoError(t, err)

	ExpectedEqual(t, i.Collection(), c)
}

func TestModel_Marshal(t *testing.T) {
	var err error
	var b []byte

	c := &TestModelCollection{}

	m := &TestModelStruct{
		FieldA: "test-aaa",
		FieldB: 42,
	}

	_, err = m.Marshal()

	ExpectedError(t, err, "failed to marshal, nil receiver")

	_, err = Embed(m, c)

	ExpectedNoError(t, err)

	b, err = json.Marshal(m)

	ExpectedNoError(t, err)

	ExpectedEqual(t, b, []byte("{\"FieldA\":\"test-aaa\",\"FieldB\":42}"))

	b, err = m.Marshal()

	ExpectedNoError(t, err)

	bMap := &marshaller{
		Model:    &model{timestamps: &timestamps{}},
		Instance: &TestModelStruct{},
	}

	err = json.Unmarshal(b, &bMap)

	ExpectedNoError(t, err)

	ExpectedEqual(t, bMap.Model, m.m)

	m.Model = Model{}

	ExpectedEqual(t, bMap.Instance, m)
}

func TestModel_Unmarshal(t *testing.T) {
	var err error

	c := &TestModelCollection{}

	mA := &TestModelStruct{}

	err = mA.Unmarshal([]byte(""))

	ExpectedError(t, err, "failed to unmarshal, nil receiver")

	_, err = Embed(mA, c)

	ExpectedNoError(t, err)

	err = mA.Unmarshal([]byte(""))

	ExpectedError(t, err, "unexpected end of JSON input")

	id := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	date := "2020-07-01T01:01:00+02:00"
	eDate := "2020-07-01 01:01:00 +0200 CEST"
	fieldA := "test-aaa"
	fieldB := 42

	bA := fmt.Sprintf("{\"Model\":{\"Id\":\"%s\",\"CreatedAt\":\"%s\",\"UpdatedAt\":\"%s\",\"DeletedAt\":\"%s\"},\"Instance\":{\"FieldA\":\"%s\",\"FieldB\":%d}}", id, date, date, date, fieldA, fieldB)

	err = mA.Unmarshal([]byte(bA))

	ExpectedNoError(t, err)

	ExpectedEqual(t, mA.Id().String(), id)
	ExpectedEqual(t, mA.CreatedAt().String(), eDate)
	ExpectedEqual(t, mA.UpdatedAt().String(), eDate)
	ExpectedEqual(t, mA.DeletedAt().String(), eDate)
	ExpectedEqual(t, mA.FieldA, fieldA)
	ExpectedEqual(t, mA.FieldB, fieldB)

	mB := &TestModelStruct{}

	_, _ = Embed(mB, c)

	ExpectedNotEqual(t, mB.Id(), mA.Id())
	ExpectedNotEqual(t, mB.CreatedAt(), mA.CreatedAt())
	ExpectedNotEqual(t, mB.UpdatedAt(), mA.UpdatedAt())
	ExpectedNotEqual(t, mB.DeletedAt(), mA.DeletedAt())
	ExpectedNotEqual(t, mB.FieldA, mA.FieldA)
	ExpectedNotEqual(t, mB.FieldB, mA.FieldB)

	bB, err := mA.Marshal()

	err = mB.Unmarshal(bB)

	ExpectedNoError(t, err)

	ExpectedEqual(t, mB.Id(), mA.Id())
	ExpectedEqual(t, mB.CreatedAt(), mA.CreatedAt())
	ExpectedEqual(t, mB.UpdatedAt(), mA.UpdatedAt())
	ExpectedEqual(t, mB.DeletedAt(), mA.DeletedAt())
	ExpectedEqual(t, mB.FieldA, mA.FieldA)
	ExpectedEqual(t, mB.FieldB, mA.FieldB)
}

func TestModel_Save(t *testing.T) {
	var err error

	mA := &TestModelStruct{}

	err = mA.Save()

	ExpectedError(t, err, "failed to save model: model not initialized")

	iB, err := Embed(&TestModelStruct{}, &TestModelCollectionError{})

	ExpectedNoError(t, err)

	err = iB.Save()

	ExpectedError(t, err, "failed to save model: error save")

	ExpectedZeroValue(t, iB.CreatedAt())
	ExpectedZeroValue(t, iB.UpdatedAt())
	ExpectedZeroValue(t, iB.DeletedAt())

	c := &TestModelCollection{}
	mC := &TestModelStruct{}

	_, err = Embed(mC, c)

	ExpectedNoError(t, err)

	err = mC.Save()

	ExpectedNoError(t, err)

	ExpectedNoZeroValue(t, mC.CreatedAt())
	ExpectedNoZeroValue(t, mC.UpdatedAt())
	ExpectedZeroValue(t, mC.DeletedAt())

	ExpectedEqualF(t, mC.UpdatedAt().Equal(mC.CreatedAt()), true, false, "Created at and updated at should be equal")

	timestampsBackupC1 := mC.m.BackupTimestamps()

	err = mC.Save()

	ExpectedNoError(t, err)

	ExpectedNoZeroValue(t, mC.CreatedAt())
	ExpectedNoZeroValue(t, mC.UpdatedAt())
	ExpectedZeroValue(t, mC.DeletedAt())

	ExpectedEqualF(t, mC.CreatedAt().Equal(timestampsBackupC1.CreatedAt), true, false, "Created at should not be changed for existing model")
	ExpectedEqualF(t, mC.UpdatedAt().After(mC.CreatedAt()), true, false, "Updated at should be after created at")

	mC.c = &TestModelCollectionError{}

	timestampsBackupC2 := mC.m.BackupTimestamps()

	err = mC.Save()

	ExpectedError(t, err, "failed to save model: error save")

	ExpectedNoZeroValue(t, mC.CreatedAt())
	ExpectedNoZeroValue(t, mC.UpdatedAt())
	ExpectedZeroValue(t, mC.DeletedAt())

	ExpectedEqualF(t, mC.CreatedAt().Equal(timestampsBackupC2.CreatedAt), true, false, "timestamps should be restored on error")
	ExpectedEqualF(t, mC.UpdatedAt().Equal(timestampsBackupC2.UpdatedAt), true, false, "timestamps should be restored on error")
}

func TestModel_Exists(t *testing.T) {
	c := &TestModelCollection{}

	var i Interface = &TestModelStruct{}

	ExpectedZeroValue(t, i.Exists())

	m, err := Embed(i, c)

	ExpectedNoError(t, err)

	ExpectedZeroValueF(t, m.Exists(), false, "model was not saved, should not exist (yet)")

	err = m.Save()

	ExpectedNoError(t, err)

	ExpectedEqualF(t, m.Exists(), true, false, "model was saved, should exist")
}

func TestModel_Id(t *testing.T) {
	var err error

	m := &TestModelPtr{}
	c := &TestModelCollection{}

	ExpectedZeroValue(t, m.Id())

	_, err = Embed(m, c)

	ExpectedNoError(t, err)

	ExpectedEqual(t, m.Id().Variant(), uuid.RFC4122)
	ExpectedEqual(t, m.Id().Version(), uuid.Version(4))
}

func TestModel_CreatedAt(t *testing.T) {
	var err error

	m := &TestModelPtr{}
	c := &TestModelCollection{}

	ExpectedZeroValue(t, m.CreatedAt())

	_, err = Embed(m, c)

	ExpectedNoError(t, err)

	ExpectedZeroValue(t, m.CreatedAt())
}

func TestModel_UpdatedAt(t *testing.T) {
	var err error

	m := &TestModelPtr{}
	c := &TestModelCollection{}

	ExpectedZeroValue(t, m.UpdatedAt())

	_, err = Embed(m, c)

	ExpectedNoError(t, err)

	ExpectedZeroValue(t, m.UpdatedAt())
}

func TestModel_DeletedAt(t *testing.T) {
	var err error

	m := &TestModelPtr{}
	c := &TestModelCollection{}

	ExpectedZeroValue(t, m.DeletedAt())

	_, err = Embed(m, c)

	ExpectedNoError(t, err)

	ExpectedZeroValue(t, m.DeletedAt())
}

func TestUnmarshal(t *testing.T) {
	var err error

	c := &TestModelCollection{}

	mA := &TestModelStruct{}

	_, err = Unmarshal([]byte(""), mA, c)

	ExpectedError(t, err, "unexpected end of JSON input")

	id := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	date := "2020-01-01T01:01:00+01:00"
	eDate := "2020-01-01 01:01:00 +0100 CET"
	fieldA := "test-aaa"
	fieldB := 42

	bA := fmt.Sprintf("{\"Model\":{\"Id\":\"%s\",\"CreatedAt\":\"%s\",\"UpdatedAt\":\"%s\",\"DeletedAt\":\"%s\"},\"Instance\":{\"FieldA\":\"%s\",\"FieldB\":%d}}", id, date, date, date, fieldA, fieldB)

	mB := &TestModelStruct{}

	iB, err := Unmarshal([]byte(bA), mB, c)

	ExpectedNoError(t, err)

	ExpectedEqual(t, mB.Id().String(), id)
	ExpectedEqual(t, mB.CreatedAt().String(), eDate)
	ExpectedEqual(t, mB.UpdatedAt().String(), eDate)
	ExpectedEqual(t, mB.DeletedAt().String(), eDate)
	ExpectedEqual(t, mB.FieldA, fieldA)
	ExpectedEqual(t, mB.FieldB, fieldB)

	ExpectedEqual(t, iB, mB)

	mC := &TestModelStruct{}

	_, _ = Embed(mC, c)

	ExpectedNotEqual(t, mC.Id(), mB.Id())
	ExpectedNotEqual(t, mC.CreatedAt(), mB.CreatedAt())
	ExpectedNotEqual(t, mC.UpdatedAt(), mB.UpdatedAt())
	ExpectedNotEqual(t, mC.DeletedAt(), mB.DeletedAt())
	ExpectedNotEqual(t, mC.FieldA, mB.FieldA)
	ExpectedNotEqual(t, mC.FieldB, mB.FieldB)

	bB, err := mB.Marshal()

	_, err = Unmarshal(bB, mC, c)

	ExpectedError(t, err, "failed to embed model: invalid model type: model.TestModelStruct: it can not embed new Model: field already initialized")
}

func TestTimestamps_BackupTimestamps(t *testing.T) {
	timestamps := timestamps{
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now(),
		DeletedAt: time.Now().Add(time.Hour),
	}

	backup := timestamps.BackupTimestamps()

	ExpectedEqual(t, backup, timestamps)

	timestamps.CreatedAt = timestamps.CreatedAt.Add(-1 * time.Hour)

	ExpectedNotEqual(t, backup, timestamps)
}

func TestTimestamps_RestoreTimestamps(t *testing.T) {
	timestamps := timestamps{
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now(),
		DeletedAt: time.Now().Add(time.Hour),
	}

	backup := timestamps.BackupTimestamps()

	ExpectedEqual(t, backup, timestamps)

	timestamps.CreatedAt = timestamps.CreatedAt.Add(-1 * time.Hour)

	ExpectedNotEqual(t, backup, timestamps)

	timestamps.RestoreTimestamps(backup)

	ExpectedEqual(t, timestamps, backup)

	backup.CreatedAt = timestamps.CreatedAt.Add(-1 * time.Hour)

	ExpectedNotEqual(t, backup, timestamps)
}