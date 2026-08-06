package v2_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/NimoTech/NimoOS-AppManagement/codegen"
	"github.com/NimoTech/NimoOS-AppManagement/common"
	"github.com/NimoTech/NimoOS-AppManagement/model"
	"github.com/NimoTech/NimoOS-AppManagement/pkg/docker"
	v2 "github.com/NimoTech/NimoOS-AppManagement/route/v2"
	"github.com/NimoTech/NimoOS-AppManagement/service"
	"github.com/NimoTech/NimoOS-Common/utils"
	"github.com/NimoTech/NimoOS-Common/utils/file"
	"go.uber.org/goleak"
	"gotest.tools/v3/assert"
)

func TestWebAppGridItemAdapter(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")) // https://github.com/census-instrumentation/opencensus-go/issues/1191

	defer func() {
		// workaround due to https://github.com/patrickmn/go-cache/issues/166
		docker.Cache = nil
		runtime.GC()
	}()

	storeRoot := t.TempDir()

	appsPath := filepath.Join(storeRoot, common.AppsDirectoryName)
	err := file.MkDir(appsPath)
	assert.NilError(t, err)

	// build test catalog
	err = file.MkDir(filepath.Join(appsPath, "test1"))
	assert.NilError(t, err)

	composeFilePath := filepath.Join(appsPath, "test1", common.ComposeYAMLFileName)

	err = file.WriteToFullPath([]byte(common.SampleComposeAppYAML), composeFilePath, 0o644)
	assert.NilError(t, err)

	composeApp, err := service.LoadComposeAppFromConfigFile("test1", composeFilePath)
	assert.NilError(t, err)

	storeInfo, err := composeApp.StoreInfo(true)
	assert.NilError(t, err)

	composeAppWithStoreInfo := codegen.ComposeAppWithStoreInfo{
		Compose:   (*codegen.ComposeApp)(composeApp),
		StoreInfo: storeInfo,
		Status:    utils.Ptr("running"),
	}

	gridItem, err := v2.WebAppGridItemAdapterV2(&composeAppWithStoreInfo)
	assert.NilError(t, err)

	assert.Equal(t, *gridItem.Icon, storeInfo.Icon)
	assert.Equal(t, *gridItem.Image, composeApp.Services[0].Image)
	assert.Equal(t, gridItem.Hostname, storeInfo.Hostname)
	assert.Equal(t, *gridItem.Port, storeInfo.PortMap)
	assert.Equal(t, *gridItem.Index, storeInfo.Index)
	assert.Equal(t, *gridItem.Status, "running")
	assert.DeepEqual(t, *gridItem.Title, storeInfo.Title)
	assert.Equal(t, *gridItem.AuthorType, codegen.ByNimoos)
	assert.Equal(t, *gridItem.IsUncontrolled, false)
}

func TestWebAppGridItemAdapterContainerDesktop(t *testing.T) {
	app := &model.MyAppList{
		ID: "cid123", Name: "my-dl", State: "running", Image: "img:1",
		Desktop: true, DesktopTitle: "下载器", Icon: "/icon.png",
		Port: "8080", Index: "/", Protocol: "http",
		WidgetPath: "/widget", WidgetW: 3, WidgetH: 2,
	}
	item, err := v2.WebAppGridItemAdapterContainer(app)
	if err != nil {
		t.Fatal(err)
	}
	if item.Desktop == nil || !*item.Desktop {
		t.Fatal("desktop flag missing")
	}
	if item.Name == nil || *item.Name != "my-dl" {
		t.Fatalf("desktop 容器 Name 应为容器名(稳定 key),got %v", item.Name)
	}
	if (*item.Title)["en_us"] != "下载器" && (*item.Title)[common.DefaultLanguage] != "下载器" {
		t.Fatalf("title missing: %v", item.Title)
	}
	if item.Icon == nil || *item.Icon != "/icon.png" || item.Port == nil || *item.Port != "8080" {
		t.Fatalf("meta missing: %+v", item)
	}
	if item.Widget == nil || item.Widget.Path != "/widget" || *item.Widget.W != 3 || *item.Widget.H != 2 {
		t.Fatalf("widget missing: %+v", item.Widget)
	}
}

func TestWebAppGridItemAdapterContainerPlain(t *testing.T) {
	// 回归红线:非 desktop 容器输出与改造前完全一致
	app := &model.MyAppList{ID: "cid456", Name: "other", State: "exited", Image: "img:2"}
	item, err := v2.WebAppGridItemAdapterContainer(app)
	if err != nil {
		t.Fatal(err)
	}
	if item.Name == nil || *item.Name != "cid456" {
		t.Fatal("plain 容器 Name 必须仍是容器 ID(现状)")
	}
	if item.Desktop != nil || item.Widget != nil || item.Icon != nil || item.Port != nil {
		t.Fatalf("plain 容器不得有新字段: %+v", item)
	}
}
