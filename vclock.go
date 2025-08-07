package vclock

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"sort"
)

type VClock map[string]uint64

type CompareResult = int

const (
	CompareEqual      = CompareResult(0x00)
	CompareBefore     = CompareResult(0x01)
	CompareAfter      = CompareResult(0x02)
	CompareConcurrent = CompareResult(0x03)
)

func New() VClock {
	return map[string]uint64{}
}

func FromRaw(str string) (VClock, error) {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}

	decoder := gob.NewDecoder(bytes.NewReader(data))

	clock := New()

	if err = decoder.Decode(&clock); err != nil {
		return nil, err
	}

	return clock, nil
}

func (vc VClock) Raw() (string, error) {
	buf := bytes.NewBuffer(nil)
	encoder := gob.NewEncoder(buf)
	if err := encoder.Encode(vc); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func (vc VClock) Clone() VClock {
	ret := New()
	for k, v := range vc {
		ret[k] = v
	}
	return ret
}

func (vc VClock) String() string {
	data, _ := json.Marshal(vc)
	return string(data)
}

func (vc VClock) Merge(other VClock) {
	for k, v := range other {
		if vc[k] < v {
			vc[k] = v
		}
	}
}

func (vc VClock) Tick(key string) {
	vc[key]++
}

func (vc VClock) Set(key string, value uint64) {
	vc[key] = value
}

func (vc VClock) Compare(other VClock) CompareResult {
	ret := CompareEqual

	if len(vc) > len(other) {
		ret |= CompareAfter
	}

	if len(vc) < len(other) {
		ret |= CompareBefore
	}

	for k, v := range other {
		val := vc[k]
		if val < v {
			ret |= CompareBefore
		} else if val > v {
			ret |= CompareAfter
		}

		if ret == CompareConcurrent {
			return CompareConcurrent
		}
	}

	for k, v := range vc {
		val := other[k]
		if val < v {
			ret |= CompareAfter
		} else if val > v {
			ret |= CompareBefore
		}
		if ret == CompareConcurrent {
			return CompareConcurrent
		}
	}

	return ret
}

func (vc VClock) PassiveInc(before, now VClock) {
	for k, v := range now {
		if old, ok := before[k]; ok && old == v {
			vc[k]++
		} else {
			vc[k] = 1
		}
	}
}

func (vc VClock) More(threshold uint64) []string {
	ret := make([]string, 0)
	for k, v := range vc {
		if v > threshold {
			ret = append(ret, k)
		}
	}
	sort.Strings(ret)
	return ret
}

func (vc VClock) Less(threshold uint64) []string {
	ret := make([]string, 0)
	for k, v := range vc {
		if v < threshold {
			ret = append(ret, k)
		}
	}
	sort.Strings(ret)
	return ret
}

func (vc VClock) IDs() []string {
	ret := make([]string, 0, len(vc))
	for k := range vc {
		ret = append(ret, k)
	}
	sort.Strings(ret)
	return ret
}
