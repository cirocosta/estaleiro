package fs

import (
	"github.com/cirocosta/estaleiro/osrelease"
)

type OsReleaseV1 struct {
	Kind string              `yaml:"kind"`
	Data osrelease.OsRelease `yaml:"data"`
}

func NewOsReleaseV1(info osrelease.OsRelease) OsReleaseV1 {
	return OsReleaseV1{
		Kind: "osrelease/v1",
		Data: info,
	}
}
