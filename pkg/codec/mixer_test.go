package codec

import "testing"

func TestMixerBijection(t *testing.T) {
	m := ModulusForLength(6)
	mixer, err := NewMixerFromSalt("test-salt", m)
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[uint64]struct{}, 10000)
	for seq := uint64(1); seq <= 10000; seq++ {
		mixed := mixer.Mix(seq)
		if mixed >= m {
			t.Fatalf("mixed out of range: %d", mixed)
		}
		if _, ok := seen[mixed]; ok {
			t.Fatalf("collision at seq=%d mixed=%d", seq, mixed)
		}
		seen[mixed] = struct{}{}
		if mixer.Unmix(mixed) != seq {
			t.Fatalf("unmix failed seq=%d", seq)
		}
	}
}

func TestModulus(t *testing.T) {
	if ModulusForLength(6) != 62*62*62*62*62*62 {
		t.Fatal(ModulusForLength(6))
	}
}
