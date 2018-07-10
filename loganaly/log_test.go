package loganaly

import (
	"os"
	"testing"
)
var gopath = os.Getenv("GOPATH")
//var fileName=gopath+"/src/logperf/conf/common/log123.txt"//主节点
var writeFileName="G:/result.txt"


func TestCommon(t *testing.T){
	path:=gopath+"/src/logperf/conf/common"
	Common(path+"/log123.txt",writeFileName)
	Common(path+"/log100.txt",writeFileName)
	Common(path+"/log98.txt",writeFileName)
}

func TestAlply(t *testing.T){
	path:=gopath+"/src/logperf/conf/addalply"
	Common(path+"/123节点.txt",writeFileName)
	Common(path+"/19节点.txt",writeFileName)
	Common(path+"/17节点.txt",writeFileName)
}

func TestReduceTime(t *testing.T){
	path:=gopath+"/src/logperf/conf/reducetime"
	Common(path+"/123节点.txt",writeFileName)
	Common(path+"/19节点.txt",writeFileName)
	Common(path+"/17节点.txt",writeFileName)
}