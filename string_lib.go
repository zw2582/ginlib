package ginlib

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 截取字符串 start 起点下标 length 需要截取的长度
func Substr(str string, start int, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}

	return string(rs[start:end])
}

// 截取字符串 start 起点下标 end 终点下标(不包括)
func Substr2(str string, start int, end int) string {
	rs := []rune(str)
	length := len(rs)

	if start < 0 || start > length {
		panic("start is wrong")
	}

	if end < 0 || end > length {
		panic("end is wrong")
	}

	return string(rs[start:end])
}

// Md5encode md5编码
func Md5encode(src string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(src)))
}

// UniqueId 生成GUUID字串，长度24位
func UniqueId() string {
	return bson.NewObjectId().Hex()
}

func OrderNo(prefix string) string {
	rnss, _ := rand.Int(rand.Reader, new(big.Int).SetInt64(int64(9999)))
	rn := rnss.Int64()
	rns := fmt.Sprintf("%04d", rn)
	nowstr := time.Now().Format("20060102150405")
	return fmt.Sprintf("%s%s%s", prefix, nowstr, rns)
}

// 解析gbk
func DecodeGBK(s []byte) ([]byte, error) {
	I := bytes.NewReader(s)
	O := transform.NewReader(I, simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(O)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func ValidMobile(mobileNum string) bool {
	const regular = `^((\+?86)|(\(\+86\)))?((((13[^4]{1})|(14[5-9]{1})|147|(15[^4]{1})|166|(17\d{1})|(18\d{1})|(19[89]{1}))\d{8})|((134[^9]{1}|1410|1440)\d{7}))$`
	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobileNum)
}

func ValidEmail(email string) bool {
	const regular = `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	reg := regexp.MustCompile(regular)
	return reg.MatchString(email)
}

// ValidContainChinese 包含中文检测
func ValidContainChinese(str string) bool {
	const regular = `[^\x00-\x80]+`
	reg := regexp.MustCompile(regular)
	return reg.MatchString(str)
}

// ValidChineName 验证中文姓名
func ValidChineName(str string) bool {
	const regular = "^[\u4E00-\u9FA5]{2,10}$"
	reg := regexp.MustCompile(regular)
	return reg.MatchString(str)
}

func Sha1Encode(raw string) string {
	b := sha1.Sum([]byte(raw))
	return base64.StdEncoding.EncodeToString(b[:])
}

func InetAtoN(ip string) int64 {
	bits := strings.Split(ip, ".")

	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum int64

	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)

	return sum
}

func SumMu(a bool, x, y interface{}) interface{} {
	if a {
		return x
	} else {
		return y
	}
}

// InviteCode 根据id生成邀请码，id最大值为 4000000000
func InviteCode(id uint) (string, error) {
	if id > 4000000000 {
		return "", errors.New("数值太大")
	}

	base := uint(100)
	num := base + id
	baseX := fmt.Sprintf("%X", num)
	buX := ""

	lenBase := len(baseX)
	lessLen := 7 - lenBase

	if lessLen > 0 {
		min := int(math.Pow10(lessLen))
		t, _ := rand.Int(rand.Reader, big.NewInt(int64(min-1)))
		buX = fmt.Sprintf("%X", int64(min)+t.Int64())

		ru := []string{"G", "H", "I", "J", "K", "L", "M", "N"}
		buX = ru[lessLen-1] + buX
	}

	return baseX + buX, nil
}

// Compare 该函数比较两个版本号是否相等，是否大于或小于的关系
// 返回值：0表示v1与v2相等；1表示v1大于v2；2表示v1小于v2
func Compare(v1, v2 string) int {
	// 替换一些常见的版本符号
	replaceMap := map[string]string{"V": "", "v": "", "-": "."}
	//keywords := {"alpha,beta,rc,p"}
	for k, v := range replaceMap {
		if strings.Contains(v1, k) {
			strings.Replace(v1, k, v, -1)
		}
		if strings.Contains(v2, k) {
			strings.Replace(v2, k, v, -1)
		}
	}
	verStr1 := strings.Split(v1, ".")
	verStr2 := strings.Split(v2, ".")
	ver1 := strSlice2IntSlice(verStr1)
	ver2 := strSlice2IntSlice(verStr2)
	// 找出v1和v2哪一个最短
	var shorter int
	if len(ver1) > len(ver2) {
		shorter = len(ver2)
	} else {
		shorter = len(ver1)
	}
	// 循环比较
	for i := 0; i < shorter; i++ {
		if ver1[i] == ver2[i] {
			if shorter-1 == i {
				if len(ver1) == len(ver2) {
					return 0
				} else {
					if len(ver1) > len(ver2) {
						return 1
					} else {
						return 2
					}
				}
			}
		} else if ver1[i] > ver2[i] {
			return 1
		} else {
			return 2
		}
	}
	return -1
}

func strSlice2IntSlice(strs []string) []int64 {
	if len(strs) == 0 {
		return []int64{}
	}
	retInt := make([]int64, 0, len(strs))
	for _, str := range strs {
		i, err := strconv.ParseInt(str, 10, 64)
		if err == nil {
			retInt = append(retInt, i)
		}
	}
	return retInt
}

func VersionCompare(v1, v2, operator string) bool {
	com := Compare(v1, v2)
	switch operator {
	case "==":
		if com == 0 {
			return true
		}
	case "<":
		if com == 2 {
			return true
		}
	case ">":
		if com == 1 {
			return true
		}
	case "<=":
		if com == 0 || com == 2 {
			return true
		}
	case ">=":
		if com == 0 || com == 1 {
			return true
		}
	}
	return false
}

// RandPassword 生成随机密码
func RandPassword(length int) (string, error) {
	randData := make([]byte, length)
	_, err := rand.Read(randData)

	r := make([]rune, 0)
	//特殊字符:35 ~ 38
	//数字:48 ~ 57
	//大写字母:65 ~ 90
	//小写字母:97 ~ 122
	rules := [][]int{
		{96, 123}, {64, 91}, {47, 58}, {34, 39},
	}
	for _, val := range randData {
		tmp := int(val)
		matched := false
		for _, rule := range rules {
			if tmp <= rule[0] {
				continue
			}
			if tmp = tmp % rule[1]; tmp > rule[0] {
				r = append(r, rune(tmp))
				matched = true
				break
			}
		}
		if !matched {
			//其余的归于归于数字
			r = append(r, rune((tmp%10)+48))
		}
	}
	return string(r), err
}

func Desensitize(word string) string {
	if len(word) < 3 {
		return word
	}
	sl := int(float64(len(word)) / 2.5)
	return word[:sl-1] + strings.Repeat("*", sl) + word[2*sl-1:]
}
