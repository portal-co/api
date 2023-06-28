package mangi

func Mangle(x string, base rune) string {
	y := []rune{}
	for _, r := range x {
		if r == '_' {
			y = append(y, []rune("__")...)
		} else if r == base {
			y = append(y, '_')
		} else {
			y = append(y, r)
		}
	}
	return string(y)
}
func Demangle(x string, base rune) string {
	y := []rune{}
	var a bool
	for i, r := range x {
		if !a {
			if r == '_' {
				if []rune(x)[i+1] == '_' {
					y = append(y, '_')
					a = true
				} else {
					y = append(y, base)
				}
			} else {
				y = append(y, r)
			}
		}
		a = false
	}
	return string(y)
}
