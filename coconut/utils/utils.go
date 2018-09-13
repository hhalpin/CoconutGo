package utils

import (
	"errors"

	"github.com/milagro-crypto/amcl/version3/go/amcl"
	"github.com/milagro-crypto/amcl/version3/go/amcl/BLS381"
)

func hashString(sha int, m string) ([]byte, error) {
	b := []byte(m)
	// below is based on the amcl implementation: https://github.com/milagro-crypto/amcl/blob/22f62d8215adf5672017c11d2f6885afb00268c4/version3/go/MPIN.go#L83
	var R []byte

	if sha == amcl.SHA256 {
		H := amcl.NewHASH256()
		H.Process_array(b)
		R = H.Hash()
	} else if sha == amcl.SHA384 {
		H := amcl.NewHASH384()
		H.Process_array(b)
		R = H.Hash()
	} else if sha == amcl.SHA512 {
		H := amcl.NewHASH512()
		H.Process_array(b)
		R = H.Hash()
	}

	if R == nil {
		return []byte{}, errors.New("Nil hash result")
	}

	const RM int = int(BLS381.MODBYTES)
	var W [RM]byte
	if sha >= RM {
		for i := 0; i < RM; i++ {
			W[i] = R[i]
		}
	} else {
		for i := 0; i < sha; i++ {
			W[i+RM-sha] = R[i]
		}
		for i := 0; i < RM-sha; i++ {
			W[i] = 0
		}
	}
	return W[:], nil
}

// todo: does it need to be public?
// is this a valid way of doing it? check edge cases with different algorithms
func HashStringToBig(sha int, m string) (*BLS381.BIG, error) {
	hash, err := hashString(sha, m)
	if err != nil {
		return nil, err
	}
	return BLS381.FromBytes(hash), nil
}

func HashStringToG1(sha int, m string) (*BLS381.ECP, error) {
	hash, err := hashString(sha, m)
	if err != nil {
		return nil, err
	}
	return BLS381.ECP_mapit(hash), nil
}
