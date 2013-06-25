package main

import (
	"io"
	"fmt"
	"time"
	"errors"
	"io/ioutil"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	//	"github.com/lateefj/juggler" Not using generic no need !
)

const (
	CONTENT_TYPE   = "text/text"
	TEST_FILE_NAME = "juggler/test.txt"
)

type S3File struct {
	name   string
	size   int64
	bucket *s3.Bucket
	reader io.Reader
	ct     string
	acl    s3.ACL
}

func newFakeS3File(n string, s int64, b *s3.Bucket, ct string, acl s3.ACL) *S3File {
	fake := NewReader(s)
	return &S3File{n, s, b, fake, ct, acl}
}

// Basically an io.Reader implementation 
type S3Func func(data *S3File) *S3File

// Value that holds some data element and a queue when that data elemetn is ready
type S3V struct {
	Ready bool
	Queue chan *S3File
	Data  *S3File
}

func NewS3V() *S3V {
	return &S3V{false, make(chan *S3File, 1), nil}
}
// Implement a specific one for int to io.Reader as a working example
type S3KV struct {
	M map[int]*S3V // Could we use int or string keys and that would make things faster?
}

// Get 
func (kv *S3KV) Get(k int) (*S3File, error) {
	// Make sure the k is in the list of keys
	if v, ok := kv.M[k]; ok {
		// If the data is not ready then wait until it is
		if !v.Ready {
			v.Data = <-v.Queue
			v.Ready = true
		}
		return v.Data, nil
	}
	// Return an error if the key doesn't exist
	return nil, errors.New(fmt.Sprintf("Key does '%d' not exist!\n", k))
}

// Abstracts creating a value and setting it in the map
func (kv *S3KV) set(k int) *S3V {
	v := NewS3V()
	kv.M[k] = v
	return v
}

// Takes a reader and returns a reader
func (kv *S3KV) SetPRF(k int, prf S3Func, data *S3File) {
	v := kv.set(k)
	go func() {
		d := prf(data)
		v.Queue <- d
	}()
}

func NewS3KV() *S3KV {
	return &S3KV{make(map[int]*S3V)}
}

// Specific implementation for reader
type S3O struct {
	kv   *S3KV
	size int // might need to do some sync stuff here?
}

func (o *S3O) Range() chan *S3File {
	l := make(chan *S3File)
	go func() { // Push them all in order on a channel :)
		for i := 0; i < o.size; i++ {
			v, err := o.kv.Get(i)
			if err != nil { // Something really bad happens if we panic
				panic(err)
			}
			l <- v
		}
		close(l)
	}()
	return l
}

// For functions that take a parameter and return a value (maybe should be default?)
func (o *S3O) AddPRF(prf S3Func, data *S3File) {
	o.kv.SetPRF(o.size, prf, data)
	o.size++
}
func NewS3O() *S3O {
	return &S3O{NewS3KV(), 0}
}
func storeFile(f *S3File) *S3File {
	f.bucket.PutReader(f.name, f.reader, f.size, f.ct, f.acl)
	return f
}

func getFile(f *S3File) *S3File {
	r, _ := f.bucket.GetReader(f.name)
	f.reader = r
	return f
}

func singleWrite(conf *Conf, n string, b *s3.Bucket) {
	s := newFakeS3File(n, int64(conf.FULL_FILE_SIZE), b, CONTENT_TYPE, s3.Private)
	storeFile(s)
}

func concurrentWrite(conf *Conf, n string, b *s3.Bucket, count int) {
	o := NewS3O()
	size := int64(conf.FULL_FILE_SIZE) / int64(conf.NUM_PARTIAL_FILES)
	for i := 0; i < conf.NUM_PARTIAL_FILES; i++ {
		s := newFakeS3File(fmt.Sprintf("%s-partial-%d", n, i), size, b, CONTENT_TYPE, s3.Private)
		o.AddPRF(storeFile, s)
	}
	i := 0
	for _ = range o.Range() {
		i++
	}
}

func singleReader(conf *Conf, n string, b *s3.Bucket) {
	s := newFakeS3File(n, int64(conf.FULL_FILE_SIZE), b, CONTENT_TYPE, s3.Private)
	f := getFile(s)
	ioutil.ReadAll(f.reader)
}

func concurrentReader(conf *Conf, n string, b *s3.Bucket, count int) {
	o := NewS3O()
	size := int64(conf.FULL_FILE_SIZE) / int64(conf.NUM_PARTIAL_FILES)
	for i := 0; i < conf.NUM_PARTIAL_FILES; i++ {
		s := newFakeS3File(fmt.Sprintf("%s-partial-%d", n, i), size, b, CONTENT_TYPE, s3.Private)
		o.AddPRF(getFile, s)
	}
	for f := range o.Range() {
		ioutil.ReadAll(f.reader)
	}
}

func avg(times []float64) float64 {
	tt := float64(0)
	for _, t := range times {
		tt = tt + t
	}
	return tt / float64(len(times))
}
func main() {
	conf, err := LoadConfig()
	if err != nil {
		panic(err)
	}
	auth := aws.Auth{conf.ACCESS_KEY, conf.SECRET_KEY}
	if err != nil {
		panic(err)
	}
	//e := ec2.New(auth, aws.USEast)
	s := s3.New(auth, aws.USEast)
	b := s.Bucket(conf.TEST_BUCKET)
	sTimes := make([]float64, conf.RUNS)
	for i := 0; i < conf.RUNS; i++ {
		s := time.Now()
		singleWrite(conf, TEST_FILE_NAME, b)
		e := time.Now().Sub(s)
		sTimes = append(sTimes, e.Seconds())
	}
	aSingleWrite := avg(sTimes)
	fmt.Printf("Average time to single upload is %f\n", aSingleWrite)
	ctTimes := make([]float64, conf.RUNS)
	for i := 0; i < conf.RUNS; i++ {
		s := time.Now()
		concurrentWrite(conf, TEST_FILE_NAME, b, 10)
		e := time.Now().Sub(s)
		ctTimes = append(ctTimes, e.Seconds())
	}
	aConWrite := avg(ctTimes)
	fmt.Printf("Average time to 10 concurrent upload is %f\n", aConWrite)
	writeSpeedUp := aSingleWrite / aConWrite
	fmt.Printf("Concurrent speed up %f \n", writeSpeedUp)
	rsTimes := make([]float64, conf.RUNS)
	for i := 0; i < conf.RUNS; i++ {
		s := time.Now()
		singleReader(conf, TEST_FILE_NAME, b)
		e := time.Now().Sub(s)
		rsTimes = append(rsTimes, e.Seconds())
	}
	aSingleRead := avg(rsTimes)
	fmt.Printf("Average time to single reading is %f\n", aSingleRead)
	rctTimes := make([]float64, conf.RUNS)
	for i := 0; i < conf.RUNS; i++ {
		s := time.Now()
		concurrentReader(conf, TEST_FILE_NAME, b, 10)
		e := time.Now().Sub(s)
		rctTimes = append(rctTimes, e.Seconds())
	}
	aConRead := avg(rctTimes)
	fmt.Printf("Average time to 10 concurrent reading is %f\n", aConRead)
	readSpeedUp := aSingleRead / aConRead
	fmt.Printf("Concurrent speed up %f \n", readSpeedUp)

}
