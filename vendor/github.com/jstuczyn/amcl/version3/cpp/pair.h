#ifndef PAIR_ZZZ_H
#define PAIR_ZZZ_H

#include "fp12_YYY.h"
#include "ecp2_ZZZ.h"
#include "ecp_ZZZ.h"

using namespace amcl;

namespace ZZZ {
/* Pairing constants */

extern const XXX::BIG CURVE_Bnx; /**< BN curve x parameter */
extern const XXX::BIG CURVE_Cru; /**< BN curve Cube Root of Unity */

extern const XXX::BIG CURVE_W[2];	 /**< BN curve constant for GLV decomposition */
extern const XXX::BIG CURVE_SB[2][2]; /**< BN curve constant for GLV decomposition */
extern const XXX::BIG CURVE_WB[4];	 /**< BN curve constant for GS decomposition */
extern const XXX::BIG CURVE_BB[4][4]; /**< BN curve constant for GS decomposition */

/* Pairing function prototypes */
/**	@brief Calculate Miller loop for Optimal ATE pairing e(P,Q)
 *
	@param r FP12 result of the pairing calculation e(P,Q)
	@param P ECP2 instance, an element of G2
	@param Q ECP instance, an element of G1

 */
extern void PAIR_ate(YYY::FP12 *r,ECP2 *P,ECP *Q);
/**	@brief Calculate Miller loop for Optimal ATE double-pairing e(P,Q).e(R,S)
 *
	Faster than calculating two separate pairings
	@param r FP12 result of the pairing calculation e(P,Q).e(R,S), an element of GT
	@param P ECP2 instance, an element of G2
	@param Q ECP instance, an element of G1
	@param R ECP2 instance, an element of G2
	@param S ECP instance, an element of G1
 */
extern void PAIR_double_ate(YYY::FP12 *r,ECP2 *P,ECP *Q,ECP2 *R,ECP *S);
/**	@brief Final exponentiation of pairing, converts output of Miller loop to element in GT
 *
	Here p is the internal modulus, and r is the group order
	@param x FP12, on exit = x^((p^12-1)/r)
 */
extern void PAIR_fexp(YYY::FP12 *x);
/**	@brief Fast point multiplication of a member of the group G1 by a BIG number
 *
	May exploit endomorphism for speed.
	@param Q ECP member of G1.
	@param b BIG multiplier

 */
extern void PAIR_G1mul(ECP *Q,XXX::BIG b);
/**	@brief Fast point multiplication of a member of the group G2 by a BIG number
 *
	May exploit endomorphism for speed.
	@param P ECP2 member of G1.
	@param b BIG multiplier

 */
extern void PAIR_G2mul(ECP2 *P,XXX::BIG b);
/**	@brief Fast raising of a member of GT to a BIG power
 *
	May exploit endomorphism for speed.
	@param x FP12 member of GT.
	@param b BIG exponent

 */
extern void PAIR_GTpow(YYY::FP12 *x,XXX::BIG b);
/**	@brief Tests FP12 for membership of GT
 *
	@param x FP12 instance
	@return 1 if x is in GT, else return 0

 */
extern int PAIR_GTmember(YYY::FP12 *x);

}

#endif
