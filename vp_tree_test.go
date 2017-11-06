package search

import (
	"math"
	"math/rand"
	"testing"
)

type Point struct {
	Lat, Lon float64
	Date     int
	_dist    float64
	node     *VPTreeNode
}

func (p *Point) Dist() float64 {
	return p._dist
}

func (p *Point) SetDist(dist float64) {
	p._dist = dist
}

func (p *Point) ApplyAffinity(dist float64, target VPTreeItem) float64 {
	return dist
}

func (p *Point) ShouldSkip(target VPTreeItem) bool {
	return false
}

func (p *Point) GetNode() *VPTreeNode {
	return p.node
}

func (p *Point) SetNode(node *VPTreeNode) {
	p.node = node
}

type PointDistancer struct {
}

func (p *PointDistancer) Distance(a, b VPTreeItem) float64 {
	p1, p2 := a.(*Point), b.(*Point)
	// dd := float64(p1.Date - p2.Date)
	return HaversineEarth(p1.Lat, p1.Lon, p2.Lat, p2.Lon)
	//return VincentyDistance(p1.Lat, p1.Lon, p2.Lat, p2.Lon)
}

func TestVPTreeAllPointsFindable(t *testing.T) {

	var distancer PointDistancer
	var tree VPTree
	tree.Distancer = &distancer

	// tree.MaxChildren = 0

	points := make([]VPTreeItem, 0)
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			point := Point{
				Lat:  float64(i),
				Lon:  float64(j),
				Date: i + j}
			points = append(points, &point)
		}
	}

	tree.SetItems(points)

	for i := 1; i < 10; i++ {
		for j := 0; j < 10; j++ {
			point := Point{
				Lat:  float64(i),
				Lon:  float64(j),
				Date: i + j}
			results, distances := tree.Search(&point, 1)
			if len(results) != 1 {
				t.Log("Results should have 1 item, not", len(results))
				t.FailNow()
			}
			if len(distances) != 1 {
				t.Log("Distances should have 1 item, not", len(distances))
				t.FailNow()
			}
			res := results[0].(*Point)
			dist := distances[0]
			if res.Lat != float64(i) || res.Lon != float64(j) {
				t.Log("Returned Incorrect Result", res, "not", point)
				t.FailNow()
			}
			if dist != float64(0) {
				t.Log("Distance not idempotent, expected 0 not", dist)
				t.FailNow()
			}
		}
	}

}

func TestVPTreeSimpleSearch(t *testing.T) {

	var distancer PointDistancer
	var tree VPTree
	tree.Distancer = &distancer
	// tree.MaxChildren = 0

	points := make([]VPTreeItem, 0)
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			point := Point{
				Lat:  float64(i),
				Lon:  float64(j),
				Date: i + j}
			points = append(points, &point)
		}
	}

	tree.SetItems(points)

	p := Point{
		Lat:  5,
		Lon:  5,
		Date: 10}
	results, distances := tree.Search(&p, 1)

	if results == nil {
		t.Fatal("Results should not be nil")
	}

	if distances == nil {
		t.Fatal("Distances should not be nil")
	}

	if len(results) != 1 {
		t.Log("Results should have 1 item, not", len(results))
		t.FailNow()
	}

	if len(distances) != 1 {
		t.Log("Distances should have 1 item, not", len(distances))
		t.FailNow()
	}

	res := results[0].(*Point)
	dist := distances[0]
	if res.Lat != 5 || res.Lon != 5 {
		t.Log("Returned Incorrect Result", res)
		t.Fail()
	}

	if dist != float64(0) {
		t.Log("Distance not idempotent, expected 0 not", dist)
		t.FailNow()
	}

}

