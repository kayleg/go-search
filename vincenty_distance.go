package search

import (
	"math"
)

// WGS-84 ellipsoid params
const (
	a   = 6378137        // Semi-Major Axis in meters
	b   = 6356752.314245 // Semi-Minor Axis in meters
	f   = (a - b) / a    // Flattening
	dTr = 0.0174532925   // Degrees to Radians
)

// VincentyDistance returns an accurate approximate distance between two
// coordinates. The result is in meters
func VincentyDistance(lat1, lon1, lat2, lon2 float64) float64 {
	L := (lon2 - lon1) * dTr
	U1 := math.Atan((1 - f) * math.Tan(lat1*dTr))
	U2 := math.Atan((1 - f) * math.Tan(lat2*dTr))
	sinU1, cosU1 := math.Sin(U1), math.Cos(U1)
	sinU2, cosU2 := math.Sin(U2), math.Cos(U2)

	lambda := L
	var lambdaP, sinSigma, sigma, cosSqAlpha, cosSigma, cos2SigmaM float64
	iterLimit := 100

	loop := true
	for loop && iterLimit > 0 {
		sinLambda, cosLambda := math.Sin(lambda), math.Cos(lambda)
		cosU2sinLambda := cosU2 * sinLambda
		cosU2cosLambda := cosU2 * cosLambda
		cosU1SinU2 := cosU1 * sinU2
		cosU1SinU2mSinU1tCosU2cosLambda := cosU1SinU2 - sinU1*cosU2cosLambda
		sinSigma = math.Sqrt(cosU2sinLambda*cosU2sinLambda + cosU1SinU2mSinU1tCosU2cosLambda*cosU1SinU2mSinU1tCosU2cosLambda)

		if sinSigma == 0 { // co-incident points
			return 0
		}
		cosSigma = sinU1*sinU2 + cosU1*cosU2cosLambda
		sigma = math.Atan2(sinSigma, cosSigma)
		sinAlpha := cosU1 * cosU2sinLambda / sinSigma
		cosSqAlpha = 1 - sinAlpha*sinAlpha
		cos2SigmaM = cosSigma - 2*sinU1*sinU2/cosSqAlpha

		// Equatorial Line: cosSqAlpha = 0
		if math.IsNaN(cos2SigmaM) {
			cos2SigmaM = 0
		}

		C := f / 16 * cosSqAlpha * (4 + f*(4-3*cosSqAlpha))
		lambdaP = lambda
		lambda = L + (1-C)*f*sinAlpha*(sigma+C*sinSigma*(cos2SigmaM+C*cosSigma*(-1+2*cos2SigmaM*cos2SigmaM)))
		iterLimit--
		loop = math.Abs(lambda-lambdaP) > 1e-12
	}

	// Did not converge
	if iterLimit == 0 {
		return math.NaN()
	}

	uSq := cosSqAlpha * (a*a - b*b) / (b * b)
	A := 1 + uSq/16384*(4096+uSq*(-768+uSq*(320-175*uSq)))
	B := uSq / 1024 * (256 + uSq*(-128+uSq*(74-47*uSq)))
	deltaSigma := B * sinSigma * (cos2SigmaM + B/4*(cosSigma*(-1+2*cos2SigmaM*cos2SigmaM)-B/6*cos2SigmaM*(-3+4*sinSigma*sinSigma)*(-3+4*cos2SigmaM*cos2SigmaM)))
	return b * A * (sigma - deltaSigma) //Fix at 3 to round to 1mm precision
}
