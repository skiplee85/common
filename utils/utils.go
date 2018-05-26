package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
)

const (
	ymdFormat  = "2006-01-02"
	timeFormat = "2006-01-02 15:04:05"
)

// Time 统一输出时间
type Time struct {
	time.Time
}

// MarshalJSON 实现它的json序列化方法
func (t Time) MarshalJSON() ([]byte, error) {
	var stamp = fmt.Sprintf("\"%s\"", t.Format(timeFormat))
	return []byte(stamp), nil
}

// UnmarshalJSON : 自定义从json->自定义类型
func (t *Time) UnmarshalJSON(data []byte) error {
	tt, err := time.Parse(timeFormat, strings.Replace(string(data), "\"", "", -1))
	if err != nil {
		return err
	}
	*t = Time{tt}
	return nil
}

// SetBSON 从mongo反序列
func (t *Time) SetBSON(raw bson.Raw) error {
	var tt time.Time
	raw.Unmarshal(&tt)
	*t = Time{tt}
	return nil
}

// GetBSON 保存至mongo时的类型
func (t Time) GetBSON() (interface{}, error) {
	return t.Time, nil
}

// String
func (t Time) String() string {
	return t.Format(timeFormat)
}

// Now 获取utils.Time
func Now() Time {
	return Time{time.Now()}
}

// GenToken 生成随机字符串
func GenToken(salt string) string {
	hash := md5.New()
	io.WriteString(hash, strconv.FormatInt(time.Now().UnixNano(), 10)+salt)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// SubIP 截取IP部分，去掉端口。
func SubIP(ip net.Addr) string {
	str := ip.String()
	len := strings.LastIndex(str, ":")
	return str[0:len]
}

// DayStart 获取天的开始时间
func DayStart(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// MonthStart 获取月的开始时间
func MonthStart(t time.Time) time.Time {
	year, month, _ := t.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
}

// WeekStart 获取周一开始时间
func WeekStart(t time.Time) time.Time {
	t = DayStart(t)
	wd := t.Weekday()
	if wd == time.Monday {
		return t
	}
	offset := int(time.Monday - wd)
	if offset > 0 {
		offset -= 7
	}
	return t.AddDate(0, 0, offset)
}

// RandInt 获取0-n的随机数
func RandInt(n int) int {
	src := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(src)
	return rnd.Intn(n)
}

// FillStruct 将 data 的值填充到 result 对应的字段。
func FillStruct(data map[string]interface{}, result interface{}) {
	t := reflect.ValueOf(result).Elem()
	for k, v := range data {
		val := t.FieldByName(k)
		val.Set(reflect.ValueOf(v))
	}
}

// BeforeDate 获取输入日期前一天时间
func BeforeDate(date string, days int) string {
	st, _ := time.Parse(ymdFormat, date)
	st = st.Add(-24 * time.Hour * time.Duration(days))
	return st.Format(ymdFormat)
}

// Abs 绝对值
func Abs(d int) int {
	if d > 0 {
		return d
	}
	return -d
}

// IsInArray 某个数是否在数组内
func IsInArray(s int, arr []int) bool {
	for _, v := range arr {
		if s == v {
			return true
		}
	}
	return false
}

// EarthDistance 输入两个经纬度，计算它们之间的距离(km)。
//计算公式
//C = sin(LatA*Pi/180)*sin(LatB*Pi/180) + cos(LatA*Pi/180)*cos(LatB*Pi/180)*cos((MLonA-MLonB)*Pi/180)
//
//Distance = R*Arccos(C)*Pi/180
func EarthDistance(lat1, lng1, lat2, lng2 float64) float64 {
	radius := 6378.137
	rad := math.Pi / 180.0
	lat1 = lat1 * rad
	lng1 = lng1 * rad
	lat2 = lat2 * rad
	lng2 = lng2 * rad
	theta := lng2 - lng1
	dist := math.Acos(math.Sin(lat1)*math.Sin(lat2) + math.Cos(lat1)*math.Cos(lat2)*math.Cos(theta))
	return dist * radius
}
