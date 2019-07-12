package util

import (
	"fmt"
	"reflect"
	"testing"
)

var stringArr1 = []string{"a", "a", "b", "d"}
var stringArr2 = []string{"b", "c", "e"}
var intArr1 = []uint64{1, 1, 2, 4}
var intArr2 = []uint64{2, 3, 5}
var tempInterface interface{}

func TestDistinct(t *testing.T) {
	var myTests = []struct {
		input    interface{}
		pass     bool
		expected interface{}
	}{
		{stringArr1, true, []string{"a", "b", "d"}},
		{stringArr2, true, []string{"b", "c", "e"}},
		{intArr1, true, []uint64{1, 2, 4}},
		{intArr2, true, []uint64{2, 3, 5}},
		{[]int{}, true, []int{}},
	}

	for _, tt := range myTests {
		actual, ok := Distinct(tt.input)

		if tt.pass && ok && !testEq(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		} else if !tt.pass && ok {
			t.Errorf("expected fail but received: %v, ok: %v", actual, ok)
		}
	}
}

func TestIntersect(t *testing.T) {
	var myTests = []struct {
		input1   interface{}
		input2   interface{}
		pass     bool
		expected interface{}
	}{
		{stringArr1, stringArr2, true, []string{"b"}},
		{intArr1, intArr2, true, []uint64{2}},
		{stringArr1, intArr1, false, tempInterface},
		{[]string{}, []string{"1"}, true, []string{}},
	}

	for _, tt := range myTests {
		actual, ok := Intersect(tt.input1, tt.input2)

		if tt.pass && ok && !testEq(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		} else if !tt.pass && ok {
			t.Errorf("expected fail but received: %v, ok: %v", actual, ok)
		}
	}
}

func TestUnion(t *testing.T) {
	var myTests = []struct {
		input1   interface{}
		input2   interface{}
		pass     bool
		expected interface{}
	}{
		{stringArr1, stringArr2, true, []string{"a", "b", "c", "d", "e"}},
		{intArr1, intArr2, true, []uint64{1, 2, 3, 4, 5}},
		{stringArr1, intArr1, false, tempInterface},
		{[]string{}, []string{"1"}, true, []string{"1"}},
	}

	for _, tt := range myTests {
		actual, ok := Union(tt.input1, tt.input2)

		if tt.pass && ok && !testEq(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		} else if !tt.pass && ok {
			t.Errorf("expected fail but received: %v, ok: %v", actual, ok)
		}
	}
}

func TestDifference(t *testing.T) {
	var myTests = []struct {
		input1   interface{}
		input2   interface{}
		pass     bool
		expected interface{}
	}{
		{stringArr1, stringArr2, true, []string{"a", "c", "d", "e"}},
		{intArr1, intArr2, true, []uint64{1, 3, 4, 5}},
		{stringArr1, intArr1, false, tempInterface},
		{[]string{}, []string{"1"}, true, []string{"1"}},
	}

	for _, tt := range myTests {
		actual, ok := Difference(tt.input1, tt.input2)

		if tt.pass && ok && !testEq(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		} else if !tt.pass && ok {
			t.Errorf("expected fail but received: %v, ok: %v", actual, ok)
		}
	}
}

func TestIntersectString(t *testing.T) {
	var myTests = []struct {
		input1   []string
		input2   []string
		expected []string
	}{
		{stringArr1, stringArr2, []string{"b"}},
	}

	for _, tt := range myTests {
		actual := IntersectString(tt.input1, tt.input2)

		if !testString(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		}
	}
}

func TestIntersectStringArr(t *testing.T) {
	var myTests = []struct {
		input    [][]string
		expected []string
	}{
		{[][]string{stringArr1, stringArr2}, []string{"b"}},
	}

	for _, tt := range myTests {
		actual := IntersectStringArr(tt.input)

		if !testString(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		}
	}
}

func TestUnionString(t *testing.T) {
	var myTests = []struct {
		input1   []string
		input2   []string
		expected []string
	}{
		{stringArr1, stringArr2, []string{"a", "b", "c", "d", "e"}},
	}

	for _, tt := range myTests {
		actual := UnionString(tt.input1, tt.input2)

		if !testString(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		}
	}
}

func TestUnionStringArr(t *testing.T) {
	var myTests = []struct {
		input    [][]string
		expected []string
	}{
		{[][]string{stringArr1, stringArr2}, []string{"a", "b", "c", "d", "e"}},
	}

	for _, tt := range myTests {
		actual := UnionStringArr(tt.input)

		if !testString(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		}
	}
}

