package main

import(
    "testing"
    "fmt"
    "sync/atomic"
)

func Test_Main(t *testing.T){
    /*dir := "d:/golang/src/russ*//*"
    listFiles(dir)*/
    fmt.Println(atomic.CompareAndSwapInt32(&isBuilding,0,1))
}