func TestVPTreeParallelSearch(t *testing.T) {

	var distancer PointDistancer
	var tree VPTree
	tree.Distancer = &distancer
	// tree.MaxChildren = 0

	points := make([]VPTreeItem, 0)
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			point := Point{
				Lat:  float64(i),
				Lon:  float64(j),
				Date: i + j}
			points = append(points, &point)
		}
	}

	tree.SetItems(points)

	for i := 1; i < 100; i++ {
		for j := 0; j < 100; j++ {
			go func(lat, lon float64, date int, dist float64) {
				point := Point{
					Lat:  lat,
					Lon:  lon,
					Date: date}
				results, distances := tree.Search(&point, 1)
				if len(results) != 1 {
					t.Log("Results should have 1 item, not", len(results))
					t.FailNow()
				}
				if len(distances) != 1 {
					t.Log("Distances should have 1 item, not", len(distances))
					t.FailNow()
				}
				res := results[0].(*Point)
				dist = distances[0]
				if res.Lat != lat || res.Lon != lon {
					t.Log("Returned Incorrect Result", res, "not", point)
					t.FailNow()
				}
				if dist != float64(0) {
					t.Log("Distance not idempotent, expected 0 not", dist)
					t.FailNow()
				}
			}(float64(i), float64(j), i+j, 0)
		}
	}

}

func TestVPTreeInsert(t *testing.T) {

	var distancer PointDistancer
	var tree VPTree
	tree.Distancer = &distancer
	// tree.MaxChildren = 0

	points := make([]VPTreeItem, 0)
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			point := Point{
				Lat:  float64(i),
				Lon:  float64(j),
				Date: i + j}
			points = append(points, &point)
		}
	}

	tree.SetItems(points)

	p := Point{
		Lat:  5.5,
		Lon:  5.5,
		Date: 10}
	tree.Insert(&p)

	results, distances := tree.Search(&p, 1)

	if results == nil {
		t.Fatal("Results should not be nil")
	}

	if len(results) != 1 {
		t.Log("Results should have 1 item, not", len(results))
		t.FailNow()
	}
	if distances == nil {
		t.Fatal("Distances should not be nil")
	}
	if len(distances) != 1 {
		t.Log("Distances should have 1 item, not", len(distances))
		t.FailNow()
	}

	res := results[0].(*Point)
	dist := distances[0]

	if res.Lat != 5.5 || res.Lon != 5.5 {
		t.Log("Returned Incorrect Result", res)
		t.Fail()
	}
	if dist != float64(0) {
		t.Log("Distance not idempotent, expected 0 not", dist)
		t.FailNow()
	}

}

func TestVPTreeRemove(t *testing.T) {

	var distancer PointDistancer
	var tree VPTree
	tree.Distancer = &distancer
	// tree.MaxChildren = 0

	points := make([]VPTreeItem, 0)
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			point := Point{
				Lat:  float64(i),
				Lon:  float64(j),
				Date: i + j}
			points = append(points, &point)
		}
	}

	tree.SetItems(points)

	p := Point{
		Lat:  5,
		Lon:  5,
		Date: 10}
	tree.Remove(&p)

	results, distances := tree.Search(&p, 1)

	if results == nil {
		t.Fatal("Results should not be nil")
	}

	if len(results) != 1 {
		t.Log("Results should still return 1, not", len(results))
		t.FailNow()
	}

	if len(distances) != 1 {
		t.Log("Distances should have 1 item, not", len(distances))
		t.FailNow()
	}

	res := results[0].(*Point)
	dist := distances[0]
	if res.Lat == float64(5) && res.Lon == float64(5) {
		t.Log("Returned Incorrect Result", res, "not", p)
		t.Fail()
	}
	if dist == float64(0) {
		t.Log("Distance is idempotent, expected not 0 but was", dist)
		t.FailNow()
	}

}

