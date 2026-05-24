package base62

//  62进制编码的模块
// 字符集: 0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ
// 索引:   0-9      10-35                              36-61
//
// 十进制转62进制对照表:
// ┌───────────┬────────┐
// │  十进制   │  62进制 │
// ├───────────┼────────┤
// │     0     │    0   │
// │     1     │    1   │
// │    10     │    A   │
// │    11     │    B   │
// │    61     │    z   │
// │    62     │   10   │
// │    63     │   11   │
// └───────────┴────────┘

const base62 = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// 为了避免恶意请求，打乱上方字符串
// 例如：0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ -> 0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz

func Int2String(seq uint64) string {
	if seq == 0 {
		return string(base62[0])
	}
	bl := []byte{}
	for seq > 0 {
		mod := seq % 62
		seq /= 62
		bl = append(bl, base62[mod])
	}
	return string(reverse(bl))
}

func reverse(s []byte) []byte {
	for i := 0; i < len(s)/2; i++ {
		s[i], s[len(s)-1-i] = s[len(s)-1-i], s[i]
	}
	return s
}
