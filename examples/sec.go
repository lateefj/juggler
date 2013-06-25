// Experimenting with concurrent encryption ....

package main

import (
	"io"
	"bytes"
	"crypto/cipher"
	"github.com/lateefj/crypter/capi"
	"github.com/lateefj/juggler"
)

func newKey() []byte {
	return capi.GenKey(capi.AES_KEY_SIZE)
}

type cryptable struct {
	block  cipher.Block
	reader io.Reader
}

func newFakeCryptable(b cipher.Block, s int64) cryptable {
	fake := NewReader(s)
	return cryptable{b, fake}
}

type cryptResponse struct {
	reader io.Reader
	err    error
}

func encReader(face interface{}) interface{} {
	c := face.(cryptable)
	cr := cryptResponse{nil, nil}
	w := bytes.NewBuffer(make([]byte, 0))
	iv := capi.GenIV()
	h := capi.NewHeader(iv)
	err := capi.WriteHeader(h, w)
	if err != nil {
		cr.err = err
		return cr
	}
	err = capi.Encrypt(c.block, h.IV, c.reader, w)
	if err != nil {
		cr.err = err
		return cr
	}
	cr.reader = w
	return cr
}

func decReader(face interface{}) interface{} {
	c := face.(cryptable)
	cr := cryptResponse{nil, nil}
	h, err := capi.ReadHeader(c.reader)
	if err != nil {
		cr.err = err
		return cr
	}
	w := bytes.NewBuffer(make([]byte, 0))
	err = capi.Decrypt(c.block, h.IV, c.reader, w)
	if err != nil {
		cr.err = err
		return cr
	}
	cr.reader = w
	return cr
}

func singleEnc(block cipher.Block, s int64) (io.Reader, error) {
	c := newFakeCryptable(block, s)
	cr := encReader(c).(cryptResponse)
	return cr.reader, cr.err
}

type ReadableEnc struct {
	o *juggler.O
}

func (re *ReadableEnc) Read(b []byte) (int, error) {
	i := 0
	s := len(b)
	for face := range re.o.Range() {
		cr := face.(cryptResponse)
		w, err := cr.reader.Read(b)
		if err != nil {
			return i, err
		}
		i = i + w
		if i == s {
			break
		}
	}
	return i, nil
}

func concurrentEnc(conf *Conf, block cipher.Block, s int64) (io.Reader, error) {

	size := int64(conf.FULL_FILE_SIZE) / int64(conf.NUM_PARTIAL_FILES)
	o := juggler.NewO()

	for i := 0; i < conf.NUM_PARTIAL_FILES; i++ {
		c := newFakeCryptable(block, size)
		o.AddPRF(encReader, c)
	}
	return &ReadableEnc{o}, nil
}
