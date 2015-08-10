package search

import (
	"math"
	"testing"
)

func TestVincentyDistanceOnEquator(t *testing.T) {
	res := VincentyDistance(0, 2, 0, 0)

	if math.IsNaN(res) {
		t.Log("Returned NaN")
		t.FailNow()
	}

	expected := 222638.982

	if math.Abs(res-expected) > 0.05 {
		t.Log("Incorrect Value:", res, "expected", expected)
		t.Fail()
	}
}

func TestIdempotent(t *testing.T) {
	res := VincentyDistance(37.3319, -122.3069, 37.3319, -122.3069)

	if res != float64(0) {
		t.Log("Point is not idempotent, got", res)
		t.Fail()
	}
}
