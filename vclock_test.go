package vclock_test

import (
	"math/rand/v2"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wmentor/vclock"
)

func TestMerge(t *testing.T) {
	t.Parallel()

	vc1 := vclock.New()

	vc1.Set("1", 1)
	vc1.Set("2", 2)
	vc1.Set("3", 3)
	vc1.Set("4", 4)

	vc2 := vclock.New()

	vc2.Set("2", 4)
	vc2.Set("3", 6)
	vc2.Set("5", 10)

	vc1.Merge(vc2)

	require.Equal(t, uint64(1), vc1["1"])
	require.Equal(t, uint64(4), vc1["2"])
	require.Equal(t, uint64(6), vc1["3"])
	require.Equal(t, uint64(4), vc1["4"])
	require.Equal(t, uint64(10), vc1["5"])
}

func TestEncodeDecode(t *testing.T) {
	t.Parallel()

	vc1 := vclock.New()

	for i := uint64(1); i < 50; i++ {
		vc1[strconv.FormatUint(i, 10)] = rand.Uint64()
	}

	raw, err := vc1.Raw()
	require.NoError(t, err)

	vc2, err := vclock.FromRaw(raw)
	require.NoError(t, err)

	require.Equal(t, vc1, vc2)
}

func TestCompare1(t *testing.T) {
	t.Parallel()

	vc1 := vclock.New()

	vc1.Set("1", 1)
	vc1.Set("2", 2)
	vc1.Set("3", 3)
	vc1.Set("4", 4)
	vc1.Set("5", 5)

	vc2 := vc1.Clone()

	require.Equal(t, vclock.CompareEqual, vc1.Compare(vc2))

	vc2.Tick("3")

	require.Equal(t, vclock.CompareBefore, vc1.Compare(vc2))
	require.Equal(t, vclock.CompareAfter, vc2.Compare(vc1))

	vc2.Tick("5")

	require.Equal(t, vclock.CompareBefore, vc1.Compare(vc2))
	require.Equal(t, vclock.CompareAfter, vc2.Compare(vc1))

	vc3 := vc1.Clone()

	vc3.Set("6", 1)

	require.Equal(t, vclock.CompareBefore, vc1.Compare(vc3))
	require.Equal(t, vclock.CompareAfter, vc3.Compare(vc1))

	require.Equal(t, vclock.CompareConcurrent, vc3.Compare(vc2))
	require.Equal(t, vclock.CompareConcurrent, vc2.Compare(vc3))

	vc1.Tick("1")

	require.Equal(t, vclock.CompareConcurrent, vc1.Compare(vc2))
	require.Equal(t, vclock.CompareConcurrent, vc2.Compare(vc1))

	require.Equal(t, vclock.CompareConcurrent, vc1.Compare(vc3))
	require.Equal(t, vclock.CompareConcurrent, vc3.Compare(vc1))
}

func TestCompare2(t *testing.T) {
	t.Parallel()

	vc1 := vclock.New()

	vc1.Set("1", 1)
	vc1.Set("2", 2)
	vc1.Set("3", 3)

	vc2 := vclock.New()

	vc2.Set("2", 2)
	vc2.Set("3", 3)
	vc2.Set("4", 1)

	require.Equal(t, vclock.CompareConcurrent, vc1.Compare(vc2))
	require.Equal(t, vclock.CompareConcurrent, vc2.Compare(vc1))
}

func TestPassiveInc(t *testing.T) {
	t.Parallel()

	pvc := vclock.New()

	before := vclock.New()
	now := vclock.New()

	now.Tick("1")
	now.Tick("2")
	now.Tick("3")
	now.Tick("4")

	expect := now.Clone()

	pvc.PassiveInc(before, now)

	require.Equal(t, expect, pvc)

	before.Merge(now)

	now.Tick("1")
	now.Tick("2")

	expect.Tick("3")
	expect.Tick("4")

	pvc.PassiveInc(before, now)

	require.Equal(t, expect, pvc)

	before.Merge(now)

	now.Tick("1")
	now.Tick("3")

	expect.Set("3", 1)
	expect.Tick("4")
	expect.Tick("2")

	pvc.PassiveInc(before, now)

	require.Equal(t, expect, pvc)

	before.Merge(now)

	now.Tick("1")
	now.Tick("3")

	expect.Tick("4")
	expect.Tick("2")

	pvc.PassiveInc(before, now)

	require.Equal(t, expect, pvc)
}

func TestMore(t *testing.T) {
	t.Parallel()

	vc := vclock.New()

	vc.Tick("1")
	vc.Tick("2")
	vc.Tick("3")

	for i := 0; i < 30; i++ {
		vc.Tick("1")
		if i%2 == 0 {
			vc.Tick("2")
		}
		if i%3 == 0 {
			vc.Tick("3")
		}
	}

	require.Equal(t, []string{}, vc.More(40))
	require.Equal(t, []string{"1"}, vc.More(30))
	require.Equal(t, []string{"1", "2"}, vc.More(15))
	require.Equal(t, []string{"1", "2", "3"}, vc.More(10))
}
