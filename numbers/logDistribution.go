package numbers

/*DEBUG:
import(
	"fmt"
)*/

//BinomialDistLog returns log(BinomialDist), where log is the natural logarithm.
//This is ideal for very small probabilities to avoid underflow.
func BinomialDistLog(n int, k int, p float64) float64 {
	coefficient := BinomCoefficientLog(n, k)
	s := LogPowInt(p, k)
	f := LogPowInt(1.0-p, n-k)
	expression := MultiplyLog(s, f)
	return MultiplyLog(coefficient, expression)
}

//BinomialDistLogMap returns log(BinomialDist), where log is the natural logarithm.
//This function is similar to BinomialDistLog but passes in a map[int][]float64, where the int key
//refers to n and the []float64 map values are the corresponding binomial coefficients for index k.
//Useful to not recalculate the binomial coefficient each time when binomial densities must be constantly evaluated in logSpace, like in MCMC.
func BinomialDistLogMap(n int, k int, p float64, binomMap *map[int][]float64) float64 {
	if _, ok := (*binomMap)[n]; !ok {
		(*binomMap)[n] = AddBinomMapEntry(n)
	}
	s := LogPowInt(p, k)
	f := LogPowInt(1.0-p, n-k)
	expression := MultiplyLog(s, f)
	return MultiplyLog(expression, (*binomMap)[n][k])
}

//AddBinomMapEntry adds an entry to a binomMap containing a slice of binomial coefficients in logSpace for a particular n value.
func AddBinomMapEntry(n int) []float64 {
	var answer []float64
	answer = make([]float64, n+1)
	for k := 1; k < n+1; k++ {
		answer[k] = BinomCoefficientLog(n, k)
	}
	//DEBUG:fmt.Printf("Answer: %v\n", answer)
	return answer
}
