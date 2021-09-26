package column

import (
	"fmt"
	"regexp"
	"strconv"
)

type address struct {
	regexp *regexp.Regexp
	num    int
}

func (a address) match(s string, index int) bool {
	if a.regexp == nil {
		return a.num-1 == index
	} else {
		return a.regexp.MatchString(s)
	}
}

type endAddress struct {
	address
	isAroundContext bool
}

type SwitchSelector struct {
	begin address
	end   endAddress
}

func between(a, max, min int) int {
	if a < min {
		return 0
	} else if a > max {
		return max
	}

	return a
}

func (s SwitchSelector) Select(strings []string) ([]string, error) {
	max := len(strings)
	min := 0
	var rt []string
	if s.end.isAroundContext {
		for i, v := range strings {
			if s.begin.match(v, i) {
				if s.end.num < 0 {
					rt = append(rt, strings[between(i+s.end.num, max, min):between(i+1, max, min)]...)
				} else {
					rt = append(rt, strings[between(i, max, min):between(i+s.end.num+1, max, min)]...)
				}
			}
		}
	} else {
		st := false
		for i, v := range strings {
			if st {
				rt = append(rt, v)
				if s.end.match(v, i) {
					st = false
				}
			} else {
				st = s.begin.match(v, i)
				if st {
					rt = append(rt, v)
				}
			}
		}
	}

	return rt, nil
}

var numberAddress, _ = regexp.Compile(`^\d+$`)
var regexpAddress, _ = regexp.Compile(`^/.+/$`)
var aroundContextAddress, _ = regexp.Compile(`^[+-]\d+$`)

func newAddress(q string) (address, error) {
	if numberAddress.MatchString(q) {
		num, err := strconv.Atoi(q)
		return address{
			num: num,
		}, err
	} else if regexpAddress.MatchString(q) {
		r, err := regexp.Compile(q[1 : len(q)-1])
		return address{
			regexp: r,
		}, err
	}

	return address{}, fmt.Errorf("%s is not valid address", q)
}

func newEndAddress(q string) (endAddress, error) {
	if aroundContextAddress.MatchString(q) {
		num, err := strconv.Atoi(q)
		return endAddress{
			isAroundContext: true,
			address:         address{num: num},
		}, err
	} else {
		adr, err := newAddress(q)
		return endAddress{
			address:         adr,
			isAroundContext: false,
		}, err
	}
}

func NewSwitchSelector(begin, end string) (SwitchSelector, error) {
	ba, err := newAddress(begin)
	if err != nil {
		return SwitchSelector{}, err
	}
	ea, err := newEndAddress(end)
	if err != nil {
		return SwitchSelector{}, err
	}

	return SwitchSelector{begin: ba, end: ea}, nil
}
