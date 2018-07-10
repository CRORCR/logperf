package loganaly

import (
	"bufio"
	"fmt"
	"io"
	"logperf/stack"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type LogAnaly struct {
	LogMap                                          map[string]string
	NodeId                                          string            //控制四种请求必须来自同一个对象
	NodeIDMap                                       map[string]string //id和addr映射
	SaveFileFlag                                    map[string]string
	NeighborsBool, FindNodeBool, PongBool, PingBool map[string]bool
	MStack,AddrStack                                *stack.Stack
	lock sync.Mutex
}

var logAnaly *LogAnaly

func NewAnaly()*LogAnaly {
	logAnaly = &LogAnaly{LogMap: make(map[string]string), NodeIDMap: make(map[string]string), SaveFileFlag: make(map[string]string),
		NeighborsBool: make(map[string]bool), FindNodeBool: make(map[string]bool), PongBool: make(map[string]bool), PingBool: make(map[string]bool),
		MStack: new(stack.Stack), AddrStack: new(stack.Stack),
	}
	return logAnaly
}

func Common(fileName string,writeFileName string) {
	logAnaly=NewAnaly()
	logAnaly.lock.Lock()
	defer logAnaly.lock.Unlock()
	readFile(fileName)
	pushAddrStack()
	pushStack(writeFileName)

}

//读取数据存入 栈
func readFile(fileName string) {
	file, err := os.Open(fileName)
	defer file.Close()
	if err != nil {
		fmt.Println("open file failed")
		return
	}
	reader := bufio.NewReader(file)
	fmt.Println(logAnaly.MStack.IsEmpty())
	for {
		s, errRet := reader.ReadString('\n')
		logAnaly.MStack.Push(s)
		logAnaly.AddrStack.Push(s)
		if errRet != nil {
			if errRet == io.EOF {
				errRet = nil
				return
			}
			panic("read file failed")
		}
	}
}

func pushAddrStack() {
	for {
		value, err := logAnaly.AddrStack.Pop()
		if err != nil {
			return
		}
		text, _ := value.(string)
		//所有id和addr存储 nodeIDMap   id=f191c725fae43cc4 addr=192.168.3.100:30303
		if strings.Contains(text, "Skipping dial candidate") {
			pushNodeIDMap(text)
			continue
		}
		//获得peers
		if strings.Contains(text, "Adding p2p peer") {
			addPeer(text)
			continue
		}
	}
}

func pushStack(writeFileName string) {
	for {
		value, err := logAnaly.MStack.Pop()
		if err != nil {
			return
		}
		text, _ := value.(string)

		if !strings.Contains(text, "GotReply") &&
			!strings.Contains(text, "WriteFindNodeTime") &&
			!strings.Contains(text, "WritePingTime") {
			continue
		}
		if checkNeighbors(text) {
			continue
		}
		if checkFindNode(text) {
			continue
		}
		if checkPong(text) {
			continue
		}
		if checkPing(text, writeFileName) {
			continue
		}
	}
}

func pushNodeIDMap(text string) {
	writePingTime := strings.Split(text, " ")
	var key, value string //192.168.3.109:30303=14055bc410ec7242
	for _, v := range writePingTime {
		if "" == v {
			continue
		}
		if strings.Contains(v, "id=") {
			split := strings.Split(v, "=")
			value = strings.TrimSpace(split[1])
		} else if strings.Contains(v, "addr") {
			split := strings.Split(v, "=")
			i := strings.Split(split[1], ":")
			key = strings.TrimSpace(i[0])
		}
		if _, ok := logAnaly.NodeIDMap[key]; !ok {
			logAnaly.NodeIDMap[key] = value
		}
	}
}

func saveFile(logmap map[string]string, writeFileName string) {
	var pingTime, pongTime, findNodeTime, neighborsTime int
	var pingdata, pongdata, findnodedate, neighborsdata string
	//写文件未处理失败
	file, _ := os.OpenFile(writeFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	defer file.Close()
	var buff = bufio.NewWriter(file)
	//排序 加 设置时间
	var keys []string
	for k := range logmap {
		if strings.Contains(k, "WritePingTime") {
			timeStr := strings.Split(logmap[k], "NowTime:")
			dataStr := strings.Split(timeStr[0], "[")
			str := splitN(timeStr[1])
			logmap[k] = dataStr[0]
			pingdata = fmt.Sprintf("[%s|%s]", dataStr[1][:len(dataStr[1])-1], str[7:])
			pingTime, _ = strconv.Atoi(str)
		} else if strings.Contains(k, "PongReply") {
			timeStr := strings.Split(logmap[k], "NowTime:")
			dataStr := strings.Split(timeStr[0], "[")
			str := splitN(timeStr[1])
			logmap[k] = dataStr[0]
			pongdata = fmt.Sprintf("[%s|%s]", dataStr[1][:len(dataStr[1])-1], str[7:])
			pongTime, _ = strconv.Atoi(str)
		} else if strings.Contains(k, "FindNode") {
			timeStr := strings.Split(logmap[k], "NowTime:")
			dataStr := strings.Split(timeStr[0], "[")
			str := splitN(timeStr[1])
			logmap[k] = dataStr[0]
			findnodedate = fmt.Sprintf("[%s|%s]", dataStr[1][:len(dataStr[1])-1], str[7:])
			findNodeTime, _ = strconv.Atoi(str)
		} else if strings.Contains(k, "Bors") {
			timeStr := strings.Split(logmap[k], "NowTime:")
			dataStr := strings.Split(timeStr[0], "[")
			str := splitN(timeStr[1])
			logmap[k] = dataStr[0]
			neighborsdata = fmt.Sprintf("[%s|%s]", dataStr[1][:len(dataStr[1])-1], str[7:])
			neighborsTime, _ = strconv.Atoi(str)
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var f string
	buff.Write([]byte("\n"))

	for _, k := range keys {
		if strings.Contains(k, "WritePingTime") {
			logmap[k] = fmt.Sprintf("%s ping		:%d	毫秒", splitTime(logmap[k]), pingTime)
			//头部设置毫秒值
			f = fmt.Sprintf("%s%s %s\n", pingdata, k, logmap[k])
		} else if strings.Contains(k, "PongReply") {
			logmap[k] = fmt.Sprintf("%s pong		:%d	  毫秒", splitTime(logmap[k]), pongTime-pingTime)
			f = fmt.Sprintf("%s%s %s\n", pongdata, k, logmap[k])
		} else if strings.Contains(k, "FindNode") {
			logmap[k] = fmt.Sprintf("%s findnode	:%d 毫秒", splitTime(logmap[k]), findNodeTime-pongTime)
			f = fmt.Sprintf("%s%s %s\n", findnodedate, k, logmap[k])
		} else if strings.Contains(k, "Bors") {
			logmap[k] = fmt.Sprintf("%s neighbors	:%d	  毫秒", splitTime(logmap[k]), neighborsTime-findNodeTime)
			f = fmt.Sprintf("%s%s %s\n", neighborsdata, k, logmap[k])
		} else {
			f = fmt.Sprintf("%s %s\n", k, logmap[k])
		}

		buff.Write([]byte(f))
	}
	//json格式输出
	//bytes, _ := json.Marshal(logMap)
	buff.Flush()
	file.Close()
	return
}

func splitTime(str string) string {
	return strings.Trim(str, "\n")
}
func splitN(str string) string {
	if strings.Contains(str, "\n") {
		str = strings.Trim(str, "\n")
	}
	if strings.Contains(str, "\r") {
		str = strings.Trim(str, "\r")
	}
	if strings.Contains(str, "\t") {
		str = strings.Trim(str, "\t")
	}
	return str
}

func checkPing(text, writeFileName string) bool {
	if strings.Contains(text, "WritePingTime") {
		writePingTime := strings.Split(text, " ")
		var to, NowTime string
		for _, v := range writePingTime {
			if "" == v {
				continue
			}
			if strings.Contains(v, "=") {
				split := strings.Split(v, "=")
				if "TO" == split[0] {
					to = strings.Split(split[1], ":")[0]
					//node 是否存在
					if !checkNodeIdMap(to) {
						return true
					}
				} else if "NowTime" == split[0] {
					NowTime = splitN(split[1])
				}
			}
		}
		if logAnaly.SaveFileFlag[to] != "peers Neighbors FindNode PongReply" {
			return true
		}

		logAnaly.SaveFileFlag[to] = logAnaly.SaveFileFlag[to] + " WritePingTime"
		var k = to[len(to)-4:] + "WritePingTime"
		logAnaly.LogMap[k] = fmt.Sprintf("\t\tTO:  %s\t\t%sNowTime:%s\t", to, writePingTime[1], NowTime)

		//id提取,存储,删除
		sameNodeIDsaveFile := make(map[string]string)
		for k, v := range logAnaly.LogMap {
			if strings.Contains(v, to) {
				sameNodeIDsaveFile[k] = v
			}
		}
		if len(sameNodeIDsaveFile) != 4 {
			for k, v := range logAnaly.LogMap {
				if strings.Contains(v, to) {
					delete(logAnaly.LogMap, k)
				}
			}
			return true
		}
		/*fmt.Println("nodeIDMap",nodeIDMap)
		fmt.Println("logMap",logMap)*/
		for k, v := range logAnaly.NodeIDMap {
			if "" == k || !strings.Contains(v, to) {
				continue
			}
			for k2, v2 := range logAnaly.LogMap { //k:addr
				if strings.Contains(k2, k) {
					sameNodeIDsaveFile[k2] = v2
					break
				}
			}
		}

		saveFile(sameNodeIDsaveFile, writeFileName)
		//输出对应peers id和addr映射

		for k, v := range logAnaly.LogMap {
			if strings.Contains(v, to) {
				delete(logAnaly.LogMap, k)
			}
		}
	}
	return false
}

func checkPong(text string) bool {
	if strings.Contains(text, "GotReply") && strings.Contains(text, "ptype=2") {
		gotReplyList := strings.Split(text, " ")
		var from, NowTime string
		for _, v := range gotReplyList {
			if "" == v {
				continue
			}
			if strings.Contains(v, "=") {
				split := strings.Split(v, "=")
				if "FROM" == split[0] {
					from = strings.Split(split[1], ":")[0]
					if !checkNodeIdMap(from) {
						return true
					}
				} else if "NowTime" == split[0] {
					NowTime = splitN(split[1])
				}
			}
		}
		if !strings.Contains(logAnaly.SaveFileFlag[from], "peers Neighbors FindNode") || strings.Contains(logAnaly.SaveFileFlag[from], "WritePingTime") {
			return true
		} else {
			if !logAnaly.PongBool[from] {
				logAnaly.SaveFileFlag[from] = logAnaly.SaveFileFlag[from] + " PongReply"
				k := from[len(from)-4:] + "PongReply"
				logAnaly.PongBool[from] = true
				logAnaly.LogMap[k] = fmt.Sprintf("\tptype:2\tfrom:%s\t\t%sNowTime:%s\t", from, gotReplyList[1], NowTime)
			} else {
				for k, v := range logAnaly.LogMap {
					if strings.Contains(k, "PongReply") && strings.Contains(v, from) {
						logAnaly.LogMap[k] = fmt.Sprintf("\tptype:2\tfrom:%s\t\t%sNowTime:%s\t", from, gotReplyList[1], NowTime)
						return true
					}
				}
			}
			return true
		}
	}
	return false
}

func checkFindNode(text string) bool {
	if strings.Contains(text, "WriteFindNodeTime") {
		findNodeList := strings.Split(text, " ")
		var to, NowTime string
		for _, v := range findNodeList {
			if "" == v {
				continue
			}
			if strings.Contains(v, "=") {
				split := strings.Split(v, "=")
				if "TO" == split[0] {
					to = strings.Split(split[1], ":")[0]
					if !checkNodeIdMap(to) {
						return true
					}
				} else if "NowTime" == split[0] {
					NowTime = splitN(split[1])
				}
			}
		}

		if !strings.Contains(logAnaly.SaveFileFlag[to], "peers Neighbors") || strings.Contains(logAnaly.SaveFileFlag[to], "PongReply") {
			return true
		} else {
			if !logAnaly.FindNodeBool[to] {
				logAnaly.SaveFileFlag[to] = logAnaly.SaveFileFlag[to] + " FindNode"
				k := to[len(to)-4:] + "FindNode"
				logAnaly.FindNodeBool[to] = true
				logAnaly.LogMap[k] = fmt.Sprintf("\t\t\tTO:  %s\t\t%sNowTime:%s	", to, findNodeList[1], NowTime)
			} else {
				for k, v := range logAnaly.LogMap {
					if strings.Contains(k, "FindNode") && strings.Contains(v, to) {
						logAnaly.LogMap[k] = fmt.Sprintf("\t\t\tTO:  %s\t\t%sNowTime:%s	", to, findNodeList[1], NowTime)
						return true
					}
				}
			}
			return true
		}
	}
	return false
}

func checkNeighbors(text string) bool {
	if strings.Contains(text, "GotReply") && strings.Contains(text, "ptype=4") {
		gotReplyList := strings.Split(text, " ")
		var from, NowTime string
		for _, v := range gotReplyList {
			if "" == v {
				continue
			}
			if strings.Contains(v, "=") {
				split := strings.Split(v, "=")
				if "FROM" == split[0] {
					from = strings.Split(split[1], ":")[0]
					if !checkNodeIdMap(from) {
						return true
					}
				} else if "NowTime" == split[0] {
					NowTime = splitN(split[1])
				}
			}
		}

		if !strings.Contains(logAnaly.SaveFileFlag[from], "peers") || strings.Contains(logAnaly.SaveFileFlag[from], "FindNode") {
			return true
		} else {
			if !logAnaly.NeighborsBool[from] {
				logAnaly.SaveFileFlag[from] = logAnaly.SaveFileFlag[from] + " Neighbors"
				k := from[len(from)-4:] + "Bors"
				logAnaly.NeighborsBool[from] = true
				logAnaly.LogMap[k] = fmt.Sprintf("\t\tptype:4\tfrom:%s\t\t%sNowTime:%s\t", from, gotReplyList[1], NowTime)
			} else {
				for k, v := range logAnaly.LogMap {
					if strings.Contains(k, "Neighbors") && strings.Contains(v, from) {
						logAnaly.LogMap[k] = fmt.Sprintf("\t\tptype:4\tfrom:%s\t\t%sNowTime:%s\t", from, gotReplyList[1], NowTime)
						return true
					}
				}
			}
			return true
		}
	}
	return false
}

//对应的ip 匹配四种请求 顺便初始化map
//避免重复寻找ip,这里用一个flag 一次循环只会找一个ip
/*func addIP(text string) (ip string) {
	peerFind=false
	defer func() { gotReply4 = true }()
	split := strings.Split(text, " ")
	for _, v := range split {
		if strings.Contains(v, "addr") {
			ipAndPort := strings.Split(v, "=")
			ip = strings.TrimSpace(ipAndPort[1])
			logMap["1ip"] = ip
		}
	}
	return
}*/

func addPeer(text string) {
	split := strings.Split(text, " ")
	var addr, peerCount string
	for _, v := range split {
		peerInfo := strings.Split(v, "=")
		if strings.Contains(peerInfo[0], "addr") {
			addr = peerInfo[1]
			//ip对应ip
			n := splitN(addr)
			i := strings.Split(n, ":")
			if v, ok := logAnaly.NodeIDMap[i[0]]; !ok {
				return
			} else {
				logAnaly.SaveFileFlag[v] = "peers"
			}
		} else if strings.Contains(peerInfo[0], "peers") {
			peerCount = peerInfo[1]
		}
	}

	peerCount = splitN(peerCount)
	logAnaly.LogMap[addr] = fmt.Sprintf("%s", peerCount)
	return
}

func checkNodeIdMap(id string) (flag bool) {
	for _, v := range logAnaly.NodeIDMap {
		if id == v {
			return true
		}
	}
	return false
}
