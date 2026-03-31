package service

import (
	"github.com/NimoTech/NimoOS-AppManagement/codegen"
	"github.com/NimoTech/NimoOS-AppManagement/common"
	"github.com/NimoTech/NimoOS-Common/utils/logger"
	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
)

type App types.ServiceConfig

func (a *App) StoreInfo() (codegen.AppStoreInfo, error) {
	var storeInfo codegen.AppStoreInfo

	ex, ok := a.Extensions[common.ComposeExtensionNameXNimoOS]
	if !ok {
		logger.Error("extension `x-nimoos` not found")
		// return storeInfo, ErrComposeExtensionNameXNimoOSNotFound
	}

	// add image to store info for check stable version function.
	storeInfo.Image = a.Image

	if err := loader.Transform(ex, &storeInfo); err != nil {
		return storeInfo, err
	}

	return storeInfo, nil
}
