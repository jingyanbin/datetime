package datetime

import (
	"strconv"
)

type NextNumber struct {
	s      string
	length int
	pos    int
}

func NewNextNumber(s string) *NextNumber {
	return &NextNumber{s: s, length: len(s)}
}

func (my *NextNumber) Init(s string) {
	my.s = s
	my.length = len(s)
	my.pos = 0
}

//@description: 得到字符串中的, 下一个数字
//@param:       jump int "跳跃字节数" 0:自动跳过非数字字符  >0:跳过固定宽度
//@param:       w int "数字宽度" 0:自动获取数字的宽度(遇到非数字停止) >0: 获取固定宽度的数字
//@return:      int "得到数字"
//@return:      error "错误信息"
func (my *NextNumber) Next(jump, w int) (int, bool) {
	pos := my.pos
	for ; pos < my.length && (my.s[pos] < 48 || my.s[pos] > 57); pos++ {
	}
	if pos == my.length {
		return 0, false
	}
	if jump > 0 && pos-my.pos != jump {
		return 0, false
	}
	start := pos

	if w > 0 {
		for ; pos < my.length && my.s[pos] > 47 && my.s[pos] < 58 && (pos-start) < w; pos++ {
		}
	} else {
		for ; pos < my.length && my.s[pos] > 47 && my.s[pos] < 58; pos++ {
		}
	}
	if pos == start {
		return 0, false
	}
	integer := my.s[start:pos]
	num, err := strconv.Atoi(integer)
	if err != nil {
		return 0, false
	}
	my.pos = pos
	return num, true
}

//获取字符串中所有数字
func (my *NextNumber) Numbers() []int {
	var res []int
	for {
		n, found := my.Next(0, 0)
		if !found {
			break
		}
		res = append(res, n)

	}
	return res
}
