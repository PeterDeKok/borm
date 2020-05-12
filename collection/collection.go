package collection

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"path/filepath"
	"peterdekok.nl/gotools/borm/model"
	"peterdekok.nl/gotools/logger"
	"reflect"
	"strings"
	"sync"
	"time"
)

type Options struct {
	file      string
	dbTimeout time.Duration
}

type Collections struct {
	c map[string]*Collection

	name string
	db   *bolt.DB
	log  *logrus.Entry

	sync.RWMutex
}

type Collection struct {
	m  map[uuid.UUID]model.Interface
	mt reflect.Type

	name string
	log  *logrus.Entry
	root *Collections

	sync.RWMutex
}

var (
	log logger.Logger
	DefaultOptions = &Options{
		file:      "models.db",
		dbTimeout: 50 * time.Millisecond,
	}
)

func init() {
	log = logger.New("borm.collection")
}

func (opt *Options) complete() *Options {
	if reflect.ValueOf(opt.file).IsZero() {
		opt.file = DefaultOptions.file
	}

	// Adding 50 ms, due to flock retry delay being subtracted from timeout
	opt.dbTimeout += DefaultOptions.dbTimeout

	return opt
}

func (opt *Options) String() string {
	strs := make([]string, 0, 1)

	strs = append(strs, fmt.Sprintf(" file-path: %s", opt.file))
	strs = append(strs, fmt.Sprintf("db-timeout: %s", opt.dbTimeout))

	return strings.Join(strs, ", ")
}

func Init(options *Options) *Collections {
	// Note: We don't need to worry about using the same file twice.
	// BBolt will lock the file once in use.
	if options == nil {
		options = &Options{}
	}

	options.complete()

	dbName := strings.TrimSuffix(filepath.Base(options.file), filepath.Ext(options.file))

	l := log.WithField("db", dbName).WithField("options", options)

	l.Debug("Initializing collection")

	// Open the collection data file from the options.
	// The default location is the `models.db` file in the current working directory
	db, err := bolt.Open(options.file, 0600, &bolt.Options{
		Timeout:      options.dbTimeout,
		NoGrowSync:   false,
		FreelistType: bolt.FreelistArrayType,
	})

	if err != nil {
		l.WithError(err).Error("Failed to open database")

		panic(err)
	}

	l.Debug("Collection initialized")

	return &Collections{
		c: make(map[string]*Collection),

		name: dbName,
		db:   db,
		log:  l,
	}
}

func (cs *Collections) Register(mi model.Interface) (model.CollectionInterface, error) {
	iv, _, err := model.CheckInterface(mi)

	if err != nil {
		cs.log.WithError(err).Error("Failed to register model")

		return nil, fmt.Errorf("failed to register model: %s", err)
	}

	name := iv.Type().Name()

	l := cs.log.WithField("collection", name)

	cs.Lock()
	defer cs.Unlock()

	if _, ok := cs.c[name]; ok {
		err := errors.New("duplicate name")

		l.WithError(err).Error("Failed to register model")

		return nil, fmt.Errorf("failed to register model: %s", err)
	}

	c := &Collection{
		m:  make(map[uuid.UUID]model.Interface),
		mt: iv.Type(),

		name: name,
		log:  l,
		root: cs,
	}

	if err := c.load(); err != nil {
		l.WithError(err).Error("Failed to register model")

		return nil, fmt.Errorf("failed to register model: %s", err)
	}

	cs.c[name] = c

	return c, nil
}

func (cs *Collections) Get(name string) (model.CollectionInterface, error) {
	cs.RLock()
	defer cs.RUnlock()

	if c, ok := cs.c[name]; ok {
		return c, nil
	}

	return nil, fmt.Errorf("collection %s not found", name)
}

func (cs *Collections) Close() error {
	cs.Lock()
	defer cs.Unlock()

	return cs.db.Close()
}

func (c *Collection) Load() error {
	c.Lock()
	defer c.Unlock()

	return c.load()
}

func (c *Collection) load() error {
	return c.root.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.name))

		if b == nil {
			return nil
		}

		return b.ForEach(func(k, v []byte) error {
			if v == nil {
				return nil
			}

			var nmi model.Interface

			nmiV := reflect.ValueOf(&nmi)
			nmiV.Elem().Set(reflect.New(c.mt))

			if _, err := model.Unmarshal(v, nmi, c); err != nil {
				return err
			}

			c.m[nmi.Id()] = nmi

			return nil
		})
	})
}

func (c *Collection) Create(i model.Interface) error {
	if _, err := model.Embed(i, c); err != nil {
		return err
	}

	return i.Save()
}

func (c *Collection) Save(i model.Interface) error {
	ic := i.Collection()

	if c != ic {
		err := errors.New("save called with model of other collection")

		c.log.WithError(err).Error("Failed to save model")

		return fmt.Errorf("failed to save model: %s", err)
	}

	v, err := i.Marshal()

	if err != nil {
		c.log.WithError(err).Error("Failed to save model")

		return fmt.Errorf("failed to save model: %s", err)
	}

	c.Lock()
	defer c.Unlock()

	if ei, exists := c.m[i.Id()]; exists {
		if ei != i {
			err := errors.New("duplicate model")

			c.log.WithError(err).Error("Failed to save model")

			return fmt.Errorf("failed to save model: %s", err)
		}
	}

	err = c.root.db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(c.name))

		if err := b.Put([]byte(i.Id().String()), v); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.log.WithError(err).Error("Failed to save model")

		return fmt.Errorf("failed to save model: %s", err)
	}

	c.m[i.Id()] = i

	return nil
}
