package config

type Config struct {
	Image Image `hcl:"image,block"`
}

type Image struct {
	Name      string `hcl:"name,label"`
	BaseImage struct {
		Name      string `hcl:"name"`
		Reference string `hcl:"ref,optional"`
	} `hcl:"base_image,block"`
}
