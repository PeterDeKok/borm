package collection

import (
	"errors"
	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
	"peterdekok.nl/gotools/borm/model"
	. "peterdekok.nl/gotools/test"
	"testing"
	"time"
)

type TestCollectionStructA struct{ model.Model }

type TestCollectionStructB struct {
	model.Model
	FieldA string
	FieldB int
}

type TestCollectionStaticIdA struct {
	model.Model
}

func init() {
	DefaultOptions.file = "testdata/models.db"
}

func (m *TestCollectionStaticIdA) Id() uuid.UUID {
	u, _ := uuid.Parse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	return u
}

type TestCollectionStaticIdB struct {
	model.Model
}

func (m *TestCollectionStaticIdB) Id() uuid.UUID {
	u, _ := uuid.Parse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	return u
}

type TestCollectionPtrToInt int

func (m *TestCollectionPtrToInt) Id() uuid.UUID                         { return uuid.New() }
func (m *TestCollectionPtrToInt) CreatedAt() time.Time                  { return time.Time{} }
func (m *TestCollectionPtrToInt) UpdatedAt() time.Time                  { return time.Time{} }
func (m *TestCollectionPtrToInt) DeletedAt() time.Time                  { return time.Time{} }
func (m *TestCollectionPtrToInt) Exists() bool                          { return true }
func (m *TestCollectionPtrToInt) Collection() model.CollectionInterface { return nil }
func (m *TestCollectionPtrToInt) Marshal() ([]byte, error)              { return nil, nil }
func (m *TestCollectionPtrToInt) Unmarshal(_ []byte) error              { return nil }
func (m *TestCollectionPtrToInt) Save() error                           { return errors.New("error") }
func (m *TestCollectionPtrToInt) Lock()                                 {}
func (m *TestCollectionPtrToInt) Unlock()                               {}

type TestCollectionUnmarshalError struct{ model.Model }

func (ue *TestCollectionUnmarshalError) Unmarshal(_ []byte) error {
	return errors.New("error unmarshal")
}

func (ue *TestCollectionUnmarshalError) Marshal() ([]byte, error) {
	return nil, errors.New("error marshal")
}

type TestCollectionStructNoBucket struct{ model.Model }

func TestOptions_complete(t *testing.T) {
	optA := &Options{}
	optA.complete()

	ExpectedEqual(t, optA.file, "testdata/models.db")
	ExpectedEqual(t, optA.dbTimeout, 50*time.Millisecond)

	optB := &Options{
		file: "testdata/test.db",
	}
	optB.complete()

	ExpectedEqual(t, optB.file, "testdata/test.db")
	ExpectedEqual(t, optB.dbTimeout, 50*time.Millisecond)

	optC := &Options{
		file:      "testdata/test.db",
		dbTimeout: 200 * time.Millisecond,
	}
	optC.complete()

	ExpectedEqual(t, optC.file, "testdata/test.db")
	ExpectedEqual(t, optC.dbTimeout, 250*time.Millisecond)
}

func TestOptions_String(t *testing.T) {
	optA := &Options{
		file: "testdata/test.db",
	}

	ExpectedEqual(t, optA.String(), " file-path: testdata/test.db, db-timeout: 0s")

	optA.complete()

	ExpectedEqual(t, optA.String(), " file-path: testdata/test.db, db-timeout: 50ms")
}

func TestInit(t *testing.T) {
	csA := Init(nil)

	defer func() {
		if err := csA.db.Close(); err != nil {
			t.Error("Failed to close db")

			t.Fail()
		}
	}()

	ExpectedEqual(t, csA.name, "models")
	ExpectedEqual(t, csA.db.Path(), "testdata/models.db")

	err := csA.db.Update(func(tx *bolt.Tx) error {
		if !tx.Writable() {
			t.Error("database should be writable")

			t.Fail()
		}

		return nil
	})

	ExpectedNoError(t, err)

	start := time.Now()

	defer func() {
		end := time.Now()

		err, ok := recover().(error)

		if !ok || err == nil {
			t.Error("expected Init to panic")

			t.Fail()

			return
		}

		ExpectedEqual(t, err.Error(), "timeout")

		ExpectedTime(t, start, end, 50*time.Millisecond, 4*time.Millisecond)
	}()

	Init(&Options{
		dbTimeout: 50 * time.Millisecond,
	})
}

func TestCollections_Register(t *testing.T) {
	var err error

	csA := Init(nil)

	defer func() {
		if err := csA.db.Close(); err != nil {
			t.Error("Failed to close db")

			t.Fail()
		}
	}()

	err = csA.db.Update(func(tx *bolt.Tx) error {
		c := tx.Cursor()
		for k, _ := c.Last(); k != nil; k, _ = c.Prev() {
			if err := tx.DeleteBucket(k); err != nil {
				return err
			}
		}
		return nil
	})

	ExpectedNoError(t, err)

	_, err = csA.Register(new(TestCollectionPtrToInt))

	ExpectedError(t, err, "failed to register model: invalid model type: collection.TestCollectionPtrToInt (ptr to int): expected pointer to named struct")

	_, err = csA.Register(&TestCollectionStructA{})

	ExpectedNoError(t, err)

	_, err = csA.Register(&TestCollectionStructA{})

	ExpectedError(t, err, "failed to register model: duplicate name")

	err = csA.db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte("TestCollectionUnmarshalError"))

		return b.Put([]byte("some-test-key"), []byte("{}"))
	})

	ExpectedNoError(t, err)

	_, err = csA.Register(&TestCollectionUnmarshalError{})

	ExpectedError(t, err, "failed to register model: error unmarshal")

	_, err = csA.Register(&TestCollectionStructNoBucket{})

	ExpectedNoError(t, err)

	err = csA.db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte("TestCollectionStructB"))

		_, err := b.CreateBucketIfNotExists([]byte("test"))

		return err
	})

	if err != nil {
		t.Error(err)

		t.Fail()
	}

	var c model.CollectionInterface

	c, err = csA.Register(&TestCollectionStructB{})

	ExpectedNoError(t, err)

	ExpectedEqual(t, len(csA.c["TestCollectionStructB"].m), 0)

	err = c.Create(&TestCollectionStructB{
		FieldA: "test-aaa",
		FieldB: 42,
	})

	ExpectedNoError(t, err)

	delete(csA.c, "TestCollectionStructB")

	c, err = csA.Register(&TestCollectionStructB{})

	ExpectedNoError(t, err)

	ExpectedEqual(t, len(csA.c["TestCollectionStructB"].m), 1)
}

