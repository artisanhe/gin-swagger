package program

import (
	"sort"
	"testing"
)

func TestAstFiles_SearchByPos(t *testing.T) {
	files := astFiles{
		{pos: 1},
		{pos: 10},
		{pos: 4},
		{pos: 7},
	}

	// can sort
	sort.Sort(files)
	if !sort.IsSorted(files) {
		t.Fatal("astFiles can not be sorted:", files)
	}

	// return nil when out range
	if file := files.searchByPos(0); file != nil {
		t.Errorf("searchByPos want=nil,got=%v", file)
	}
	if file := files.searchByPos(11); file != nil {
		t.Errorf("searchByPos want=nil,got=%v", file)
	}

	// can search: equals to
	if file := files.searchByPos(1); file.pos != 1 {
		t.Errorf("searchByPos want=1,got=%v", file)
	}

	// can search: greater than
	if file := files.searchByPos(5); file.pos != 4 {
		t.Errorf("searchByPos want=4,got=%v", file)
	}
}
