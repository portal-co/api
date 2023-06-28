package tango

import (
	"crypto/rand"
	"io"
	"math/big"
)

func LCM(a *big.Int, b *big.Int) (c *big.Int) {
	x := big.Int{}
	c = big.NewInt(1)
	c.Mul(a, b)
	x.GCD(nil, nil, a, b)
	c.Div(c, &x)
	return
}

func Rate[K comparable](m map[K]*big.Rat, randO io.Reader) (K, error) {
	e := map[K]*big.Rat{}
	var s big.Rat
	for _, v := range m {
		s.Add(&s, v)
	}
	for k, v := range m {
		var w big.Rat
		e[k] = w.Quo(v, &s)
	}
	d := big.NewInt(1)
	for _, v := range e {
		d = LCM(d, v.Denom())
	}
	f := map[K]*big.Int{}
	for k, v := range e {
		f[k] = v.Num()
		f[k].Mul(f[k], d)
		f[k].Div(f[k], v.Denom())
	}
	t, err := rand.Int(randO, d)
	if err != nil {
		var null K
		return null, err
	}
	for {
		for k, v := range f {
			t.Sub(t, v)
			if t.Sign() != 1 {
				return k, nil
			}
		}
	}
}
