package service

import (
	"context"
	"time"

	"github.com/NimoTech/NimoOS-AppManagement/common"
	"github.com/NimoTech/NimoOS-Common/utils/logger"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

// MonitorDockerEvents 常驻订阅 docker daemon 事件流,把容器 start/die/destroy 转发到
// MessageBus,覆盖终端/外部工具直接操作容器、本服务 API 无从感知的场景(桌面秒级同步)。
// 流断开(daemon 重启等)后退避重连;ctx 取消即退出。
func MonitorDockerEvents(ctx context.Context) {
	logger.Info("docker events monitor started")
	for {
		err := watchDockerEventsOnce(ctx)
		if ctx.Err() != nil {
			return
		}
		logger.Error("docker events stream disconnected - reconnecting", zap.Error(err))
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

func watchDockerEventsOnce(ctx context.Context) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	f := filters.NewArgs()
	f.Add("type", "container")
	for _, e := range []string{"start", "die", "destroy"} {
		f.Add("event", e)
	}

	msgCh, errCh := cli.Events(ctx, types.EventsOptions{Filters: f})
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			return err
		case msg := <-msgCh:
			go PublishEventWrapper(ctx, common.EventTypeContainerStateChanged, containerEventProperties(msg))
		}
	}
}

func containerEventProperties(msg events.Message) map[string]string {
	return map[string]string{
		common.PropertyTypeContainerID.Name:     msg.Actor.ID,
		common.PropertyTypeContainerName.Name:   msg.Actor.Attributes["name"],
		common.PropertyTypeContainerAction.Name: msg.Action,
	}
}
