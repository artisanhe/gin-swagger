package codegen_test

import (
	"testing"

	"git.chinawayltd.com/golib/gin-swagger/codegen"
	"github.com/stretchr/testify/assert"
)

func TestPrinter(tt *testing.T) {
	t := assert.New(tt)

	t.Equal("package some_package\n", codegen.DeclPackage("some_package"))
	t.Equal("type Test int\n", codegen.DeclType("Test", "int"))
}
