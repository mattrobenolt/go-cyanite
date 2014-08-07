package store

import (
	"github.com/mattrobenolt/mineshaft/aggregate"
	"github.com/mattrobenolt/mineshaft/index"
	"github.com/mattrobenolt/mineshaft/metric"
	"github.com/mattrobenolt/mineshaft/schema"

	"log"
	"net/url"
	"sync"
	"time"
)

type Store struct {
	driver      Driver
	schema      *schema.Schema
	aggregation *aggregate.Aggregation
	index       *index.Store
}

func (s *Store) Set(p *metric.Point) error {
	var wg sync.WaitGroup

	start := time.Now()
	buckets := s.schema.Match(p.Path).Buckets
	agg := s.aggregation.Match(p.Path)

	// Log the response time
	defer func() {
		log.Println(p, buckets, agg, time.Now().Sub(start))
	}()

	go func() {
		wg.Add(1)
		s.index.Update(p.Path)
		wg.Done()
	}()
	for _, bucket := range buckets {
		wg.Add(1)
		go func(bucket *schema.Bucket) {
			err := s.driver.WriteToBucket(p, agg, bucket)
			if err != nil {
				log.Println(err)
			}
			wg.Done()
		}(bucket)
	}

	wg.Wait()
	return nil
}

func (s *Store) Close() {
	if s.driver != nil {
		s.driver.Close()
	}
}

func (s *Store) SetDriver(driver Driver) {
	s.driver = driver
}

func (s *Store) SetSchema(schema *schema.Schema) {
	s.schema = schema
}

func (s *Store) SetAggregation(agg *aggregate.Aggregation) {
	s.aggregation = agg
}

func (s *Store) SetIndexer(index *index.Store) {
	s.index = index
}

type Driver interface {
	Init(*url.URL) error
	WriteToBucket(*metric.Point, *aggregate.Rule, *schema.Bucket) error
	Close()
}

func Register(key string, d Driver) {
	registry[key] = d
}

func GetDriver(url *url.URL) Driver {
	d, ok := registry[url.Scheme]
	if !ok {
		panic("store: driver not found")
	}
	err := d.Init(url)
	if err != nil {
		panic(err)
	}
	return d
}

func NewFromConnection(url *url.URL) *Store {
	d := GetDriver(url)
	return &Store{driver: d}
}

var registry = make(map[string]Driver)
