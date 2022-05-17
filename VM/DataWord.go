package VM

import (
	"fmt"
	"math/bits"
	"strconv"
)

const size = 8

type DataWord []uint32

func NewDataWord() DataWord {
	return DataWord{0, 0, 0, 0, 0, 0, 0, 0}
}

func arrayToDataWord(array []uint32) DataWord {
	result := NewDataWord()

	for i, v := range array {
		result[i] = v
	}
	return result
}

func (x DataWord) SetDataWord(byteArr []byte) {
	for i, j := 0, 0; i < size; i++ {
		for lShift := 0; lShift < 4; lShift, j = lShift+1, j+1 {
			x[i] = x[i] | (uint32(byteArr[j]) << (uint32(lShift) * 8))
		}
	}
}

func intToByte(integer uint32) []uint8 {
	result := make([]uint8, 4)

	for i := 0; i < 4; i++ {
		for j := 0; j < 8; j++ {
			if (1<<(i*8)+j)&int(integer) != 0 {
				result[i] = result[i] | (1 << j)
			}
		}
	}

	return result
}

func (x DataWord) toByteArray() []uint8 {

	result := make([]uint8, 1)

	for i := 0; i < len(x); i++ {
		result = append(result, intToByte(x[i])...)
	}
	return result
}

func (x DataWord) ToBinary() string {
	result := ""
	for word := 0; word < len(x); word++ {
		for bit := 0; bit < 32; bit++ {

			if x[word]&(1<<bit) == 0 {
				result = "0" + result
			} else {
				result = "1" + result
			}
		}
	}
	return result
}

//func (x DataWord) msgToDataWord(msg Message) {
//
//}

func (x DataWord) toStringHex() string {
	newX := x.ToBinary()
	xInHex, _ := strconv.ParseInt(newX, 2, 64)
	return fmt.Sprintf("%x", xInHex)
}

func (x DataWord) toInt32() uint32 {
	return x[0]
}

func (x DataWord) SetUint32(a uint32, i uint) {
	x[i] = a
}

func (x DataWord) Add(y DataWord) DataWord {
	var carry uint32 = 0
	result := NewDataWord()
	for i := 0; i < len(result); i++ {
		(result)[i], carry = bits.Add32(x[i], y[i], carry)
	}
	return result
}

func (x DataWord) Sub(y DataWord) DataWord {
	var borrow uint32 = 0
	result := NewDataWord()
	for i := 0; i < len(result); i++ {
		(result)[i], borrow = bits.Sub32(x[i], y[i], borrow)
	}
	return result
}

func (x DataWord) Multiply(y DataWord) (DataWord, bool) {
	result := mul(x, y)
	ans := result[:size]

	var isOverFlow = false
	for i := size; i < size*2; i++ {
		isOverFlow = isOverFlow || result[i] != 0
	}

	return arrayToDataWord(ans), isOverFlow
}

func mul(x, y DataWord) (result [size * 2]uint32) {

	for Yi := 0; Yi < len(y); Yi++ {
		var carry uint32 = 0
		Ri := Yi
		Xi := 0
		for ; Xi < len(x); Xi = Xi + 1 {
			var lastRes = result[Xi+Ri]

			carry, result[Xi+Ri] = multiplyHelper(lastRes, x[Xi], y[Yi], carry)
		}
		result[Ri+Xi] = carry
	}
	return
}

func multiplyHelper(z, x, y, carry uint32) (hi, lo uint32) {
	hi, lo = bits.Mul32(x, y)
	lo, carry = bits.Add32(lo, carry, 0)
	hi, _ = bits.Add32(hi, 0, carry)
	lo, carry = bits.Add32(lo, z, 0)
	hi, _ = bits.Add32(hi, 0, carry)
	return hi, lo
}

func reciprocal2by1(d uint32) uint32 {
	reciprocal, _ := bits.Div32(^d, ^uint32(0), d)
	return reciprocal
}

func udivrem2by1(uh, ul, d, reciprocal uint32) (quot, rem uint32) {
	qh, ql := bits.Mul32(reciprocal, uh)
	ql, carry := bits.Add32(ql, ul, 0)
	qh, _ = bits.Add32(qh, uh, carry)
	qh++

	r := ul - qh*d

	if r > ql {
		qh--
		r += d
	}

	if r >= d {
		qh++
		r -= d
	}

	return qh, r
}

