package spc

import (
	"testing"
	"time"
)

func TestStart(t *testing.T){
	var err error
	var errchannel=make(chan error)
	go func() {
		err:=Start()
		errchannel<-err
	}()
	select {
	case err1 := <-errchannel:
		err = err1
	case <-time.After(5*time.Second):
		err=nil
	}
	if err==nil{
		t.Fatal("Error Should Not be Nil As No incluster config is present")
	}
}