func TestDifferenceString(t *testing.T) {
	var myTests = []struct {
		input1   []string
		input2   []string
		expected []string
	}{
		{stringArr1, stringArr2, []string{"a", "c", "d", "e"}},
	}

	for _, tt := range myTests {
		actual := DifferenceString(tt.input1, tt.input2)

		if !testString(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		}
	}
}

func TestDifferenceStringArr(t *testing.T) {
	var myTests = []struct {
		input    [][]string
		expected []string
	}{
		{[][]string{stringArr1, stringArr2}, []string{"a", "c", "d", "e"}},
	}

	for _, tt := range myTests {
		actual := DifferenceStringArr(tt.input)

		if !testString(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		}
	}
}

func TestDistinctString(t *testing.T) {
	var myTests = []struct {
		input    []string
		expected []string
	}{
		{stringArr1, []string{"a", "b", "d"}},
		{stringArr2, []string{"b", "c", "e"}},
	}

	for _, tt := range myTests {
		actual := DistinctString(tt.input)

		if !testString(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		}
	}
}

/////////////
/////////////

func TestIntersectUint64(t *testing.T) {
	var myTests = []struct {
		input1   []uint64
		input2   []uint64
		expected []uint64
	}{
		{intArr1, intArr2, []uint64{2}},
	}

	for _, tt := range myTests {
		actual := IntersectUint64(tt.input1, tt.input2)

		if !testuInt64(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		}
	}
}

func TestDistinctIntersectUint64(t *testing.T) {
	var myTests = []struct {
		input1   []uint64
		input2   []uint64
		expected []uint64
	}{
		{[]uint64{1, 2, 4}, []uint64{2, 3, 5}, []uint64{2}},
	}

	for _, tt := range myTests {
		actual := DistinctIntersectUint64(tt.input1, tt.input2)

		if !testuInt64(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		}
	}
}

func TestDistinctIntersectUint64Arr(t *testing.T) {
	var myTests = []struct {
		input    [][]uint64
		expected []uint64
	}{
		{[][]uint64{{1, 2, 4}, {2, 3, 5}}, []uint64{2}},
	}

	for _, tt := range myTests {
		actual := DistinctIntersectUint64Arr(tt.input)

		if !testuInt64(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		}
	}
}

func TestIntersectUint64Arr(t *testing.T) {
	var myTests = []struct {
		input    [][]uint64
		expected []uint64
	}{
		{[][]uint64{intArr1, intArr2}, []uint64{2}},
	}

	for _, tt := range myTests {
		actual := IntersectUint64Arr(tt.input)

		if !testuInt64(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		}
	}
}

func TestSortedIntersectUint64(t *testing.T) {
	var myTests = []struct {
		input1   []uint64
		input2   []uint64
		expected []uint64
	}{
		{[]uint64{1, 2, 4}, []uint64{2, 3, 5}, []uint64{2}},
	}

	for _, tt := range myTests {
		actual := SortedIntersectUint64(tt.input1, tt.input2)

		if !testuInt64(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		}
	}
}

func TestSortedIntersectUint64Arr(t *testing.T) {
	var myTests = []struct {
		input    [][]uint64
		expected []uint64
	}{
		{[][]uint64{{1, 2, 4}, {2, 3, 5}}, []uint64{2}},
	}

	for _, tt := range myTests {
		actual := SortedIntersectUint64Arr(tt.input)

		if !testuInt64(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		}
	}
}

func TestUnionUint64(t *testing.T) {
	var myTests = []struct {
		input1   []uint64
		input2   []uint64
		expected []uint64
	}{
		{intArr1, intArr2, []uint64{1, 2, 3, 4, 5}},
	}

	for _, tt := range myTests {
		actual := UnionUint64(tt.input1, tt.input2)

		if !testuInt64(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		}
	}
}

func TestUnionUint64Arr(t *testing.T) {
	var myTests = []struct {
		input    [][]uint64
		expected []uint64
	}{
		{[][]uint64{intArr1, intArr2}, []uint64{1, 2, 3, 4, 5}},
	}

	for _, tt := range myTests {
		actual := UnionUint64Arr(tt.input)

		if !testuInt64(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		}
	}
}

