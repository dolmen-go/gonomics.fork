package lift

import (
	"testing"
)

var ElementOverlapProbabilityTests = []struct {
	elements1File    string
	elements2File    string
	NoGapRegionsFile string
	Expected         []float64
}{
	{"testdata/EnrichmentElement1.bed",
		"testdata/EnrichmentElement2.bed",
		"testdata/EnrichmentNoGap.bed",
		[]float64{0.05782312925170068, 0.05782312925170068, 0.14814814814814814, 0.14814814814814814},
	},
	{"testdata/EnrichmentElement2.bed",
		"testdata/EnrichmentElement1.bed",
		"testdata/EnrichmentNoGap.bed",
		[]float64{0.08503401360544217, 0.09621993127147767, 0.14652014652014653, 0.15555555555555556},
	},
}

func TestElementOverlapProbabilities(t *testing.T) {
	var e1, e2, noGap []Lift
	var observed []float64
	for _, v := range ElementOverlapProbabilityTests {
		e1 = GoRead(v.elements1File)
		e2 = GoRead(v.elements2File)
		noGap = GoRead(v.NoGapRegionsFile)
		observed = ElementOverlapProbabilities(e1, e2, noGap)
		if !sliceEqual(observed, v.Expected) {
			t.Errorf("Error in ElementOverlapProbabilities. Output"+
				"not as expected. Observed: %v.", observed)
		}
	}
}

var EnrichmentPValueTests = []struct {
	OverlapProbs              []float64
	OverlapCount              int
	ExpectedExact             []float64
	ExpectedNormalApproximate []float64
}{
	{[]float64{0, 0, 0.1, 0.2},
		1,
		[]float64{1, 0.30000000000000004, 0.28, 0.9800000000000001},
		[]float64{1, 0.30000000000000004, 0.3019197410818303, 0.9659041370550892},
	},
}

func TestEnrichmentPValue(t *testing.T) {
	var observedExact, observedNormal []float64
	for _, v := range EnrichmentPValueTests {
		observedExact = EnrichmentPValueExact(v.OverlapProbs, v.OverlapCount)
		if !sliceEqual(observedExact, v.ExpectedExact) {
			t.Errorf("Error in EnrichmentPValueExact. Observed not as expected. Observed: %v.", observedExact)
		}
		observedNormal = EnrichmentPValueApproximation(v.OverlapProbs, v.OverlapCount)
		if !sliceEqual(observedNormal, v.ExpectedNormalApproximate) {
			t.Errorf("Error in EnrichmentPValueApproximation. Observed not as expected. Observed: %v.", observedNormal)
		}
	}
}

var EnrichmentPValueUpperLowerTests = []struct {
	elements1File string
	elements2File string
	noGapFile     string
	overlapCount  int
	verbose       int
	expectedUpper []float64
	expectedLower []float64
}{
	{"testdata/EnrichmentElement1.bed",
		"testdata/EnrichmentElement2.bed",
		"testdata/EnrichmentNoGap.bed",
		1,
		0,
		[]float64{1, 0.5925925925925926, 0.4734297880667843, 0.8928799245824091},
		[]float64{1, 0.23129251700680273, 0.21199358209298289, 0.981452036743397},
	},
}

func TestEnrichmentPValueUpperLower(t *testing.T) {
	var e1, e2, noGap []Lift
	var observedUpper, observedLower []float64
	for _, v := range EnrichmentPValueUpperLowerTests {
		e1 = GoRead(v.elements1File)
		e2 = GoRead(v.elements2File)
		noGap = GoRead(v.noGapFile)
		observedUpper = EnrichmentPValueUpperBound(e1, e2, noGap, v.overlapCount, v.verbose)
		if !sliceEqual(observedUpper, v.expectedUpper) {
			t.Errorf("Error in EnrichmentPValueUpperBound. Observed not as expected. Observed: %v.", observedUpper)
		}
		observedLower = EnrichmentPValueLowerBound(e1, e2, noGap, v.overlapCount, v.verbose)
		if !sliceEqual(observedLower, v.expectedLower) {
			t.Errorf("Error in EnrichmentPValueLowerBound. Observed not as expected. Observed: %v.", observedLower)
		}
	}
}

func sliceEqual(a []float64, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if (a[i]-b[i])/b[i] > 0.00001 {
			return false
		}
	}
	return true
}
