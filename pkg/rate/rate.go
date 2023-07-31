package rate

import (
	"crypto/sha256"
	"math/big"
)

func Rate(x string, max int64) int64 {
	s := sha256.New().Sum([]byte(x))
	var i big.Int
	i.SetBytes(s)
	i.Mod(&i, big.NewInt(max))
	return i.Int64()
}
