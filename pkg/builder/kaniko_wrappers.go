package builder

import (
	"github.com/GoogleContainerTools/kaniko/pkg/buildcontext"
	"github.com/GoogleContainerTools/kaniko/pkg/config"
	"github.com/GoogleContainerTools/kaniko/pkg/executor"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type KanikoInterface interface {
	GetBuildContext(srcContext string, opts buildcontext.BuildOptions) (buildcontext.BuildContext, error)
	DoBuild(opts *config.KanikoOptions) (v1.Image, error)
	DoPush(image v1.Image, opts *config.KanikoOptions) error
}

type Kaniko struct{}

var _ KanikoInterface = &Kaniko{}

func (k *Kaniko) GetBuildContext(srcContext string, opts buildcontext.BuildOptions) (buildcontext.BuildContext, error) {
	return buildcontext.GetBuildContext(srcContext, opts)
}

func (k *Kaniko) DoBuild(opts *config.KanikoOptions) (v1.Image, error) {
	return executor.DoBuild(opts)
}

func (k *Kaniko) DoPush(image v1.Image, opts *config.KanikoOptions) error {
	return executor.DoPush(image, opts)
}
