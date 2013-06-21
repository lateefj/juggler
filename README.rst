=======
Juggler
=======


Introduction
------------

These patterns originated from some ideas I had on how to improve performacne on a project that did a lot of IO in the cloud. Based on `Rob Pike's awesome presentation`__ on currency I wanted to implement these IO patterns in Go. However the project I was working on was in Python and it didn't make sense at the time to start rewrite the storage in Golang. So I wrote the codap_ inspired by Go.
So in total rockstar just like U2_ bring back Helter Skelter on Rattle and Hum I am bringing back these concurrent patterns to Go! I call it Juggler because that is how I image the goroutine schedule is across processes (threads) when there are only a few coroutines. 

.. _presentation: https://github.com/lateefj/codap
__ presentation_
.. _codap: https://github.com/lateefj/codap
.. _U2: http://en.wikipedia.org/wiki/Helter_Skelter_(song)


