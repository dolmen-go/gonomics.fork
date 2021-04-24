package popgen

import (
	"github.com/vertgenlab/gonomics/numbers"
	"math"
)

//BuildFCache builds a slice of len(n) where each index i contains log(F(i | n, alpha)), where F is popgen.AFSSampleDensity
func BuildFCache(n int, alpha float64, binomCache [][]float64, integralError float64) []float64 {
	var answer []float64 = make([]float64, n, n)
	for j := 1; j < n; j++ {
		answer[j] = AfsSampleDensity(n, j, alpha, binomCache, integralError)
	}
	return answer
}

//GetFCacheSum uses the fCAche built in BuildFCache and calculates the sum from j=1 to n-1
func GetFCacheSum(fCache []float64) float64 {
	var answer float64 = math.Inf(-1)
	for j := 1; j < len(fCache); j++ {
		answer = numbers.AddLog(answer, fCache[j])
	}
	return answer
}

//AncestralAscertainmentDenominator is P(Asc | alpha), a constant normalizing factor for P(i | n, alpha) with the ancestral ascertainment correction.
func AncestralAscertainmentDenominator(fCache []float64, fCacheSum float64, d int) float64 {
	var answer float64 = math.Inf(-1)
	var current float64
	for j := 1; j < len(fCache); j++ {
		current = numbers.MultiplyLog(numbers.DivideLog(fCache[j], fCacheSum), AncestralAscertainmentProbability(len(fCache), j, d))
		answer = numbers.AddLog(answer, current)
	}
	return answer
}

//DerivedAscertainmentDenominator is P(Asc | alpha), a constant normalizing factor for P(i | n, alpha) with the derived ascertainment correction.
func DerivedAscertainmentDenominator(fCache []float64, fCacheSum float64, d int) float64 {
	var answer float64 = math.Inf(-1)
	var current float64
	for j := 1; j < len(fCache); j++ {
		current = numbers.MultiplyLog(numbers.DivideLog(fCache[j], fCacheSum), DerivedAscertainmentProbability(len(fCache), j, d))
		answer = numbers.AddLog(answer, current)
	}
	return answer
}

//AncestralAscertainmentProbability returns P(Asc | i, alpha) for ancestral allele ascertainment corrections.
func AncestralAscertainmentProbability(n int, i int, d int) float64 {
	return numbers.DivideLog(numbers.BinomCoefficientLog(n-i, d), numbers.BinomCoefficientLog(n, d))
}

//DerivedAscertainmentProbability returns P(Asc | i, alpha) for derived allele ascertainment corrections.
func DerivedAscertainmentProbability(n int, i int, d int) float64 {
	return numbers.DivideLog(numbers.BinomCoefficientLog(i, d), numbers.BinomCoefficientLog(n, d))
}

//AlleleFrequencyProbabilityAncestralAscertainment returns P(i | Asc, alpha) when the variant set has an ancestral allele ascertainment bias.
func AlleleFrequencyProbabilityAncestralAscertainment(alpha float64, i int, n int, d int, binomCache [][]float64, integralError float64) float64 {
	fCache := BuildFCache(n, alpha, binomCache, integralError)
	fCacheSum := GetFCacheSum(fCache)
	pIGivenAlpha := numbers.DivideLog(fCache[i], fCacheSum)

	return numbers.DivideLog(numbers.MultiplyLog(pIGivenAlpha, AncestralAscertainmentProbability(n, i, d)), AncestralAscertainmentDenominator(fCache, fCacheSum, d))
}

//AlleleFrequencyProbabilityDerivedAscertainment returns P(i | Asc, alpha) when the variant set has a derived allele ascertainment bias.
func AlleleFrequencyProbabilityDerivedAscertainment(alpha float64, i int, n int, d int, binomCache [][]float64, integralError float64) float64 {
	fCache := BuildFCache(n, alpha, binomCache, integralError)
	fCacheSum := GetFCacheSum(fCache)
	pIGivenAlpha := numbers.DivideLog(fCache[i], fCacheSum)

	return numbers.DivideLog(numbers.MultiplyLog(pIGivenAlpha, DerivedAscertainmentProbability(n, i, d)), DerivedAscertainmentDenominator(fCache, fCacheSum, d))
}

//AfsLikelihoodDerivedAscertainment is like AfsLikelihood, but makes a correction for divergence-based ascertainment when variant sets were selected for derived alleles between two groups of d individuals.
//More explanation can be found in Katzman et al, this is the inverse of Eq. 11 in the methods.
func AfsLikelihoodDerivedAscertainment(afs Afs, alpha []float64, binomMap [][]float64, d int, integralError float64) float64 {
	var answer float64 = 0.0
	// loop over all segregating sites
	for j := 0; j < len(afs.Sites); j++ {
		answer = numbers.MultiplyLog(answer, AlleleFrequencyProbabilityDerivedAscertainment(alpha[j], afs.Sites[j].I, afs.Sites[j].N, d, binomMap, integralError))
	}
	return answer
}

//AfsLikelihoodAncestralAscertainment is like AfsLikelihood, but makes a correction for divergence-based ascertainment when variant sets were selected for ancestral alleles (as in conserved regions like UCEs). d is the number of genomes from each group in the ascertainment process.
func AfsLikelihoodAncestralAscertainment(afs Afs, alpha []float64, binomMap [][]float64, d int, integralError float64) float64 {
	var answer float64 = 0.0
	//loop over all segregating sites
	for j := 0; j < len(afs.Sites); j++ {
		answer = numbers.MultiplyLog(answer, AlleleFrequencyProbabilityAncestralAscertainment(alpha[j], afs.Sites[j].I, afs.Sites[j].N, d, binomMap, integralError))
	}
	return answer
}
