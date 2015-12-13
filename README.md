# PChan
Priority Channel implemented for golang

Initializing:
```
pc := pchan.NewPChan()
```

Ideally you can use PChan exactly like a normal interface{} channel.  The only difference is that when you `Send` into the channel, you must also specify a priority with which to `Receive`.  Essentially, the idea is that if you sequentially call: `pc.Send(1, "first")`, `pc.Send(99, "second")`, and then call `pc.Receive()` you would receive the "second" string. This contradicts the regular channel behavior which requires a `Receive()` of the "first" string before a `Receive()` of "second".

Usage:
```
pc := pchan.NewPChan()

for i := 0; i < 10; i += 1 {
  y := i
  go func() {
    pc.Send(y, y)
  }()
}

// After all have been put into the Channel...
fmt.Printf("Yay I got: %#v", pc.Receive())
```


For a priority channel this will ALWAYS print:
```
> Yay I got: 9
```
whereas for a regular channel the print result is indeterminant!!!

#!!!WARNINGS!!!

1) Keep in mind that since it's an interface{} channel, you must cast if you want to use the result...

WORKS:
```
pc := pchan.NewPChan()
go func() {
  if pc.Receive().(int) > 2 {
    fmt.Println("Yay!")
  }
}
pc.Send(1, 3)
```

DOESN'T WORK:
```
pc := pchan.NewPChan()
go func() {
  if pc.Receive() > 2 { //Won't compile since interface{} type can't be compared to int
    fmt.Println("Yay!")
  }
}
pc.Send(1, 3)
```

2) All channel requests ARE blocking! They WILL stop execution if there's not something already attempting to `Receive` or `Send` data!  There is NO way to specify a size at the moment for this structure to prevent this.

Happy priority channeling!
