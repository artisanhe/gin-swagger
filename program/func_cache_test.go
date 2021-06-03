package program

import (
	"sort"
	"testing"
)

func TestFns_SearchByPos(t *testing.T) {
	fs := fns{
		{pos: 1},
		{pos: 10},
		{pos: 4},
		{pos: 7},
	}

	// can sort
	sort.Sort(fs)
	if !sort.IsSorted(fs) {
		t.Fatal("fns can not be sorted:", fs)
	}

	// return nil when out range
	if file := fs.searchByPos(0); file != nil {
		t.Errorf("searchByPos want=nil,got=%v", file)
	}
	if file := fs.searchByPos(11); file != nil {
		t.Errorf("searchByPos want=nil,got=%v", file)
	}

	// can search: equals to
	if file := fs.searchByPos(1); file.pos != 1 {
		t.Errorf("searchByPos want=1,got=%v", file)
	}

	// can search: greater than
	if file := fs.searchByPos(5); file.pos != 4 {
		t.Errorf("searchByPos want=4,got=%v", file)
	}
}
