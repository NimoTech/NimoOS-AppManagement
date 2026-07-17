package v2

import (
	"fmt"
	"net/http"

	"github.com/NimoTech/NimoOS-AppManagement/codegen"
	"github.com/NimoTech/NimoOS-AppManagement/common"
	"github.com/NimoTech/NimoOS-AppManagement/model"
	"github.com/NimoTech/NimoOS-AppManagement/service"
	"github.com/NimoTech/NimoOS-Common/utils"
	"github.com/NimoTech/NimoOS-Common/utils/logger"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

func (a *AppManagement) GetAppGrid(ctx echo.Context) error {
	// v2 Apps
	composeAppsWithStoreInfo, err := composeAppsWithStoreInfo(ctx.Request().Context(), composeAppsWithStoreInfoOpts{
		checkIsUpdateAvailable: false,
	})
	if err != nil {
		message := err.Error()
		logger.Error("failed to list compose apps with store info", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, codegen.ResponseInternalServerError{Message: &message})
	}

	v2AppGridItems := lo.FilterMap(lo.Values(composeAppsWithStoreInfo), func(app codegen.ComposeAppWithStoreInfo, i int) (codegen.WebAppGridItem, bool) {
		if isSystemComposeApp(app.Compose) {
			return codegen.WebAppGridItem{}, false
		}

		item, err := WebAppGridItemAdapterV2(&app)
		if err != nil {
			logger.Error("failed to adapt web app grid item", zap.Error(err), zap.String("app", app.Compose.Name))
			return codegen.WebAppGridItem{}, false
		}

		return *item, true
	})

	// v1 Apps
	nimoOSApps, containers := service.MyService.Docker().GetContainerAppList(nil, nil, nil)

	v1AppGridItems := lo.Map(*nimoOSApps, func(app model.MyAppList, i int) codegen.WebAppGridItem {
		item, err := WebAppGridItemAdapterV1(&app)
		if err != nil {
			logger.Error("failed to adapt web app grid item", zap.Error(err), zap.String("app", app.Name))
			return codegen.WebAppGridItem{}
		}
		return *item
	})

	// containers from compose apps
	composeAppContainers := []codegen.ContainerSummary{}
	for _, app := range composeAppsWithStoreInfo {
		composeApp := (service.ComposeApp)(*app.Compose)
		containerLists, err := composeApp.Containers(ctx.Request().Context())
		if err != nil {
			logger.Error("failed to get containers for compose app", zap.Error(err), zap.String("app", composeApp.Name))
			continue
		}

		for _, containcontainerList := range containerLists {
			composeAppContainers = append(composeAppContainers, containcontainerList...)
		}
	}

	containerAppGridItems := lo.FilterMap(*containers, func(app model.MyAppList, i int) (codegen.WebAppGridItem, bool) {
		if lo.ContainsBy(composeAppContainers, func(container codegen.ContainerSummary) bool { return container.ID == app.ID }) {
			// already exists as compose app, skipping...
			return codegen.WebAppGridItem{}, false
		}

		// check if this is a replacement container for a compose app when applying new settings or updating.
		//
		// we need this logic so that user does not see the temporary replacement container in the UI.
		{
			container, err := service.MyService.Docker().GetContainerByName(app.Name)
			if err != nil {
				logger.Error("failed to get container by name", zap.Error(err), zap.String("container", app.Name))
				return codegen.WebAppGridItem{}, false
			}

			// see recreateContainer() func from https://github.com/docker/compose/blob/v2/pkg/compose/convergence.go
			if replaceLabel, ok := container.Labels[api.ContainerReplaceLabel]; ok {
				if lo.ContainsBy(
					composeAppContainers,
					func(container codegen.ContainerSummary) bool {
						return container.ID == replaceLabel
					},
				) {
					// this is a replacement container for a compose app, skipping...
					return codegen.WebAppGridItem{}, false
				}
			}
		}

		item, err := WebAppGridItemAdapterContainer(&app)
		if err != nil {
			logger.Error("failed to adapt web app grid item", zap.Error(err), zap.String("app", app.Name))
			return codegen.WebAppGridItem{}, false
		}
		return *item, true
	})

	// merge v1 and v2 apps
	appGridItems := []codegen.WebAppGridItem{}
	appGridItems = append(appGridItems, v2AppGridItems...)
	appGridItems = append(appGridItems, v1AppGridItems...)
	appGridItems = append(appGridItems, containerAppGridItems...)

	return ctx.JSON(http.StatusOK, codegen.GetWebAppGridOK{
		Message: utils.Ptr("This data is for internal use ONLY - will not be supported for public use."),
		Data:    &appGridItems,
	})
}

func WebAppGridItemAdapterV2(composeAppWithStoreInfo *codegen.ComposeAppWithStoreInfo) (*codegen.WebAppGridItem, error) {
	if composeAppWithStoreInfo == nil {
		return nil, fmt.Errorf("v2 compose app is nil")
	}

	// validation
	composeApp := (*service.ComposeApp)(composeAppWithStoreInfo.Compose)
	if composeApp == nil {
		return nil, fmt.Errorf("failed to get compose app")
	}

	item := &codegen.WebAppGridItem{
		AppType: codegen.V2app,
		Name:    &composeApp.Name,
		Title: lo.ToPtr(map[string]string{
			common.DefaultLanguage: composeApp.Name,
		}),
		IsUncontrolled: utils.Ptr(false),
	}

	composeAppStoreInfo := composeAppWithStoreInfo.StoreInfo
	if composeAppStoreInfo != nil {

		// item properties from store info
		item.Hostname = composeAppStoreInfo.Hostname
		item.Icon = &composeAppStoreInfo.Icon
		item.Index = &composeAppStoreInfo.Index
		item.Port = &composeAppStoreInfo.PortMap
		item.Scheme = composeAppStoreInfo.Scheme
		item.Status = composeAppWithStoreInfo.Status
		item.StoreAppID = composeAppStoreInfo.StoreAppID
		item.Title = &composeAppStoreInfo.Title
		item.IsUncontrolled = composeAppStoreInfo.IsUncontrolled

		var mainApp *types.ServiceConfig
		for i, service := range composeApp.Services {
			if service.Name == *composeAppStoreInfo.Main {
				mainApp = &composeApp.Services[i]
				item.Image = &mainApp.Image // Hengxin needs this image property for some reason...
			}
			break
		}
	}

	// item type
	itemAuthorType := composeApp.AuthorType()
	item.AuthorType = &itemAuthorType
	if composeAppWithStoreInfo.IsUncontrolled == nil {
		item.IsUncontrolled = utils.Ptr(false)
	} else {
		item.IsUncontrolled = composeAppWithStoreInfo.IsUncontrolled
	}

	// nimoos.* 桌面 label:compose 应用在主服务 labels 里声明(spec §5.1)
	if composeAppStoreInfo != nil && composeAppStoreInfo.Main != nil {
		for i := range composeApp.Services {
			if composeApp.Services[i].Name != *composeAppStoreInfo.Main {
				continue
			}
			if dm := service.ParseDesktopLabels(composeApp.Services[i].Labels); dm != nil {
				item.Desktop = utils.Ptr(true)
				if dm.WidgetPath != "" {
					item.Widget = &codegen.WebAppGridItemWidget{
						Path: dm.WidgetPath,
						W:    utils.Ptr(dm.WidgetW),
						H:    utils.Ptr(dm.WidgetH),
					}
					setWidgetRange(item.Widget, dm.WidgetMinW, dm.WidgetMinH, dm.WidgetMaxW, dm.WidgetMaxH)
				}
			}
			break
		}
	}

	return item, nil
}

// setWidgetRange 只在 label 声明了对应值(>0)时带字段——未声明保持无字段,老桌面忽略。
func setWidgetRange(w *codegen.WebAppGridItemWidget, minw, minh, maxw, maxh int) {
	if minw > 0 {
		w.Minw = utils.Ptr(minw)
	}
	if minh > 0 {
		w.Minh = utils.Ptr(minh)
	}
	if maxw > 0 {
		w.Maxw = utils.Ptr(maxw)
	}
	if maxh > 0 {
		w.Maxh = utils.Ptr(maxh)
	}
}

func WebAppGridItemAdapterV1(app *model.MyAppList) (*codegen.WebAppGridItem, error) {
	if app == nil {
		return nil, fmt.Errorf("v1 app is nil")
	}

	item := &codegen.WebAppGridItem{
		AppType:  codegen.V1app,
		Name:     &app.ID,
		Status:   &app.State,
		Image:    &app.Image,
		Hostname: &app.Host,
		Icon:     &app.Icon,
		Index:    &app.Index,
		Port:     &app.Port,
		Scheme:   (*codegen.Scheme)(&app.Protocol),
		Title: &map[string]string{
			common.DefaultLanguage: app.Name,
		},
		IsUncontrolled: &app.IsUncontrolled,
	}

	return item, nil
}

// isSystemComposeApp reports whether a compose project is a NimoOS-internal
// component (e.g. the ML backend behind built-in Photos) and should be hidden
// from the user-facing App Panel. Marked via service label `nimoos.system: "true"`.
func isSystemComposeApp(p *codegen.ComposeApp) bool {
	if p == nil {
		return false
	}
	for _, svc := range p.Services {
		if svc.Labels["nimoos.system"] == "true" {
			return true
		}
	}
	return false
}

func WebAppGridItemAdapterContainer(container *model.MyAppList) (*codegen.WebAppGridItem, error) {
	if container == nil {
		return nil, fmt.Errorf("container is nil")
	}

	item := &codegen.WebAppGridItem{
		AppType: codegen.Container,
		Name:    &container.ID,
		Status:  &container.State,
		Image:   &container.Image,
		Title: &map[string]string{
			common.DefaultLanguage: container.Name,
		},
		IsUncontrolled: &container.IsUncontrolled,
	}

	if container.Desktop {
		item.Name = &container.Name // 容器名 = 前端稳定 key(重建容器不变)
		title := container.DesktopTitle
		if title == "" {
			title = container.Name
		}
		item.Title = &map[string]string{common.DefaultLanguage: title}
		item.Desktop = utils.Ptr(true)
		if container.Icon != "" {
			item.Icon = &container.Icon
		}
		if container.Port != "" {
			item.Port = &container.Port
		}
		if container.Index != "" {
			item.Index = &container.Index
		}
		if container.Protocol != "" {
			item.Scheme = (*codegen.Scheme)(&container.Protocol)
		}
		if container.WidgetPath != "" {
			item.Widget = &codegen.WebAppGridItemWidget{
				Path: container.WidgetPath,
				W:    utils.Ptr(container.WidgetW),
				H:    utils.Ptr(container.WidgetH),
			}
			setWidgetRange(item.Widget, container.WidgetMinW, container.WidgetMinH, container.WidgetMaxW, container.WidgetMaxH)
		}
	}

	return item, nil
}
