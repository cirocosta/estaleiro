package fs

const (
	KeysV1Kind = "keys/v1"
)

type KeysV1 struct {
	Kind string     `yaml:"kind"`
	Data KeysV1Data `yaml:"data"`
}

type KeysV1Data struct {
	Initial bool  `yaml:"initial"`
	Keys    []Key `yaml:"keys"`
}

type Key struct {
	// hex-encoded fingerprint of the public key.
	//
	Fingerprint string   `yaml:"fingerprint"`
	Identities  []string `yaml:"identities"`
}

func NewKeysV1(initial bool, keys []Key) KeysV1 {
	return KeysV1{
		Kind: KeysV1Kind,
		Data: KeysV1Data{
			Initial: initial,
			Keys:    keys,
		},
	}
}
