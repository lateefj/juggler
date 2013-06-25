package main

import (
	"io"
	"fmt"
	"time"
	"io/ioutil"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	"github.com/lateefj/juggler"
)

const (
	CONTENT_TYPE   = "text/text"
	TEST_FILE_NAME = "juggler/test.txt"
)

type s3File struct {
	name   string
	size   int64
	bucket *s3.Bucket
	reader io.Reader
}

func newS3File(n string, s int64, b *s3.Bucket) s3File {
	fake := NewReader(s)
	return s3File{n, s, b, fake}
}

func storeFile(i interface{}) {
	f := i.(s3File)
	f.bucket.PutReader(f.name, f.reader, f.size, CONTENT_TYPE, s3.Private)
}
func getFile(i interface{}) {
	f := i.(s3File)
	r, _ := f.bucket.GetReader(f.name)
	ioutil.ReadAll(r)
}

func singleWrite(conf *Conf, n string, b *s3.Bucket) {
	s := newS3File(n, int64(conf.FULL_FILE_SIZE), b)
	storeFile(s)
}
func concurrentWrite(conf *Conf, n string, b *s3.Bucket, count int) {
	o := juggler.NewO()
	size := int64(conf.FULL_FILE_SIZE) / int64(conf.NUM_PARTIAL_FILES)
	for i := 0; i < conf.NUM_PARTIAL_FILES; i++ {
		s := newS3File(fmt.Sprintf("%s-partial-%d", n, i), size, b)
		o.AddPF(storeFile, s)
	}
	i := 0
	for _ = range o.Range() {
		i++
	}
}

func singleReader(conf *Conf, n string, b *s3.Bucket) {
	s := newS3File(n, int64(conf.FULL_FILE_SIZE), b)
	getFile(s)
}

func concurrentReader(conf *Conf, n string, b *s3.Bucket, count int) {
	o := juggler.NewO()
	size := int64(conf.FULL_FILE_SIZE) / int64(conf.NUM_PARTIAL_FILES)
	for i := 0; i < conf.NUM_PARTIAL_FILES; i++ {
		s := newS3File(fmt.Sprintf("%s-partial-%d", n, i), size, b)
		o.AddPF(getFile, s)
	}
	i := 0
	for _ = range o.Range() {
		i++
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
