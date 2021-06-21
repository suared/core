package types

import (
	"log"
	"strconv"
)

const (
	//TYPESTRING string
	TYPESTRING = iota
	//TYPEINT int
	TYPEINT
	//TYPEFLOAT float
	TYPEFLOAT
	//TYPEDICT dict
	TYPEDICT
)

//KeyVal - Generic Key/Value struct for the library
type KeyVal struct {
	Key string
	Val string
	Typ uint16
}

//ValueAsInt return int
func (keyVal KeyVal) ValueAsInt() int {
	res, err := strconv.Atoi(keyVal.Val)
	if err != nil {
		log.Printf("Error: Unable to convert value to Int for KeyValAsInt, received: %v, sending -1", keyVal.Val)
		res = -1
	}
	return res
}

//ValueAsFloat return float
func (keyVal KeyVal) ValueAsFloat() float64 {
	res, err := strconv.ParseFloat(keyVal.Val, 5)
	if err != nil {
		log.Printf("Error: Unable to convert value to Float for KeyValAsFloat, received: %v, sending -1", keyVal.Val)
		res = -1
	}
	return res
}

//Str keyVal
func Str(key string, val string) KeyVal {
	return KeyVal{Key: key, Val: val, Typ: TYPESTRING}
}

//Int keyVal
func Int(key string, val int) KeyVal {
	return KeyVal{Key: key, Val: strconv.Itoa(val), Typ: TYPEINT}
}

//Float keyVal
func Float(key string, val float64) KeyVal {
	return KeyVal{Key: key, Val: strconv.FormatFloat(val, 'f', 5, 64), Typ: TYPEFLOAT}
}

/* do this later, don't thinkI will need in short run, probably add a val array that the type then dictates to undo
func NewDictKeyVal(key string, vals ...KeyVal) KeyVal {
	return KeyVal{Key: key, Val: val, Typ: TYPEDICT}
}
*/