func TestVPTreeAllPointsFindableAfterRemove(t *testing.T) {

	var distancer PointDistancer
	var tree VPTree
	tree.Distancer = &distancer
	// tree.MaxChildren = 0

	points := make([]VPTreeItem, 0)
	for i := 26.0; i < 27.0; i += 0.02 {
		for j := -81.0; j < -80.0; j += 0.02 {
			point := Point{
				Lat:  float64(i),
				Lon:  float64(j),
				Date: int(i + j)}
			points = append(points, &point)
		}
	}

	tree.SetItems(points)

	p := Point{
		Lat:  26.4,
		Lon:  -80.4,
		Date: int(26.4 + -80.4)}
	tree.Remove(&p)

	for i := 26.0; i < 27; i += 0.02 {
		for j := -81.0; j < -80.0; j += 0.02 {
			if math.Abs(i-26.4) > 0.0001 && math.Abs(j-(-80.4)) > 0.0001 {
				point := Point{
					Lat:  float64(i),
					Lon:  float64(j),
					Date: int(i + j)}
				results, distances := tree.Search(&point, 1)
				if len(results) != 1 {
					t.Log("Results should have 1 item, not", len(results))
					t.FailNow()
				}
				if len(distances) != 1 {
					t.Log("Distances should have 1 item, not", len(distances))
					t.FailNow()
				}
				res := results[0].(*Point)
				dist := distances[0]
				if res.Lat != float64(i) || res.Lon != float64(j) {
					t.Log("Returned Incorrect Result", res, "not", point)
					t.Fail()
				}
				if dist != float64(0) {
					t.Log("Distance not idempotent, expected 0 not", dist)
					t.FailNow()
				}
			}
		}
	}

}

func TestVPTreeRebuild(t *testing.T) {

	var distancer PointDistancer
	var tree VPTree
	tree.Distancer = &distancer
	// tree.MaxChildren = 0

	points := make([]VPTreeItem, 0)
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			point := Point{
				Lat:  float64(i),
				Lon:  float64(j),
				Date: i + j}
			points = append(points, &point)
		}
	}

	tree.SetItems(points)

	p := Point{
		Lat:  5.5,
		Lon:  5.5,
		Date: 10}
	tree.Insert(&p)

	tree.Rebuild()

	results, distances := tree.Search(&p, 1)

	if results == nil {
		t.Fatal("Results should not be nil")
	}

	if len(results) != 1 {
		t.Log("Results should have 1 item, not", len(results))
		t.FailNow()
	}
	if distances == nil {
		t.Fatal("Distances should not be nil")
	}
	if len(distances) != 1 {
		t.Log("Distances should have 1 item, not", len(distances))
		t.FailNow()
	}

	res := results[0].(*Point)
	dist := distances[0]

	if res.Lat != 5.5 || res.Lon != 5.5 {
		t.Log("Returned Incorrect Result", res)
		t.Fail()
	}
	if dist != float64(0) {
		t.Log("Distance not idempotent, expected 0 not", dist)
		t.FailNow()
	}

}

func TestRebuildAfterRemove(t *testing.T) {
	var distancer PointDistancer
	var tree VPTree
	tree.Distancer = &distancer
	// tree.MaxChildren = 0

	points := make([]VPTreeItem, 0)
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			point := Point{
				Lat:  float64(i),
				Lon:  float64(j),
				Date: i + j}
			points = append(points, &point)
		}
	}

	tree.SetItems(points)

	p := Point{
		Lat:  5,
		Lon:  5,
		Date: 10}
	tree.Remove(&p)
	tree.Rebuild()

	results, _ := tree.Search(&p, 1)

	if results == nil {
		t.Fatal("Results should not be nil")
	}

	if len(results) != 1 {
		t.Log("Results should be still return 1, not", len(results))
		t.FailNow()
	}

	res := results[0].(*Point)
	if res.Lat == float64(5) && res.Lon == float64(5) {
		t.Log("Returned Removed Point", res)
		t.Fail()
	}

	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			if i != 5 && j != 5 {
				point := Point{
					Lat:  float64(i),
					Lon:  float64(j),
					Date: i + j}
				results, _ := tree.Search(&point, 1)
				if len(results) != 1 {
					t.Log("Results should have 1 item, not", len(results))
					t.FailNow()
				}
				res := results[0].(*Point)
				if res.Lat != float64(i) || res.Lon != float64(j) {
					t.Log("Returned Incorrect Result", res, "not", point)
					t.Fail()
				}
			}
		}
	}

}

