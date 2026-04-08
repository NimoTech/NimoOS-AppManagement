package pkg

import (
	"github.com/NimoTech/NimoOS-AppManagement/codegen"
	"github.com/NimoTech/NimoOS-AppManagement/common"
	"github.com/NimoTech/NimoOS-AppManagement/service"
	"github.com/compose-spec/compose-go/loader"
)

func VaildDockerCompose(yaml []byte) (err error) {
	err = nil
	// recover
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	docker, err := service.NewComposeAppFromYAML(yaml, false, false)

	ex, ok := docker.Extensions[common.ComposeExtensionNameXNimoOS]
	if !ok {
		return service.ErrComposeExtensionNameXNimoOSNotFound
	}

	var storeInfo codegen.ComposeAppStoreInfo
	if err = loader.Transform(ex, &storeInfo); err != nil {
		return
	}

	return
}
