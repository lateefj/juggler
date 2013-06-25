package main

import (
	"io"
	"io/ioutil"
	"encoding/json"
	"flag"
)

type Conf struct {
	ACCESS_KEY        string `json:"ACCESS_KEY"`
	SECRET_KEY        string `json:"SECRET_KEY"`
	TEST_BUCKET       string `json:"TEST_BUCKET"`
	FULL_FILE_SIZE    int    `json:"FULL_FILE_SIZE"`
	RUNS              int    `json:"RUNS"`
	NUM_PARTIAL_FILES int    `json:"NUM_PARTIAL_FILES"`
}

func LoadConfig() (*Conf, error) {
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