func TestDifferenceUint64(t *testing.T) {
	var myTests = []struct {
		input1   []uint64
		input2   []uint64
		expected []uint64
	}{
		{intArr1, intArr2, []uint64{1, 3, 4, 5}},
	}

	for _, tt := range myTests {
		actual := DifferenceUint64(tt.input1, tt.input2)

		if !testuInt64(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		}
	}
}

func TestDifferenceUint64Arr(t *testing.T) {
	var myTests = []struct {
		input    [][]uint64
		expected []uint64
	}{
		{[][]uint64{intArr1, intArr2}, []uint64{1, 3, 4, 5}},
	}

	for _, tt := range myTests {
		actual := DifferenceUint64Arr(tt.input)

		if !testuInt64(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		}
	}
}

func TestDistinctUint64(t *testing.T) {
	var myTests = []struct {
		input    []uint64
		expected []uint64
	}{
		{intArr1, []uint64{1, 2, 4}},
		{intArr2, []uint64{2, 3, 5}},
	}

	for _, tt := range myTests {
		actual := DistinctUint64(tt.input)

		if !testuInt64(tt.expected, actual) {
			t.Errorf("expected: %v, received: %v", tt.expected, actual)
		}
	}
}

// Examples
func ExampleDistinct() {
	var a = []int{1, 1, 2, 3}

	z, ok := Distinct(a)
	if !ok {
		fmt.Println("Cannot find distinct")
	}

	slice, ok := z.Interface().([]int)
	if !ok {
		fmt.Println("Cannot convert to slice")
	}
	fmt.Println(slice, reflect.TypeOf(slice)) // [1, 2, 3] []int
}

func ExampleIntersect() {
	var a = []int{1, 1, 2, 3}
	var b = []int{2, 4}

	z, ok := Intersect(a, b)
	if !ok {
		fmt.Println("Cannot find intersect")
	}

	slice, ok := z.Interface().([]int)
	if !ok {
		fmt.Println("Cannot convert to slice")
	}
	fmt.Println(slice, reflect.TypeOf(slice)) // [2] []int
}

func ExampleUnion() {
	var a = []int{1, 1, 2, 3}
	var b = []int{2, 4}

	z, ok := Union(a, b)
	if !ok {
		fmt.Println("Cannot find union")
	}

	slice, ok := z.Interface().([]int)
	if !ok {
		fmt.Println("Cannot convert to slice")
	}
	fmt.Println(slice, reflect.TypeOf(slice)) // [1, 2, 3, 4] []int
}

func ExampleDifference() {
	var a = []int{1, 1, 2, 3}
	var b = []int{2, 4}

	z, ok := Difference(a, b)
	if !ok {
		fmt.Println("Cannot find difference")
	}

	slice, ok := z.Interface().([]int)
	if !ok {
		fmt.Println("Cannot convert to slice")
	}
	fmt.Println(slice, reflect.TypeOf(slice)) // [1, 3] []int
}

// Thanks! http://stackoverflow.com/a/15312097/3512709
func testEq(a, b interface{}) bool {

	if a == nil && b == nil {
		fmt.Println("Both nil")
		return true
	}

	if a == nil || b == nil {
		fmt.Println("One nil")
		return false
	}

	aSlice, ok := takeArg(a, reflect.Slice)
	if !ok {
		fmt.Println("Can't takeArg a")
		return ok
	}
	bSlice, ok := b.(reflect.Value)
	if !ok {
		fmt.Println("Can't takeArg b")
		return ok
	}
	aLen := aSlice.Len()
	bLen := bSlice.Len()

	if aLen != bLen {
		fmt.Println("Arr lengths not equal")
		return false
	}

OUTER:
	for i := 0; i < aLen; i++ {
		foundVal := false
		for j := 0; j < bLen; j++ {
			if aSlice.Index(i).Interface() == bSlice.Index(j).Interface() {
				foundVal = true
				continue OUTER
			}
		}

		if !foundVal {
			return false
		}
	}

	return true
}

func testString(a, b []string) bool {

	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

OUTER:
	for _, aEl := range a {
		foundVal := false
		for _, bEl := range b {
			if aEl == bEl {
				foundVal = true
				continue OUTER
			}
		}

		if !foundVal {
			return false
		}
	}

	return true
}

func testuInt64(a, b []uint64) bool {

	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

OUTER:
	for _, aEl := range a {
		foundVal := false
		for _, bEl := range b {
			if aEl == bEl {
				foundVal = true
				continue OUTER
			}
		}

		if !foundVal {
			return false
		}
	}

	return true
}