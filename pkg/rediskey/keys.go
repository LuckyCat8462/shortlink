package rediskey

const (
	PrefixCacheByCode = "surl:c:"
	PrefixIdempotency = "surl:h:"
	BloomKey          = "surl:bf"
	PrefixPV          = "surl:pv:"
	PrefixUV          = "surl:uv:"
	PrefixLastClick   = "surl:la:"
	SequenceKey       = "surl:seq"
	PrefixRateRedirect = "surl:rl:r:"
	PrefixRateCreate   = "surl:rl:c:"
)

func CacheByCode(code string) string { return PrefixCacheByCode + code }
func Idempotency(hash string) string { return PrefixIdempotency + hash }
func PV(code string) string          { return PrefixPV + code }
func UV(code string) string          { return PrefixUV + code }
func LastClick(code string) string   { return PrefixLastClick + code }
func RateRedirect(ip string) string  { return PrefixRateRedirect + ip }
func RateCreate(key string) string   { return PrefixRateCreate + key }
