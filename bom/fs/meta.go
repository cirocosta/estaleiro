package fs

const (
	MetaV1Kind = "meta/v1"
)

type MetaV1 struct {
	Kind string `yaml:"kind"`
	Data Meta   `yaml:"data"`
}

type Meta struct {
	ProductName string `yaml:"product_name"`
	Image       string `yaml:"image"`
}

func NewMetaV1(meta Meta) MetaV1 {
	return MetaV1{
		Kind: MetaV1Kind,
		Data: meta,
	}
}
