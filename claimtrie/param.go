package claimtrie

const (
	DefaultMaxActiveDelay    Height = 4032
	DefaultActiveDelayFactor Height = 32
)

// https://lbry.io/news/hf1807
const (
	DefaultOriginalClaimExpirationTime       Height = 262974
	DefaultExtendedClaimExpirationTime       Height = 2102400
	DefaultExtendedClaimExpirationForkHeight Height = 400155
)

var (
	paramMaxActiveDelay                    = DefaultMaxActiveDelay
	paramActiveDelayFactor                 = DefaultActiveDelayFactor
	paramOriginalClaimExpirationTime       = DefaultOriginalClaimExpirationTime
	paramExtendedClaimExpirationTime       = DefaultExtendedClaimExpirationTime
	paramExtendedClaimExpirationForkHeight = DefaultExtendedClaimExpirationForkHeight
)
