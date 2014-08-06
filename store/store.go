package store

import (
	"github.com/mattrobenolt/mineshaft/aggregate"
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
}

func (s *Store) Set(p *metric.Point) error {
	start := time.Now()
	buckets := s.schema.Match(p.Path).Buckets
	agg := s.aggregation.Match(p.Path)
	defer func() {
		log.Println(p, buckets, agg, time.Now().Sub(start))
	}()
	var wg sync.WaitGroup
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
	d.Init(url)
	return d
}

var registry = make(map[string]Driver)
