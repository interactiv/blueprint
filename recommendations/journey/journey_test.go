package journey_test

import (
	"../journey"
	"github.com/interactiv/expect"
	"testing"
)

func TestGetDefaultJourneys(t *testing.T) {
	journeys := journey.GetDefaultJourneys()
	expect.Expect(len(journeys), t).ToBe(5)
}

func TestCost(t *testing.T) {
	e := expect.New(t)
	e.Expect(int(journey.Cost1)).ToEqual(1)
	e.Expect(int(journey.Cost2)).ToEqual(2)
	e.Expect(int(journey.Cost3)).ToEqual(3)
	e.Expect(int(journey.Cost4)).ToEqual(4)
	e.Expect(int(journey.Cost5)).ToEqual(5)
	/*
		type Foo string

		var foo Foo = "bar"
		var bar Foo = "bar"
		e.Expect(string(foo)).ToBe("bar")
		e.Expect(bar).ToBe(foo)
	*/

}

func TestCostString(t *testing.T) {
	e := expect.New(t)
	e.Expect(journey.Cost1.String()).ToEqual("$")
	e.Expect(journey.Cost2.String()).ToEqual("$$")
	e.Expect(journey.Cost3.String()).ToEqual("$$$")
	e.Expect(journey.Cost4.String()).ToEqual("$$$$")
	e.Expect(journey.Cost5.String()).ToEqual("$$$$$")
}

func TestParseCost(t *testing.T) {
	e := expect.New(t)
	var r *journey.CostRange
	r = journey.ParseCostRange("$$...$$$")
	e.Expect(r.From).ToBe(journey.Cost2)
	e.Expect(r.To).ToBe(journey.Cost3)
	r = journey.ParseCostRange("$...$$$$$")
	e.Expect(r.From).ToBe(journey.Cost1)
	e.Expect(r.To).ToBe(journey.Cost5)
}

func TestConstRangeString(t *testing.T) {
	e := expect.New(t)
	e.Expect("$$...$$$$").ToBe((&journey.CostRange{
		From: journey.Cost2,
		To:   journey.Cost4,
	}).String())
}
