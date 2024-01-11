package confobject

import (
	"github.com/smarty/assertions/should"
)

type Assertion func(actual interface{}, expectedList ...interface{}) string

func Assert(actual interface{}, assert Assertion, expected ...interface{}) (bool, string) {
	if result := so(actual, assert, expected...); len(result) == 0 {
		return true, result
	} else {
		return false, result
	}
}

func so(actual interface{}, assert func(interface{}, ...interface{}) string, expected ...interface{}) string {
	return assert(actual, expected...)
}

var (
	configAssertions map[string]Assertion
)

func init() {
	configAssertions = map[string]Assertion{
		"Equal":          should.Equal,
		"NotEqual":       should.NotEqual,
		"AlmostEqual":    should.AlmostEqual,
		"NotAlmostEqual": should.NotAlmostEqual,
		"BeNil":          should.BeNil,
		"NotBeNil":       should.NotBeNil,
		"BeTrue":         should.BeTrue,
		"BeFalse":        should.BeFalse,
		"BeZeroValue":    should.BeZeroValue,

		"BeGreaterThan":          should.BeGreaterThan,
		"BeGreaterThanOrEqualTo": should.BeGreaterThanOrEqualTo,
		"BeLessThan":             should.BeLessThan,
		"BeLessThanOrEqualTo":    should.BeLessThanOrEqualTo,

		"Contain":    should.Contain,
		"NotContain": should.NotContain,
		// "ContainKey":    should.ContainKey,
		// "NotContainKey": should.NotContainKey,
		"BeIn":       should.BeIn,
		"NotBeIn":    should.NotBeIn,
		"BeEmpty":    should.BeEmpty,
		"NotBeEmpty": should.NotBeEmpty,
		"BeBlank":    should.BeBlank,
		"NotBeBlank": should.NotBeBlank,
	}

}
