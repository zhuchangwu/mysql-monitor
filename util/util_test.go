package util

import (
	"fmt"
	"strings"
	"testing"
)

func TestSubLastString(t *testing.T) {
	fmt.Println(SubLastString("wertyu,wertyu,qbc", ","))
}
func TestSubFirstString(t *testing.T) {
	fmt.Println(SubFirstString("wertyu,wertyu,qbc", ","))
}
func TestSubStringWithStartEnd(t *testing.T) {
	fmt.Println(SubStringWithStartEnd("helloword", 1, 5))
}

func TestSubStringBettenSub1Sub2(t *testing.T) {
	loadAvg := "09:44:36 up 198 days, 21:31,  2 users,  load average: 0.04, 0.01, 0.05"
	sub2 := SubStringBettweenSub1Sub2(loadAvg, "up", ",")
	fmt.Println(sub2)
}

/**
 *  测试获取在线人数
 */

func TestUsers(t *testing.T) {
	loadAvg := "09:44:36 up 1918 days, 21:31,  1232 users,  load average: 0.04, 0.01, 0.05"

	tempStr := SubFirstString(SubFirstString(loadAvg, ","), ",")
	index := strings.Index(tempStr, ",")
	users := SubStringWithStartEnd(tempStr, 0, index-1)
	fmt.Println(users)
}

/**
 * 按照空格切割字符串：
 * memory = "Mem:           3788        1092         161         184        2535        2269"
 */
func TestSpilitStringBySpace(t *testing.T) {
	memory := "Mem:           3788        1092         161         184        2535        2269"
	space := SpilitStringBySpace(memory)
	fmt.Println(space)
}