func (x DataWord) Div(y DataWord) []uint32 {

	//b := 1 << 32
	u, v := normalize(x, y)

	vh := v[len(v)-1]
	vl := v[len(v)-2]
	reciprocal := reciprocal2by1(vh)

	for j := len(u) - len(v) - 1; j >= 0; j-- {
		u2 := u[j+len(v)]
		u1 := u[j+len(v)-1]
		u0 := u[j+len(v)-2]

		var qhat, rhat uint32
		if u2 >= vh { // Division overflows.
			qhat = ^uint32(0)
		} else {
			qhat, rhat = udivrem2by1(u2, u1, vh, reciprocal)
			ph, pl := bits.Mul32(qhat, vl)
			if ph > rhat || (ph == rhat && pl > u0) {
				qhat--
			}
		}
	}

	q := make([]uint32, 8)

	return q
}

func normalize(u, y DataWord) ([]uint32, []uint32) {

	var yLen int
	for i := len(y) - 1; i >= 0; i-- {
		if y[i] != 0 {
			yLen = i + 1
			break
		}
	}
	shift := bits.LeadingZeros32(y[yLen-1])
	fmt.Println(shift)
	var ynStorage = NewDataWord()
	yn := ynStorage[:yLen]

	for i := yLen - 1; i > 0; i-- {
		yn[i] = (y[i] << shift) | (yn[i-1] >> (32 - shift))
	}
	yn[0] = y[0] << shift

	var uLen int
	for i := len(u) - 1; i >= 0; i-- {
		if u[i] != 0 {
			uLen = i + 1
			break
		}
	}

	var unStorage = make([]uint32, 9)
	un := unStorage[:uLen+1]
	un[uLen] = u[uLen-1] >> (32 - shift)
	for i := uLen - 1; i > 0; i-- {
		un[i] = (u[i] << shift) | (u[i-1] >> (64 - shift))
	}
	un[0] = u[0] << shift

	fmt.Println(u, un)
	fmt.Println(y, yn)

	return un, yn
}

func (x DataWord) Mod(y DataWord) (result DataWord) {

	return
}

func (x DataWord) GT(y DataWord) bool {
	_, borrow := bits.Sub32(x[0], y[0], 0)
	for i := 1; i < len(x); i++ {
		_, borrow = bits.Sub32(x[i], y[i], borrow)
	}
	return borrow == 0
}

func (x DataWord) LT(y DataWord) bool {
	return !x.GT(y)
}

func (x DataWord) SLT(y DataWord) bool {
	dataWordSign := x.sign()
	xSign := y.sign()

	if xSign > dataWordSign {
		return true
	} else if xSign < dataWordSign {
		return false
	} else {
		return x.LT(y)
	}
}

func (x DataWord) SGT(y DataWord) bool {
	dataWordSign := x.sign()
	xSign := y.sign()

	if xSign < dataWordSign {
		return true
	} else if xSign > dataWordSign {
		return false
	} else {
		return x.GT(y)
	}
}

/*
	Returns the sign of the dataWord
	if dataWord > 0 return 1
	if dataWord < 0 return -1
	if dataWord == 0 return 0
*/
func (x DataWord) sign() int {
	if x.IsZero() {
		return 0
	}
	if x[len(x)-1]&1<<31 != 0 {
		return -1
	}
	return 1
}

func (x DataWord) Eq(y DataWord) bool {
	isEqual := true
	for i := 0; i < len(x); i++ {
		isEqual = isEqual && x[i] == y[i]
	}
	return isEqual
}

func (x DataWord) IsZero() bool {
	for i := 0; i < len(x); i++ {
		if x[i] != 0 {
			return false
		}
	}
	return true
}

func (x DataWord) And(y DataWord) DataWord {
	result := NewDataWord()
	for i := 0; i < len(x); i++ {
		(result)[i] = x[i] & y[i]
	}
	return result
}

func (x DataWord) Or(y DataWord) DataWord {
	result := NewDataWord()
	for i := 0; i < len(x); i++ {
		(result)[i] = x[i] | y[i]
	}
	return result
}

func (x DataWord) Not() (result DataWord) {

	for i := 0; i < len(x); i++ {
		(result)[i] = ^x[i]
	}
	return
}

func (x DataWord) Xor(y DataWord) DataWord {
	result := NewDataWord()
	for i := 0; i < len(x); i++ {
		(result)[i] = x[i] ^ y[i]
	}
	return result
}
