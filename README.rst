=======
Juggler
=======


Introduction
------------

These patterns originated from some ideas I had on how to improve performacne on a project that did a lot of IO in the cloud. Based on `Rob Pike's awesome presentation`__ on currency I wanted to implement these IO patterns in Go. However the project I was working on was in Python and it didn't make sense at the time to start rewrite the storage in Golang. So I wrote the codap_ inspired by Go concurrency. Now that I am writing a lot more Golang I figured it was time to write these patterns down. There is an example in examples/aws.go on how to use the library. Assuming good connection to S3 it will printout some nice numbers the difference between single file upload and a file split up into multiple files.

These are coneptual implemntation, if used they should optmized so that the type is not interface{}. If casting types all the time there will be a large performance degredation. In the context of disk or network IO this is probably inconsiquential however it is important to not that these current implementations are _slow_!

Ordered Sice Example
--------------------

.. code-block:: go

  size := 30
  o := NewO()
  count := 0
  for i := 0; i < size; i++ {
    o.AddPRF(func(data interface{}) interface{} {
      return fib(data.(int))
    }, i)
  }
  for v := range o.Range() {
    if v != fib(count) {
      t.Errorf("Expect value to be %d but was %d", fib(size), v)
    }
    count++
  }


Map Example
-----------

.. code-block:: go

  // For SetPRF
  kv := NewKV()
  for i := 0; i < fibSize; i++ {
    kv.SetPRF(i, func(s interface{}) interface{} {
      return fib(s.(int))
    }, i)
  }
  for i := 0; i < fibSize; i++ {
    v, err := kv.Get(i)
    if err != nil {
      t.Errorf("Expected to have a value for %d but got error trying to get it: %s", i, err)
    }
    if v.(int) != fib(i) {
      t.Errorf("Houston we have a problem expected %d but got %d", fib(i), v)
    }
  }



.. _presentation: https://github.com/lateefj/codap
__ presentation_
.. _codap: https://github.com/lateefj/codap


