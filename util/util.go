package util

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

/**
 *
 */

func SpilitStringBySpace(str string)[]string{
	temp := make([]string, 0)
	if len(str)==0 {
		return temp
	}
	split := strings.Split(str, " ")
	for _, v := range split {
		if v!="" {
			temp = append(temp,v)
		}
	}
	return temp
}

/**
 * 返回Time中的时间部分：如：21:24:11
 */
func GetTimeString(time time.Time) string{
	h := time.Hour()
	M := time.Minute()
	S := time.Second()
	return  strconv.Itoa(h) + ":" +strconv.Itoa(M) + ":"+strconv.Itoa(S)
}


/**
 * 切分字符串：返回s串中，最后出现的sub之后的部分
 * 如:  s = werty,erty,qbc
 * 返回: qbc
 */
func SubLastString(s, sub string) string {
	if len(s) == 0 || len(sub) == 0 {
		return s
	}
	strIndex := strings.LastIndex(s, sub)
	return s[len(sub)+strIndex : len(s)]
}

/**
 * 切分字符串：返回s串中，第一次出现的sub之后的部分
 * 如:  s = werty,erty,qbc
 * 返回: erty,qbc
 */
func SubFirstString(s, sub string) string {
	if len(s) == 0 || len(sub) == 0 {
		return s
	}
	strIndex := strings.Index(s, sub)
	return s[len(sub)+strIndex : len(s)]
}

/**
 * 切分字符串：返回start和end之间的字符串，包含start和end
 * 如：SubString("helloworld",1,5) 返回：ellow
 */
func SubStringWithStartEnd(s string, start, end int) string {
	if len(s) == 0 || end < 0 {
		return ""
	}
	if start < 0 {
		start = 0
	}
	if end > len(s) {
		end = len(s)
	}
	return s[start : end+1]
}

/**
 * 切分字符串：返回s串中，sub1和sub2之间的字符串
 * 例: s = hello up 4567 day,erty,rtyu.  sub1 = upd
 *     sub1 = up
 *     sub2 = ,
 *     返回： 4567 day
 */
func SubStringBettweenSub1Sub2(s, sub1, sub2 string) string {
	if len(s) == 0 {
		return s
	}
	s = SubLastString(s, sub1)
	index := strings.Index(s, sub2)
	return s[0:index]
}

/**
 *  入参：
 *  cmdStr：shell命令
 *
 *  出参：
 *  outStr：默认为标准输出
 *  status：命令行结果返回码
 *  err：错误信息
 *
 *  如使用超时命令，超时退出状态码12
 *
 *  常见错误：
 *  status:  127
 *  err:  exit status 127,/bin/bash: uptime1: command not foun
 */
func SyncExecShell(cmdStr string) (outStr string, status int, err error) {

	// 如果我们给定的name中没有路径的分隔符，Command就会使用LookUp将name解析成一个完整的路径，否则是指使用name当作路径
	// 如果name不含路径分隔符，将使用LookPath获取完整路径；否则直接使用name。参数arg不应包含命令名。
	command := exec.Command("/bin/bash", "-c", cmdStr)

	// 执行命令并返回标准输出的切片。
	// Output中使用cmd的Run函数执行命令，Run函数的特性是：阻塞直到完成
	out, err := command.Output()

	// command.ProcessState : 包含了程序退出相关的信息，当调用 Wait或者Run后可用
	// command.ProcessState.Sys() 返回和系统进程退出相关的信息，比如在unix上，这个状态就是syscall.WaitStatus
	statusCode, ok := command.ProcessState.Sys().(syscall.WaitStatus)
	if ok {
		status = statusCode.ExitStatus()
	} else if err != nil {
		// 当无法获取到命令行返回值时，如果err不为空将返回值置为-1
		status = -1
	}
	if err != nil {
		// 当执行命令异常时，尝试在error上append错误输出
		if ee, ok := err.(*exec.ExitError); ok {
			return string(out), status, errors.New(fmt.Sprintf("%s,%s", err.Error(), string(ee.Stderr)))
		}
		return string(out), status, err
	}
	return string(out), status, nil
}
