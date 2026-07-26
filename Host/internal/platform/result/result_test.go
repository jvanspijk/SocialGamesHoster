package result

import "testing"

func TestMapAndBindShortCircuit(t *testing.T) {
	source := Ok(2)
	mapped := Map(source, func(value int) int { return value * 3 })
	bound := Bind(mapped, func(value int) Outcome[string] { return Ok("six") })

	if bound.Err != nil || bound.Value != "six" {
		t.Fatalf("unexpected outcome: %#v", bound)
	}

	failed := Fail[int](Conflict("test.conflict", "conflict"))
	called := false
	result := Bind(failed, func(value int) Outcome[string] {
		called = true
		return Ok("unreachable")
	})
	if result.Err == nil || called {
		t.Fatal("failed outcome did not short circuit")
	}
}
