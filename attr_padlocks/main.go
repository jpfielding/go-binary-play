package main
// author: rob fielding
import (
	"crypto/rand" 
	"fmt"
	e "github.com/cloudflare/circl/ecc/bls12381"
)

func ExamplePairing() {
	P, Q := e.G1Generator(), e.G2Generator() 
	a, b := new(e.Scalar), new(e.Scalar)
	aP, bQ := new(e.G1), new(e.G2)
	ea, eb := new(e.Gt), new (e.Gt)
	
	a.Random(rand. Reader)
	b.Random(rand.Reader)
	
	dst := make([]byte, 16)
	// can only G1 be hashed? 
	P.Hash([]byte("hell world"), dst)
	aP.ScalarMult(a, P)
	bQ.ScalarMult(b, Q)
	
	g := e.Pair(P, Q) 
	ga := e.Pair(aP, Q)
	gb := e.Pair( P,bQ)
	
	ea.Exp(g, a)
	eb.Exp (g, b)
	
	linearLeft := ea. IsEqual(ga) // e (P,Q)^a == e(aP,Q)
	linearRight := eb. IsEqual(gb) // e(P,Q)^b == e(P,bQ)
	
	fmt.Print(linearLeft && linearRight)
	
	// Output: true
}
func main() {
	ExamplePairing ()
}