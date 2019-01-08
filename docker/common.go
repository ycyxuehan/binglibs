///////////////////////////////////////////////////////////
// docker.go
// docker api 
// ycyxuehan kun1.huang@outlook.com
//////////////////////////////////////////////////////////

package docker

import (
	"github.com/docker/go-connections/nat"
	"github.com/docker/docker/api/types/mount"
	"strings"

)
//RestartPolicy restart proxy
type RestartPolicy string

const (
	//Always Always
	Always RestartPolicy = "always"
	//OnFailure OnFailure
	OnFailure RestartPolicy = "on-failure"
	//UnlessStopped UnlessStopped
	UnlessStopped RestartPolicy = "unless-stopped"
	//No no
	No RestartPolicy = "no"
)

//GetRestartPolicy get restart policy by string
func GetRestartPolicy(policy string)RestartPolicy{
	switch strings.ToLower(policy) {
	case "always": return Always
	case "on-failure": return OnFailure
	case "unless-stopped": return UnlessStopped
	case "no": return No
	default: return No
	}
}

//BuildImageOptions build image options
type BuildImageOptions struct {
	File string
	Tags [] string
	PullParenet bool
}
//BuildImageResponse build image response info
type BuildImageResponse struct {
	ImageHash string `json:"stream"`
}
//GetImageHash get image hash
func (b *BuildImageResponse)GetImageHash()string{
	if b.ImageHash == "" {
		return ""
	}
	hash := strings.Replace(b.ImageHash, "sha256:", "", -1)
	hash = strings.Replace(hash, "\n", "", -1)
	return hash
}

//UpdateOptions update container options
type UpdateOptions struct {
	RestartPolicy RestartPolicy `json:"RestartPolicy"`
	RestartLimit int `json:"RestartLimit"`
	Memory int64 `json:"Memory"`
	OomKilled bool `json:"OomKilled"`
}

//CreateContainerOptions the options to create a container
type CreateContainerOptions struct {
	Name string
	Image string
	Env []string
	HostName string
	Mounts []mount.Mount
	PortBindings nat.PortMap
	ExposedPorts nat.PortSet
	ExtraHosts [] string
	VolumeBinds []string
	Memory int64
	Swappiness int64
	OomKillDisable bool
	RestartPolicy string
	RestartRetryCount int
	Registry string
	RegistryUser string
	RegistryPassword string
}
//NewCreateContainerOptions new a CreateContainerOption
func NewCreateContainerOptions()*CreateContainerOptions{
	var c CreateContainerOptions
	c.Mounts = []mount.Mount{}
	c.PortBindings = nat.PortMap{}
	c.ExposedPorts = nat.PortSet{}
	c.Memory = 4096000
	return &c
}
