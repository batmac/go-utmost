package utmost_test

import (
	"math/rand"
	"testing"
	"time"

	utmost "github.com/batmac/go-utmost"
)

func TestUtmostBasic(t *testing.T) {
	limit := 5
	tm := utmost.New(limit)
	if tm.Limit() != limit {
		t.Fail()
	}
	if tm.InUse() != 0 {
		t.Fail()
	}
	if tm.MaxInUse() != 0 {
		t.Fail()
	}
	if tm.Dispensed() != 0 {
		t.Fail()
	}
	tm.Wait()
}

func TestUtmostNegative(t *testing.T) {
	tm := utmost.New(-1)
	if tm.Limit() != utmost.DefaultUtmost {
		t.Fail()
	}
	tm.Wait()
}

func launchDummy(tb testing.TB, nbTotal, nbLimit, maxTime int) *utmost.TicketsMachine {
	tb.Logf("launchDummy: nbTotal=%d, nbLimit=%d, maxTime=%d", nbTotal, nbLimit, maxTime)
	tm := utmost.New(nbLimit)
	if nbTotal < 1 || nbLimit < 1 || maxTime < 1 {
		return tm
	}
	for i := 0; i < nbTotal; i++ {
		tm.Go(func() {
			time.Sleep(time.Duration(maxTime) * time.Millisecond)
		})
	}
	if tm.Dispensed() != nbTotal {
		tb.Fatal("Dispensed:", tm.Dispensed(), "!=", nbTotal)
	}
	if tm.MaxInUse() > nbLimit {
		tb.Fatal("MaxInUse:", tm.MaxInUse(), ">", nbLimit)
	}
	// fmt.Println(tm.MaxInUse())
	// tm.Wait()
	return tm
}

func TestUtmostMany(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	_ = launchDummy(t, rand.Intn(8000), rand.Intn(8000), rand.Intn(100))
}

func TestUtmostManyManyTimes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	for i := 0; i < 100; i++ {
		TestUtmostMany(t)
	}
}

func BenchmarkUtmost(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tm := launchDummy(b, b.N, b.N/3, 10)
		b.StopTimer()
		tm.Wait()
		b.StartTimer()
	}
}
