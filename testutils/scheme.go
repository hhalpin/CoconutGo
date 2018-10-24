// schemer.go - Shared test functions for Coconut implementations
// Copyright (C) 2018  Jedrzej Stuczynski.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package schemetest provides functions used for testing both regular and concurrent coconut scheme.
package schemetest

import (
	"math/rand"
	"testing"
	"time"

	"github.com/jstuczyn/CoconutGo/coconut/concurrency/coconutclientworker"

	"github.com/jstuczyn/CoconutGo/coconut/scheme"
	"github.com/jstuczyn/CoconutGo/coconut/utils"
	"github.com/jstuczyn/amcl/version3/go/amcl"
	Curve "github.com/jstuczyn/amcl/version3/go/amcl/BLS381"
	"github.com/stretchr/testify/assert"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomInt(seen []int, max int) int {
	candidate := 1 + rand.Intn(max)
	for _, b := range seen {
		if b == candidate {
			return randomInt(seen, max)
		}
	}
	return candidate
}

// RandomInts returns random (non-repetitive) q ints, > 0, < max
func RandomInts(q int, max int) []int {
	ints := make([]int, q)
	seen := []int{}
	for i := range ints {
		r := randomInt(seen, max)
		ints[i] = r
		seen = append(seen, r)
	}
	return ints
}

// TestKeygenProperties checks basic properties of the Coconut keys, such as whether X = g2^x.
func TestKeygenProperties(t *testing.T, params coconut.CoconutParams, sk *coconut.SecretKey, vk *coconut.VerificationKey) {
	g2p := params.G2()

	assert.True(t, g2p.Equals(vk.G2()))
	assert.True(t, Curve.G2mul(vk.G2(), sk.X()).Equals(vk.Alpha()))
	assert.Equal(t, len(sk.Y()), len(vk.Beta()))

	g2 := vk.G2()
	y := sk.Y()
	beta := vk.Beta()
	for i := range beta {
		assert.Equal(t, beta[i], Curve.G2mul(g2, y[i]))
	}
}

func interpolateRandomSubsetOfKeys(p *Curve.BIG, k int, n int, keys interface{}) []interface{} {
	indices := RandomInts(k, n)
	indicesBIG := make([]*Curve.BIG, k)
	li := make([]*Curve.BIG, k)
	for i, val := range indices {
		indicesBIG[i] = Curve.NewBIGint(val)
	}
	for i := 0; i < k; i++ {
		li[i] = utils.LagrangeBasis(i, p, indicesBIG, 0)
	}
	switch v := keys.(type) {
	case []*coconut.SecretKey:
		keySub := make([]*coconut.SecretKey, k)
		for i := range keySub {
			keySub[i] = v[indices[i]-1]
		}
		q := len(keySub[0].Y())
		polys := make([]*Curve.BIG, q+1)
		polysRet := make([]interface{}, q+1)
		for i := range polys {
			polys[i] = Curve.NewBIG()
		}
		for i := range polys {
			for j := range keySub {
				if i == 0 { // x
					polys[i] = polys[i].Plus(Curve.Modmul(li[j], keySub[j].X(), p))
				} else { // ys
					polys[i] = polys[i].Plus(Curve.Modmul(li[j], keySub[j].Y()[i-1], p))
				}
			}
		}
		for i := range polys {
			polys[i].Mod(p)
			polysRet[i] = polys[i]
		}
		return polysRet

	case []*coconut.VerificationKey:
		keySub := make([]*coconut.VerificationKey, k)
		for i := range keySub {
			keySub[i] = v[indices[i]-1]
		}
		q := len(keySub[0].Beta())
		polys := make([]*Curve.ECP2, q+1)
		polysRet := make([]interface{}, q+1)
		for i := range polys {
			polys[i] = Curve.NewECP2()
		}
		for i := range polys {
			for j := range keySub {
				if i == 0 { // alpha
					polys[i].Add(Curve.G2mul(keySub[j].Alpha(), li[j]))
				} else { // beta
					polys[i].Add(Curve.G2mul(keySub[j].Beta()[i-1], li[j]))
				}
			}
			for i := range polys {
				polysRet[i] = polys[i]
			}
			return polysRet
		}
	}
	return nil // never reached anyway, but compiler complained (even with return in default case)
}

// TestTTPKeygenProperties checks whether any 2 subsets of keys when multiplied by appropriate lagrange basis
// converge to the same values
func TestTTPKeygenProperties(t *testing.T, params coconut.CoconutParams, sks []*coconut.SecretKey, vks []*coconut.VerificationKey, k int, n int) {
	p := params.P()

	polysSk1 := interpolateRandomSubsetOfKeys(p, k, n, sks)
	polysSk2 := interpolateRandomSubsetOfKeys(p, k, n, sks)
	for i := range polysSk1 {
		assert.Zero(t, Curve.Comp(polysSk1[i].(*Curve.BIG), polysSk2[i].(*Curve.BIG)))
	}

	polysVk1 := interpolateRandomSubsetOfKeys(p, k, n, vks)
	polysVk2 := interpolateRandomSubsetOfKeys(p, k, n, vks)
	for i := range polysVk1 {
		assert.True(t, polysVk1[i].(*Curve.ECP2).Equals(polysVk2[i].(*Curve.ECP2)))
	}
}

func setupAndKeygen(t *testing.T, q int, ccw *coconutclientworker.CoconutClientWorker) (coconut.CoconutParams, *coconut.SecretKey, *coconut.VerificationKey) {
	if ccw == nil {
		params, err := coconut.Setup(q)
		assert.Nil(t, err)

		sk, vk, err := coconut.Keygen(params)
		assert.Nil(t, err)
		return params, sk, vk
	}
	params, err := ccw.Setup(q)
	assert.Nil(t, err)

	sk, vk, err := ccw.Keygen(params)
	assert.Nil(t, err)
	return params, sk, vk
}

// TestSign verifies whether a coconut signature was correctly constructed
func TestSign(t *testing.T, ccw *coconutclientworker.CoconutClientWorker) {
	tests := []struct {
		q     int
		attrs []string
		err   error
		msg   string
	}{
		{q: 1, attrs: []string{"Hello World!"}, err: nil,
			msg: "For single attribute sig2 should be equal to (x + m * y) * sig1"},
		{q: 3, attrs: []string{"Foo", "Bar", "Baz"}, err: nil,
			msg: "For three attributes sig2 shguld be equal to (x + m1 * y1 + m2 * y2 + m3 * y3) * sig1"},
		{q: 2, attrs: []string{"Foo", "Bar", "Baz"}, err: coconut.ErrSignParams,
			msg: "Sign should fail due to invalid param combination"},
		{q: 3, attrs: []string{"Foo", "Bar"}, err: coconut.ErrSignParams,
			msg: "Sign should fail due to invalid param combination"},
	}

	for _, test := range tests {
		params, sk, _ := setupAndKeygen(t, test.q, ccw)
		p := params.P()

		attrsBig := make([]*Curve.BIG, len(test.attrs))
		var err error
		for i := range test.attrs {
			attrsBig[i], err = utils.HashStringToBig(amcl.SHA256, test.attrs[i])
			assert.Nil(t, err)
		}

		var sig *coconut.Signature
		if ccw == nil {
			sig, err = coconut.Sign(params.(*coconut.Params), sk, attrsBig)
		} else {
			sig, err = ccw.Sign(params.(*coconutclientworker.MuxParams), sk, attrsBig)
		}
		if test.err == coconut.ErrSignParams {
			assert.Equal(t, coconut.ErrSignParams, err, test.msg)
			continue // everything beyond that point is UB
		}
		assert.Nil(t, err)

		t1 := Curve.NewBIGcopy(sk.X())
		for i := range sk.Y() {
			t1 = t1.Plus(Curve.Modmul(attrsBig[i], sk.Y()[i], p))
		}

		sigTest := Curve.G1mul(sig.Sig1(), t1)
		assert.True(t, sigTest.Equals(sig.Sig2()), test.msg)
	}
}

// TestVerify checks whether only a valid coconut signature successfully verifies.
func TestVerify(t *testing.T, ccw *coconutclientworker.CoconutClientWorker) {
	tests := []struct {
		attrs          []string
		maliciousAttrs []string
		msg            string
	}{
		{attrs: []string{"Hello World!"}, maliciousAttrs: []string{},
			msg: "Should verify a valid signature on single public attribute"},
		{attrs: []string{"Foo", "Bar", "Baz"}, maliciousAttrs: []string{},
			msg: "Should verify a valid signature on multiple public attribute"},
		{attrs: []string{"Hello World!"}, maliciousAttrs: []string{"Malicious Hello World!"},
			msg: "Should not verify a signature when malicious attribute is introduced"},
		{attrs: []string{"Foo", "Bar", "Baz"}, maliciousAttrs: []string{"Foo2", "Bar2", "Baz2"},
			msg: "Should not verify a signature when malicious attributes are introduced"},
	}

	for _, test := range tests {
		params, sk, vk := setupAndKeygen(t, len(test.attrs), ccw)

		attrsBig := make([]*Curve.BIG, len(test.attrs))
		var err error
		for i := range test.attrs {
			attrsBig[i], err = utils.HashStringToBig(amcl.SHA256, test.attrs[i])
			assert.Nil(t, err)
		}

		var sig *coconut.Signature
		if ccw == nil {
			sig, err = coconut.Sign(params.(*coconut.Params), sk, attrsBig)
			assert.Nil(t, err)
			assert.True(t, coconut.Verify(params.(*coconut.Params), vk, attrsBig, sig), test.msg)
		} else {
			sig, err = ccw.Sign(params.(*coconutclientworker.MuxParams), sk, attrsBig)
			assert.Nil(t, err)
			assert.True(t, ccw.Verify(params.(*coconutclientworker.MuxParams), vk, attrsBig, sig), test.msg)
		}

		if len(test.maliciousAttrs) > 0 {
			mAttrsBig := make([]*Curve.BIG, len(test.maliciousAttrs))
			for i := range test.maliciousAttrs {
				mAttrsBig[i], err = utils.HashStringToBig(amcl.SHA256, test.maliciousAttrs[i])
				assert.Nil(t, err)
			}

			var sig2 *coconut.Signature
			if ccw == nil {
				sig2, err = coconut.Sign(params.(*coconut.Params), sk, mAttrsBig)
				assert.False(t, coconut.Verify(params.(*coconut.Params), vk, attrsBig, sig2), test.msg)
				assert.False(t, coconut.Verify(params.(*coconut.Params), vk, mAttrsBig, sig), test.msg)
			} else {
				sig2, err = ccw.Sign(params.(*coconutclientworker.MuxParams), sk, mAttrsBig)
				assert.False(t, ccw.Verify(params.(*coconutclientworker.MuxParams), vk, attrsBig, sig2), test.msg)
				assert.False(t, ccw.Verify(params.(*coconutclientworker.MuxParams), vk, mAttrsBig, sig), test.msg)
			}
		}
	}
}

// TestRandomize checks if randomizing a signature still produces a valid coconut signature.
func TestRandomize(t *testing.T, ccw *coconutclientworker.CoconutClientWorker) {
	tests := []struct {
		attrs []string
		msg   string
	}{
		{attrs: []string{"Hello World!"}, msg: "Should verify a randomized signature on single public attribute"},
		{attrs: []string{"Foo", "Bar", "Baz"}, msg: "Should verify a radomized signature on three public attribute"},
	}

	for _, test := range tests {
		params, sk, vk := setupAndKeygen(t, len(test.attrs), ccw)

		attrsBig := make([]*Curve.BIG, len(test.attrs))
		var err error
		for i := range test.attrs {
			attrsBig[i], err = utils.HashStringToBig(amcl.SHA256, test.attrs[i])
			assert.Nil(t, err)
		}

		var sig *coconut.Signature
		if ccw == nil {
			sig, err = coconut.Sign(params.(*coconut.Params), sk, attrsBig)
			assert.Nil(t, err)
			randSig := coconut.Randomize(params.(*coconut.Params), sig)
			assert.True(t, coconut.Verify(params.(*coconut.Params), vk, attrsBig, randSig), test.msg)
		} else {
			sig, err = ccw.Sign(params.(*coconutclientworker.MuxParams), sk, attrsBig)
			assert.Nil(t, err)
			randSig := ccw.Randomize(params.(*coconutclientworker.MuxParams), sig)
			assert.True(t, ccw.Verify(params.(*coconutclientworker.MuxParams), vk, attrsBig, randSig), test.msg)
		}
	}
}

// TestKeyAggregation checks correctness of aggregating single verification key.
// Aggregation of multiple verification keys is implicitly checked in other tests.
func TestKeyAggregation(t *testing.T, ccw *coconutclientworker.CoconutClientWorker) {
	tests := []struct {
		attrs []string
		pp    *coconut.PolynomialPoints
		msg   string
	}{
		{attrs: []string{"Hello World!"}, pp: nil,
			msg: "Should verify a signature when single set of verification keys is aggregated (single attribute)"},
		{attrs: []string{"Foo", "Bar", "Baz"}, pp: nil,
			msg: "Should verify a signature when single set of verification keys is aggregated (three attributes)"},
		{attrs: []string{"Hello World!"}, pp: coconut.NewPP([]*Curve.BIG{Curve.NewBIGint(1)}),
			msg: "Should verify a signature when single set of verification keys is aggregated (single attribute)"},
		{attrs: []string{"Foo", "Bar", "Baz"}, pp: coconut.NewPP([]*Curve.BIG{Curve.NewBIGint(1)}),
			msg: "Should verify a signature when single set of verification keys is aggregated (three attributes)"},
	}

	for _, test := range tests {
		params, sk, vk := setupAndKeygen(t, len(test.attrs), ccw)

		attrsBig := make([]*Curve.BIG, len(test.attrs))
		var err error
		for i := range test.attrs {
			attrsBig[i], err = utils.HashStringToBig(amcl.SHA256, test.attrs[i])
			assert.Nil(t, err)
		}

		var sig *coconut.Signature
		if ccw == nil {
			sig, err = coconut.Sign(params.(*coconut.Params), sk, attrsBig)
			assert.Nil(t, err)

			avk := coconut.AggregateVerificationKeys(params.(*coconut.Params), []*coconut.VerificationKey{vk}, test.pp)
			assert.True(t, coconut.Verify(params.(*coconut.Params), avk, attrsBig, sig), test.msg)
		} else {
			sig, err = ccw.Sign(params.(*coconutclientworker.MuxParams), sk, attrsBig)
			assert.Nil(t, err)

			avk := ccw.AggregateVerificationKeys(params.(*coconutclientworker.MuxParams), []*coconut.VerificationKey{vk}, test.pp)
			assert.True(t, ccw.Verify(params.(*coconutclientworker.MuxParams), avk, attrsBig, sig), test.msg)
		}
	}
}

// TestAggregateVerification checks whether signatures and verification keys from multiple authorities
// can be correctly aggregated and verified.
// This particular test does not test the threshold property, it is tested in separate test.
func TestAggregateVerification(t *testing.T, ccw *coconutclientworker.CoconutClientWorker) {
	tests := []struct {
		attrs          []string
		authorities    int
		maliciousAuth  int
		maliciousAttrs []string
		pp             *coconut.PolynomialPoints
		t              int
		msg            string
	}{
		{attrs: []string{"Hello World!"}, authorities: 1, maliciousAuth: 0, maliciousAttrs: []string{}, pp: nil, t: 0,
			msg: "Should verify aggregated signature when only single signature was used for aggregation"},
		{attrs: []string{"Hello World!"}, authorities: 3, maliciousAuth: 0, maliciousAttrs: []string{}, pp: nil, t: 0,
			msg: "Should verify aggregated signature when three signatures were used for aggregation"},
		{attrs: []string{"Foo", "Bar", "Baz"}, authorities: 1, maliciousAuth: 0, maliciousAttrs: []string{}, pp: nil, t: 0,
			msg: "Should verify aggregated signature when only single signature was used for aggregation"},
		{attrs: []string{"Foo", "Bar", "Baz"}, authorities: 3, maliciousAuth: 0, maliciousAttrs: []string{}, pp: nil, t: 0,
			msg: "Should verify aggregated signature when three signatures were used for aggregation"},
		{attrs: []string{"Hello World!"}, authorities: 1, maliciousAuth: 2,
			maliciousAttrs: []string{"Malicious Hello World!"},
			pp:             nil,
			t:              0,
			msg:            "Should fail to verify aggregated where malicious signatures were introduced"},
		{attrs: []string{"Foo", "Bar", "Baz"}, authorities: 3, maliciousAuth: 2,
			maliciousAttrs: []string{"Foo2", "Bar2", "Baz2"},
			pp:             nil,
			t:              0,
			msg:            "Should fail to verify aggregated where malicious signatures were introduced"},

		{attrs: []string{"Hello World!"}, authorities: 1, maliciousAuth: 0,
			maliciousAttrs: []string{},
			pp:             coconut.NewPP([]*Curve.BIG{Curve.NewBIGint(1)}),
			t:              1,
			msg:            "Should verify aggregated signature when only single signature was used for aggregation +threshold"},
		{attrs: []string{"Hello World!"}, authorities: 3, maliciousAuth: 0,
			maliciousAttrs: []string{},
			pp:             coconut.NewPP([]*Curve.BIG{Curve.NewBIGint(1), Curve.NewBIGint(2), Curve.NewBIGint(3)}),
			t:              2,
			msg:            "Should verify aggregated signature when three signatures were used for aggregation +threshold"},
		{attrs: []string{"Foo", "Bar", "Baz"}, authorities: 1, maliciousAuth: 0,
			maliciousAttrs: []string{},
			pp:             coconut.NewPP([]*Curve.BIG{Curve.NewBIGint(1)}),
			t:              1,
			msg:            "Should verify aggregated signature when only single signature was used for aggregation +threshold"},
		{attrs: []string{"Foo", "Bar", "Baz"}, authorities: 3, maliciousAuth: 0,
			maliciousAttrs: []string{},
			pp:             coconut.NewPP([]*Curve.BIG{Curve.NewBIGint(1), Curve.NewBIGint(2), Curve.NewBIGint(3)}),
			t:              2,
			msg:            "Should verify aggregated signature when three signatures were used for aggregation +threshold"},
	}

	for _, test := range tests {
		var params coconut.CoconutParams
		var err error
		if ccw == nil {
			params, err = coconut.Setup(len(test.attrs))
		} else {
			params, err = ccw.Setup(len(test.attrs))
		}
		assert.Nil(t, err)

		var sks []*coconut.SecretKey
		var vks []*coconut.VerificationKey

		// generate appropriate keys using appropriate method
		if test.pp == nil {
			sks = make([]*coconut.SecretKey, test.authorities)
			vks = make([]*coconut.VerificationKey, test.authorities)
			for i := 0; i < test.authorities; i++ {
				var sk *coconut.SecretKey
				var vk *coconut.VerificationKey
				if ccw == nil {
					sk, vk, err = coconut.Keygen(params.(*coconut.Params))
				} else {
					sk, vk, err = ccw.Keygen(params.(*coconutclientworker.MuxParams))
				}
				assert.Nil(t, err)
				sks[i] = sk
				vks[i] = vk
			}
		} else {
			if ccw == nil {
				sks, vks, err = coconut.TTPKeygen(params.(*coconut.Params), test.t, test.authorities)
			} else {
				sks, vks, err = ccw.TTPKeygen(params.(*coconutclientworker.MuxParams), test.t, test.authorities)
			}
			assert.Nil(t, err)
		}

		attrsBig := make([]*Curve.BIG, len(test.attrs))
		for i := range test.attrs {
			attrsBig[i], err = utils.HashStringToBig(amcl.SHA256, test.attrs[i])
			assert.Nil(t, err)
		}

		signatures := make([]*coconut.Signature, test.authorities)
		for i := 0; i < test.authorities; i++ {
			var sig *coconut.Signature
			if ccw == nil {
				sig, err = coconut.Sign(params.(*coconut.Params), sks[i], attrsBig)
			} else {
				sig, err = ccw.Sign(params.(*coconutclientworker.MuxParams), sks[i], attrsBig)
			}
			signatures[i] = sig
			assert.Nil(t, err)
		}

		var aSig *coconut.Signature
		var avk *coconut.VerificationKey

		if ccw == nil {
			aSig = coconut.AggregateSignatures(params.(*coconut.Params), signatures, test.pp)
			avk = coconut.AggregateVerificationKeys(params.(*coconut.Params), vks, test.pp)
			assert.True(t, coconut.Verify(params.(*coconut.Params), avk, attrsBig, aSig), test.msg)
		} else {
			aSig = ccw.AggregateSignatures(params.(*coconutclientworker.MuxParams), signatures, test.pp)
			avk = ccw.AggregateVerificationKeys(params.(*coconutclientworker.MuxParams), vks, test.pp)
			assert.True(t, ccw.Verify(params.(*coconutclientworker.MuxParams), avk, attrsBig, aSig), test.msg)
		}

		if test.maliciousAuth > 0 {
			msks := make([]*coconut.SecretKey, test.maliciousAuth)
			mvks := make([]*coconut.VerificationKey, test.maliciousAuth)
			for i := 0; i < test.maliciousAuth; i++ {
				var sk *coconut.SecretKey
				var vk *coconut.VerificationKey
				if ccw == nil {
					sk, vk, err = coconut.Keygen(params.(*coconut.Params))
				} else {
					sk, vk, err = ccw.Keygen(params.(*coconutclientworker.MuxParams))
				}
				assert.Nil(t, err)
				msks[i] = sk
				mvks[i] = vk
			}

			mAttrsBig := make([]*Curve.BIG, len(test.maliciousAttrs))
			for i := range test.maliciousAttrs {
				mAttrsBig[i], err = utils.HashStringToBig(amcl.SHA256, test.maliciousAttrs[i])
				assert.Nil(t, err)
			}

			mSignatures := make([]*coconut.Signature, test.maliciousAuth)
			for i := 0; i < test.maliciousAuth; i++ {
				var sig *coconut.Signature
				if ccw == nil {
					sig, err = coconut.Sign(params.(*coconut.Params), msks[i], mAttrsBig)
				} else {
					sig, err = ccw.Sign(params.(*coconutclientworker.MuxParams), msks[i], mAttrsBig)
				}
				mSignatures[i] = sig
				assert.Nil(t, err)
			}

			if ccw == nil {
				maSig := coconut.AggregateSignatures(params.(*coconut.Params), mSignatures, test.pp)
				mavk := coconut.AggregateVerificationKeys(params.(*coconut.Params), mvks, test.pp)
				maSig2 := coconut.AggregateSignatures(params.(*coconut.Params), append(signatures, mSignatures...), test.pp)
				mavk2 := coconut.AggregateVerificationKeys(params.(*coconut.Params), append(vks, mvks...), test.pp)

				assert.False(t, coconut.Verify(params.(*coconut.Params), mavk, attrsBig, maSig), test.msg)
				assert.False(t, coconut.Verify(params.(*coconut.Params), mavk2, attrsBig, maSig2), test.msg)

				assert.False(t, coconut.Verify(params.(*coconut.Params), avk, mAttrsBig, maSig), test.msg)
				assert.False(t, coconut.Verify(params.(*coconut.Params), mavk2, mAttrsBig, aSig), test.msg)

				assert.False(t, coconut.Verify(params.(*coconut.Params), avk, mAttrsBig, maSig2), test.msg)
				assert.False(t, coconut.Verify(params.(*coconut.Params), mavk2, mAttrsBig, maSig2), test.msg)
			} else {
				maSig := ccw.AggregateSignatures(params.(*coconutclientworker.MuxParams), mSignatures, test.pp)
				mavk := ccw.AggregateVerificationKeys(params.(*coconutclientworker.MuxParams), mvks, test.pp)
				maSig2 := ccw.AggregateSignatures(params.(*coconutclientworker.MuxParams), append(signatures, mSignatures...), test.pp)
				mavk2 := ccw.AggregateVerificationKeys(params.(*coconutclientworker.MuxParams), append(vks, mvks...), test.pp)

				assert.False(t, ccw.Verify(params.(*coconutclientworker.MuxParams), mavk, attrsBig, maSig), test.msg)
				assert.False(t, ccw.Verify(params.(*coconutclientworker.MuxParams), mavk2, attrsBig, maSig2), test.msg)

				assert.False(t, ccw.Verify(params.(*coconutclientworker.MuxParams), avk, mAttrsBig, maSig), test.msg)
				assert.False(t, ccw.Verify(params.(*coconutclientworker.MuxParams), mavk2, mAttrsBig, aSig), test.msg)

				assert.False(t, ccw.Verify(params.(*coconutclientworker.MuxParams), avk, mAttrsBig, maSig2), test.msg)
				assert.False(t, ccw.Verify(params.(*coconutclientworker.MuxParams), mavk2, mAttrsBig, maSig2), test.msg)
			}
		}
	}
}