func TestCollections_Get(t *testing.T) {
	csA := Init(nil)

	defer func() {
		if err := csA.db.Close(); err != nil {
			t.Error("Failed to close db")

			t.Fail()
		}
	}()

	_, err := csA.Get("TestCollectionStructA")

	ExpectedError(t, err, "collection TestCollectionStructA not found")

	cA, err := csA.Register(&TestCollectionStructA{})

	ExpectedNoError(t, err)

	cB, err := csA.Get("TestCollectionStructA")

	ExpectedNoError(t, err)

	if cA != cB {
		t.Error("Get collection reference should be equal to registered collection")

		t.Fail()
	}
}

func TestCollection_Load(t *testing.T) {
	csA := Init(nil)

	defer func() {
		if err := csA.db.Close(); err != nil {
			t.Error("Failed to close db")

			t.Fail()
		}
	}()

	err := csA.db.Update(func(tx *bolt.Tx) error {
		c := tx.Cursor()
		for k, _ := c.Last(); k != nil; k, _ = c.Prev() {
			if err := tx.DeleteBucket(k); err != nil {
				return err
			}
		}
		return nil
	})

	ExpectedNoError(t, err)

	c, err := csA.Register(&TestCollectionStructB{})

	ExpectedNoError(t, err)

	err = c.Create(&TestCollectionStructB{
		FieldA: "test-aaa",
		FieldB: 42,
	})

	ExpectedNoError(t, err)

	cA, ok := c.(*Collection)

	if ok {
		cA.m = make(map[uuid.UUID]model.Interface)
	}

	ExpectedEqual(t, len(cA.m), 0)

	err = c.Load()

	ExpectedNoError(t, err)

	ExpectedEqual(t, len(cA.m), 1)
}

func TestCollection_Create(t *testing.T) {
	csA := Init(nil)

	defer func() {
		if err := csA.db.Close(); err != nil {
			t.Error("Failed to close db")

			t.Fail()
		}
	}()

	c, err := csA.Register(&TestCollectionStructA{})

	ExpectedNoError(t, err)

	err = c.Create(new(TestCollectionPtrToInt))

	ExpectedError(t, err, "failed to embed model: invalid model type: collection.TestCollectionPtrToInt (ptr to int): expected pointer to named struct")

	err = c.Create(&TestCollectionStructA{})

	ExpectedNoError(t, err)
}

func TestCollection_Save(t *testing.T) {
	cs := Init(nil)

	defer func() {
		if err := cs.db.Close(); err != nil {
			t.Error("Failed to close db")

			t.Fail()
		}
	}()

	err := cs.db.Update(func(tx *bolt.Tx) error {
		c := tx.Cursor()
		for k, _ := c.Last(); k != nil; k, _ = c.Prev() {
			if err := tx.DeleteBucket(k); err != nil {
				return err
			}
		}
		return nil
	})

	ExpectedNoError(t, err)

	cA, err := cs.Register(&TestCollectionUnmarshalError{})

	ExpectedNoError(t, err)

	err = cA.Save(new(TestCollectionPtrToInt))

	ExpectedError(t, err, "failed to save model: save called with model of other collection")

	m, err := model.Embed(&TestCollectionUnmarshalError{}, cA)

	ExpectedNoError(t, err)

	err = cA.Save(m)

	ExpectedError(t, err, "failed to save model: error marshal")

	cB, err := cs.Register(&TestCollectionStaticIdA{})

	ExpectedNoError(t, err)

	m, err = model.Embed(&TestCollectionStaticIdA{}, cB)

	ExpectedNoError(t, err)

	err = cB.Save(m)

	ExpectedNoError(t, err)

	m, err = model.Embed(&TestCollectionStaticIdA{}, cB)

	ExpectedNoError(t, err)

	err = cB.Save(m)

	ExpectedError(t, err, "failed to save model: duplicate model")

	cC, err := cs.Register(&TestCollectionStaticIdB{})

	ExpectedNoError(t, err)

	err = cs.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(cC.(*Collection).name))

		ExpectedNoError(t, err)

		if b == nil {
			t.Error("expected bucket")

			t.Fail()

			return errors.New("error")
		}

		_, err = b.CreateBucketIfNotExists([]byte((&TestCollectionStaticIdB{}).Id().String()))

		return err
	})

	ExpectedNoError(t, err)

	m, err = model.Embed(&TestCollectionStaticIdB{}, cC)

	ExpectedNoError(t, err)

	err = cC.Save(m)

	ExpectedError(t, err, "failed to save model: incompatible value")
}

func TestCollections_Close(t *testing.T) {
	cs := Init(nil)

	err := cs.db.View(func(tx *bolt.Tx) error {
		return nil
	})

	ExpectedNoError(t, err)

	err = cs.Close()

	ExpectedNoError(t, err)

	_, err = cs.db.Begin(false)

	ExpectedError(t, err, "database not open")
}
