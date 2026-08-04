package model

import (
	"time"
)

type ServerAppListCollection struct {
	List      []ServerAppList `json:"list"`
	Recommend []ServerAppList `json:"recommend"`
	Community []ServerAppList `json:"community"`
}

type StateEnum int

const (
	StateEnumNotInstalled StateEnum = iota
	StateEnumInstalled
)

// @tiger - for response data structures, static info (e.g. title) and
//
//	dynamic info (e.g. state, query_count) should be split into separate structs.
//
//	Benefits:
//	1 - fetching dynamic info repeatedly stays cheap, since static info only needs to be fetched once
//	2 - lower maintenance cost going forward (keeping all fields flattened at one level costs more to maintain)
//
//	Also, some app-type-specific fields (e.g. Docker-related ones) could be kept in a map.
//	That way, adding polymorphic app types in the future (e.g. Snap) wouldn't require maintaining
//	multiple structs or a struct carrying unnecessary fields.
type ServerAppList struct {
	ID             uint      `gorm:"column:id;primary_key" json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Tagline        string    `json:"tagline"`
	Tags           Strings   `gorm:"type:json" json:"tags"`
	Icon           string    `json:"icon"`
	ScreenshotLink Strings   `gorm:"type:json" json:"screenshot_link"`
	Category       string    `json:"category"`
	CategoryID     int       `json:"category_id"`
	CategoryFont   string    `json:"category_font"`
	PortMap        string    `json:"port_map"`
	ImageVersion   string    `json:"image_version"`
	Tip            string    `json:"tip"`
	Envs           EnvArray  `json:"envs"`
	Ports          PortArray `json:"ports"`
	Volumes        PathArray `json:"volumes"`
	Devices        PathArray `json:"devices"`
	NetworkModel   string    `json:"network_model"`
	Image          string    `json:"image"`
	Index          string    `json:"index"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	State          StateEnum `json:"state"`
	Author         string    `json:"author"`
	MinMemory      int       `json:"min_memory"`
	MinDisk        int       `json:"min_disk"`
	Thumbnail      string    `json:"thumbnail"`
	Healthy        string    `json:"healthy"`
	Plugins        Strings   `json:"plugins"`
	Origin         string    `json:"origin"`
	Type           int       `json:"type"`
	QueryCount     int       `json:"query_count"`
	Developer      string    `json:"developer"`
	HostName       string    `json:"host_name"`
	Privileged     bool      `json:"privileged"`
	CapAdd         Strings   `json:"cap_add"`
	Cmd            Strings   `json:"cmd"`
	Architectures  Strings   `json:"architectures"`
	LatestDigest   Strings   `json:"latest_digests"`
}

type MyAppList struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Icon           string `json:"icon"`
	State          string `json:"state"`
	CustomID       string `gorm:"column:custom_id;primary_key" json:"custom_id"`
	Index          string `json:"index"`
	Port           string `json:"port"`
	Slogan         string `json:"slogan"`
	Type           string `json:"type"`
	Image          string `json:"image"`
	Volumes        string `json:"volumes"`
	Latest         bool   `json:"latest"`
	Host           string `json:"host"`
	Protocol       string `json:"protocol"`
	Created        int64  `json:"created"`
	AppStoreID     uint   `json:"appstore_id"`
	IsUncontrolled bool   `json:"is_uncontrolled"`
}

type Ports struct {
	ContainerPort uint   `json:"container_port"`
	CommendPort   int    `json:"commend_port"`
	Desc          string `json:"desc"`
	Type          int    `json:"type"` //  1:required 2:optional 3:default value need not be shown 4:system-handled  5:container content is also editable
}

type Volume struct {
	ContainerPath string `json:"container_path"`
	Path          string `json:"path"`
	Desc          string `json:"desc"`
	Type          int    `json:"type"` //  1:required 2:optional 3:default value need not be shown 4:system-handled   5:container content is also editable
}

type Envs struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Desc  string `json:"desc"`
	Type  int    `json:"type"` //  1:required 2:optional 3:default value need not be shown 4:system-handled 5:container content is also editable
}

type Devices struct {
	ContainerPath string `json:"container_path"`
	Path          string `json:"path"`
	Desc          string `json:"desc"`
	Type          int    `json:"type"` //  1:required 2:optional 3:default value need not be shown 4:system-handled 5:container content is also editable
}

type Strings []string

type MapStrings []map[string]string
