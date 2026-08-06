package service

import (
	"testing"

	"github.com/NimoTech/NimoOS-AppManagement/common"
	"github.com/docker/docker/api/types/events"
	"github.com/stretchr/testify/assert"
)

func TestContainerEventProperties(t *testing.T) {
	for _, action := range []string{"start", "die", "destroy"} {
		msg := events.Message{
			Action: action,
			Actor:  events.Actor{ID: "abc123", Attributes: map[string]string{"name": "tasklist"}},
		}
		props := containerEventProperties(msg)
		assert.Equal(t, "abc123", props[common.PropertyTypeContainerID.Name])
		assert.Equal(t, "tasklist", props[common.PropertyTypeContainerName.Name])
		assert.Equal(t, action, props[common.PropertyTypeContainerAction.Name])
	}
}
