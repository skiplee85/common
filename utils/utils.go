package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"reflect"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
)

const (
	ymdFormat  = "2006-01-02"
	timeFormat = "2006-01-02 15:04:05"
)

var (
	rnd    *rand.Rand
	str    []byte
	strLen int
)

func init() {
	src := rand.NewSource(time.Now().UnixNano())
	rnd = rand.New(src)
	str = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	strLen = len(str)
}

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

// Salt 生成随机字符串
func Salt(l int) string {
	bs := []byte{}
	for i := 0; i < l; i++ {
		bs = append(bs, str[RandInt(strLen)])
	}
	return string(bs)
}

// PasswordEncrypt 密码加密
func PasswordEncrypt(password, salt string) string {
	return MD5(MD5(password) + salt)
}

// MD5 加密
func MD5(p string) string {
	h := md5.New()
	h.Write([]byte(p))
	return hex.EncodeToString(h.Sum(nil))
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
	if n == 0 {
		return 0
	}
	return rnd.Intn(n)
}

// IsBingo 概率是否命中
func IsBingo(x, p int) bool {
	if p == 0 {
		return false
	}
	if x > RandInt(p) {
		return true
	}
	return false
}

// WeightPick 权重选择
func WeightPick(ws []int) int {
	var weight int
	gap := [][]int{}
	for _, w := range ws {
		gap = append(gap, []int{weight, weight + w})
		weight += w
	}
	if weight == 0 {
		return 0
	}
	b := RandInt(weight)
	for idx, g := range gap {
		if g[0] <= b && b < g[1] {
			return idx
		}
	}
	return 0
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

var inc = 0

// GenOrderNumber 生成订单
func GenOrderNumber() string {
	inc++
	if inc > 9 {
		inc = 0
	}
	t := time.Now()
	unix := t.UnixNano()
	return fmt.Sprintf("%s%d%d", t.Format("20060102150405"), unix, inc)
}

// =================== CBC ======================

// AesEncryptCBC 加密
func AesEncryptCBC(origData []byte, key []byte) ([]byte, error) {
	// 分组秘钥
	// NewCipher该函数限制了输入k的长度必须为16, 24或者32
	block, _ := aes.NewCipher(key)
	blockSize := block.BlockSize()                              // 获取秘钥块的长度
	origData = pkcs5Padding(origData, blockSize)                // 补全码
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize]) // 加密模式
	encrypted := make([]byte, len(origData))                    // 创建数组
	blockMode.CryptBlocks(encrypted, origData)                  // 加密
	return encrypted, nil
}

// AesDecryptCBC 解密
func AesDecryptCBC(encrypted []byte, key []byte) ([]byte, error) {
	block, _ := aes.NewCipher(key)                              // 分组秘钥
	blockSize := block.BlockSize()                              // 获取秘钥块的长度
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize]) // 加密模式
	decrypted := make([]byte, len(encrypted))                   // 创建数组
	blockMode.CryptBlocks(decrypted, encrypted)                 // 解密
	decrypted = pkcs5UnPadding(decrypted)                       // 去除补全码
	return decrypted, nil
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// =================== ECB ======================

// AesEncryptECB 加密
func AesEncryptECB(origData []byte, key []byte) []byte {
	cipher, _ := aes.NewCipher(generateKey(key))
	length := (len(origData) + aes.BlockSize) / aes.BlockSize
	plain := make([]byte, length*aes.BlockSize)
	copy(plain, origData)
	pad := byte(len(plain) - len(origData))
	for i := len(origData); i < len(plain); i++ {
		plain[i] = pad
	}
	encrypted := make([]byte, len(plain))
	// 分组分块加密
	for bs, be := 0, cipher.BlockSize(); bs <= len(origData); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Encrypt(encrypted[bs:be], plain[bs:be])
	}

	return encrypted
}

// AesDecryptECB 解密
func AesDecryptECB(encrypted []byte, key []byte) []byte {
	cipher, _ := aes.NewCipher(generateKey(key))
	decrypted := make([]byte, len(encrypted))
	//
	for bs, be := 0, cipher.BlockSize(); bs < len(encrypted); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Decrypt(decrypted[bs:be], encrypted[bs:be])
	}

	trim := 0
	if len(decrypted) > 0 {
		trim = len(decrypted) - int(decrypted[len(decrypted)-1])
	}

	return decrypted[:trim]
}

func generateKey(key []byte) (genKey []byte) {
	genKey = make([]byte, 16)
	copy(genKey, key)
	for i := 16; i < len(key); {
		for j := 0; j < 16 && i < len(key); j, i = j+1, i+1 {
			genKey[j] ^= key[i]
		}
	}
	return genKey
}

// =================== CFB ======================

// AesEncryptCFB 加密
func AesEncryptCFB(origData []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	encrypted := make([]byte, aes.BlockSize+len(origData))
	iv := encrypted[:aes.BlockSize]
	if _, err := io.ReadFull(crand.Reader, iv); err != nil {
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(encrypted[aes.BlockSize:], origData)
	return encrypted, nil
}

// AesDecryptCFB 解密
func AesDecryptCFB(encrypted []byte, key []byte) (decrypted []byte, err error) {
	block, _ := aes.NewCipher(key)
	if len(encrypted) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	iv := encrypted[:aes.BlockSize]
	encrypted = encrypted[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	decrypted = make([]byte, len(encrypted))
	stream.XORKeyStream(decrypted, encrypted)
	return decrypted, nil
}