func TestVPTreeInsertAfterRemove(t *testing.T) {

	var distancer PointDistancer
	var tree VPTree
	tree.Distancer = &distancer
	// tree.MaxChildren = 0

	points := make([]VPTreeItem, 0)
	for i := 26.0; i < 27.0; i += 0.02 {
		for j := -81.0; j < -80.0; j += 0.02 {
			point := Point{
				Lat:  float64(i),
				Lon:  float64(j),
				Date: int(i + j)}
			points = append(points, &point)
		}
	}

	tree.SetItems(points)

	p := Point{
		Lat:  26.4,
		Lon:  -80.4,
		Date: int(26.4 + -80.4)}
	tree.Remove(&p)
	p = Point{
		Lat:  26.401,
		Lon:  -80.401,
		Date: -54}
	tree.Insert(&p)

	results, distances := tree.Search(&p, 1)

	if len(results) != 1 {
		t.Log("Results should have 1 item not", len(results))
		t.FailNow()
	}

	if len(distances) != 1 {
		t.Log("Distances should have 1 item not", len(distances))
		t.FailNow()
	}

	res := results[0].(*Point)
	dist := distances[0]
	if math.Abs(res.Lat-p.Lat) > 0.000001 || math.Abs(res.Lon-p.Lon) > 0.000001 {
		t.Log("Returned Incorrect Result", res, "not", p)
		t.Fail()
	}

	if dist != float64(0) {
		t.Log("Distance not idempotent, expected 0 not", dist)
		t.Fail()
	}

}

func BenchmarkTreeBuild(b *testing.B) {
	points := make([]VPTreeItem, 0)
	for i := 0; i < 1000; i++ {
		for j := 0; j < 1000; j++ {
			point := Point{
				Lat:  float64(i),
				Lon:  float64(j),
				Date: 0}
			points = append(points, &point)
		}
	}

	var tree VPTree
	// tree.MaxChildren = 0

	var distancer PointDistancer
	tree.Distancer = &distancer

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tree.SetItems(points)
	}
}

func BenchmarkTreeSearch(b *testing.B) {
	points := make([]VPTreeItem, 0)
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			point := Point{
				Lat:  float64(i),
				Lon:  float64(j),
				Date: i + j}
			points = append(points, &point)
		}
	}

	var tree VPTree
	// tree.MaxChildren = 0

	var distancer PointDistancer
	tree.Distancer = &distancer
	tree.SetItems(points)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p := Point{
			Lat:  rand.Float64() * 100.0,
			Lon:  rand.Float64() * 100.0,
			Date: i}
		tree.Search(&p, 1)
	}
}

func BenchmarkTreeInsert(b *testing.B) {
	points := make([]VPTreeItem, 0)
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			point := Point{
				Lat:  float64(i),
				Lon:  float64(j),
				Date: 0}
			points = append(points, &point)
		}
	}

	var distancer PointDistancer
	var tree VPTree
	tree.Distancer = &distancer
	tree.SetItems(points)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p := Point{
			Lat:  rand.Float64() * 100.0,
			Lon:  rand.Float64() * 100.0,
			Date: i}
		tree.Insert(&p)
	}
}

func BenchmarkTreeSearchAfterInsert(b *testing.B) {
	points := make([]VPTreeItem, 0)
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			point := Point{
				Lat:  float64(i),
				Lon:  float64(j),
				Date: 0}
			points = append(points, &point)
		}
	}

	var distancer PointDistancer
	var tree VPTree
	tree.Distancer = &distancer
	tree.SetItems(points)

	for i := 0; i < 10000; i++ {
		p := Point{
			Lat:  rand.Float64() * 100.0,
			Lon:  rand.Float64() * 100.0,
			Date: i}
		tree.Insert(&p)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p := Point{
			Lat:  rand.Float64() * 100.0,
			Lon:  rand.Float64() * 100.0,
			Date: i}
		tree.Search(&p, 1)
	}

}
