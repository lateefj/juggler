package main

import (
	"io"
	"fmt"
	"time"
	"flag"
	"io/ioutil"
	"encoding/json"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	"github.com/lateefj/juggler"
)

const (
	FULL_FILE_SIZE    = int64(104857600)
	NUM_PARTIAL_FILES = 10 // Probably depends on the file size ...
	CONTENT_TYPE      = "text/text"
	TEST_FILE_NAME    = "juggler/test.txt"
	RUNS              = 3 // Need to be like 5 - 10 or something
)

type Conf struct {
	ACCESS_KEY  string `json:"ACCESS_KEY"`
	SECRET_KEY  string `json:"SECRET_KEY"`
	TEST_BUCKET string `json:"TEST_BUCKET"`
}

func loadConfig() (*Conf, error) {
	cpath := flag.String("c", "aws.json", "JSON configuartion file with aws credentials")
	flag.Parse()
	f, err := ioutil.ReadFile(*cpath)
	if err != nil {
		return nil, err
	}
	c := &Conf{}
	err = json.Unmarshal(f, c)
	if err != nil {
		return nil, err
	}
	return c, nil

}

type FakeReader struct {
	index int64
	size  int64
}

// Just write whatever number of bytes setup to write
func (fr *FakeReader) Read(p []byte) (int, error) {
	x := []byte("X")[0]
	if fr.index >= fr.size {
		return 0, io.EOF
	}
	amount := fr.size - fr.index
	if int64(len(p)) < amount {
		amount = int64(len(p))
	}
	i := int64(0)
	for ; i < amount; i++ {
		p[i] = x
	}
	fr.index = fr.index + amount
	return int(i), nil
}

func NewReader(size int64) *FakeReader {
	return &FakeReader{0, size}
}

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

func singleWrite(n string, b *s3.Bucket) {
	s := newS3File(n, FULL_FILE_SIZE, b)
	storeFile(s)
}
func concurrentWrite(n string, b *s3.Bucket, count int) {
	o := juggler.NewO()
	size := FULL_FILE_SIZE / NUM_PARTIAL_FILES
	for i := 0; i < NUM_PARTIAL_FILES; i++ {
		s := newS3File(fmt.Sprintf("%s-partial-%d", n, i), size, b)
		o.AddPF(storeFile, s)
	}
	i := 0
	for _ = range o.Range() {
		i++
	}
}

func singleReader(n string, b *s3.Bucket) {
	s := newS3File(n, FULL_FILE_SIZE, b)
	getFile(s)
}

func concurrentReader(n string, b *s3.Bucket, count int) {
	o := juggler.NewO()
	size := FULL_FILE_SIZE / NUM_PARTIAL_FILES
	for i := 0; i < NUM_PARTIAL_FILES; i++ {
		s := newS3File(fmt.Sprintf("%s-partial-%d", n, i), size, b)
		o.AddPF(getFile, s)
	}
	i := 0
	for _ = range o.Range() {
		i++
	}
}

func avg(times []int) float32 {
	tt := 0
	for _, t := range times {
		tt = tt + t
	}
	return float32(tt) / float32(len(times))
}
func main() {
	conf, err := loadConfig()
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
	sTimes := make([]int, RUNS)
	for i := 0; i < RUNS; i++ {
		s := time.Now()
		println("Ok writing file")
		singleWrite(TEST_FILE_NAME, b)
		e := time.Now().Second() - s.Second()
		sTimes = append(sTimes, e)
	}
	a := avg(sTimes)
	fmt.Printf("Average time to single upload is %f\n", a)
	ctTimes := make([]int, RUNS)
	for i := 0; i < RUNS; i++ {
		s := time.Now()
		println("Ok concurrent writing file")
		concurrentWrite(TEST_FILE_NAME, b, 10)
		e := time.Now().Second() - s.Second()
		ctTimes = append(ctTimes, e)
	}
	a = avg(ctTimes)
	fmt.Printf("Average time to 10 concurrent upload is %f\n", a)
	sTimes = make([]int, RUNS)
	for i := 0; i < RUNS; i++ {
		s := time.Now()
		println("Ok writing file")
		singleReader(TEST_FILE_NAME, b)
		e := time.Now().Second() - s.Second()
		sTimes = append(sTimes, e)
	}
	a = avg(sTimes)
	fmt.Printf("Average time to single reading is %f\n", a)
	ctTimes = make([]int, RUNS)
	for i := 0; i < RUNS; i++ {
		s := time.Now()
		println("Ok concurrent writing file")
		concurrentReader(TEST_FILE_NAME, b, 10)
		e := time.Now().Second() - s.Second()
		ctTimes = append(ctTimes, e)
	}
	a = avg(ctTimes)
	fmt.Printf("Average time to 10 concurrent reading is %f\n", a)

}
