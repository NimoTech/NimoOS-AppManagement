package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainerStateChangedRegistered(t *testing.T) {
	assert.Equal(t, "docker:container:state-changed", EventTypeContainerStateChanged.Name)
	assert.Equal(t, "docker:container:action", PropertyTypeContainerAction.Name)

	found := false
	for _, et := range EventTypes {
		if et.Name == EventTypeContainerStateChanged.Name {
			found = true
		}
	}
	assert.True(t, found, "EventTypeContainerStateChanged must be in EventTypes for startup registration")
}
