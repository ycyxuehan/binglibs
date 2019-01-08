///////////////////////////////////////////////////////////
// docker.go
// docker api 
// ycyxuehan kun1.huang@outlook.com
//////////////////////////////////////////////////////////

package docker

import (
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/container"
	"encoding/base64"
	"strings"
	"encoding/json"
	"io/ioutil"
	"context"
	"github.com/docker/docker/api/types"
	"os"
	"fmt"
	"github.com/docker/docker/client"
)

//Docker type docker
type Docker struct {
	ConnURI string
	APIVersion string
	header map[string]string
	conn *client.Client
}

//New new a docker client
func New(URI string, version string)(*Docker, error){
	var d Docker
	if URI == "" {
		URI = "unix:///var/run/docker.sock"
	}
	d.ConnURI = URI
	d.APIVersion = version
	d.header = make(map[string]string)
	d.header["Content-Type"] = "application/tar"
	err := d.connect()
	return &d, err
}

//
func (d *Docker)connect()error{
	if d.conn != nil{
		d.conn.Close()
	}
	conn, err := client.NewClient(d.ConnURI, d.APIVersion, nil, d.header)
	d.conn = conn
	return err
}

//BuildImage build an image
func (d *Docker)BuildImage(opt BuildImageOptions)(*BuildImageResponse, error){
	if d.conn == nil {
		return nil, fmt.Errorf("can not connect to %s", d.ConnURI)
	}
	if opt.File == "" {
		return nil, fmt.Errorf("no file to build an image")
	}
	buildContext, err := os.Open(opt.File)
	if err != nil {
		return nil, fmt.Errorf("open docker file error: %s", err.Error())
	}
	defer buildContext.Close()
	options := types.ImageBuildOptions {
		SuppressOutput: true,
		Remove: true,
		PullParent: opt.PullParenet,
		Tags: opt.Tags,
	}
	buildResponse, err := d.conn.ImageBuild(context.Background(), buildContext, options)
	if err != nil {
		return nil, fmt.Errorf("build image error: %s", err.Error())
	}
	defer buildResponse.Body.Close()
	response, err := ioutil.ReadAll(buildResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("build image succeessed but get image info error: %s", err.Error())
	}
	var bir BuildImageResponse
	err = json.Unmarshal(response, &bir)
	if err != nil {
		return nil, fmt.Errorf("build image maybe failed, message: %s", string(response))
	}
	return &bir, nil
}

//PushImage push an image
func (d *Docker)PushImage(image string, user string, passwd string)([]string, error){
	if image == "" {
		return []string{}, fmt.Errorf("image tag is empty")
	}
	if d.conn == nil {
		return []string{}, fmt.Errorf("can not connect to %s", d.ConnURI)
	}
	opt := types.ImagePushOptions{}
	tags := strings.Split(image, "/")
	if len(tags) > 1 && user != "" && passwd != ""{
		authStr, err := GetAuthString(tags[0], user, passwd)
		if err != nil {
			return []string{}, fmt.Errorf("get auth string error")
		}
		opt.RegistryAuth = authStr
	}
	ctx := context.Background()
	out, err := d.conn.ImagePush(ctx, image, opt)
	if err != nil {
		return []string{}, err
	}
	defer out.Close()
	data, err := ioutil.ReadAll(out)
	return strings.Split(string(data), "\n"), nil
}

//GetContainer get a container
func (d *Docker)GetContainer(name string, all bool)(*types.Container, error){
	if d.conn == nil {
		return nil, fmt.Errorf("can not connect to %s", d.ConnURI)
	}

	containers, err := d.GetContainers(all)
	if err != nil {
		return nil, err
	}
	for _, container := range containers {
		name := container.Names[0]
		fmt.Println("name:", name, name)
		if name == "/"+name {
			return &container, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

//GetContainers get all container
func (d *Docker)GetContainers(all bool)([]types.Container, error){
	if d.conn == nil {
		return nil, fmt.Errorf("can not connect to %s", d.ConnURI)
	}
	clo := types.ContainerListOptions{
		All: all,
	}
	containers, err := d.conn.ContainerList(context.Background(), clo)
	return containers, err
}

//StartContainer start a container
func (d *Docker)StartContainer(ID string)error{
	if ID == "" {
		return fmt.Errorf("container ID is empty")
	}
	// c, err := d.GetContainer(name, true)
	// if err != nil {
	// 	return fmt.Errorf("cannot get the container %s: %s", name, err.Error())
	// }
	ctx := context.Background()
	opt := types.ContainerStartOptions{}
	err := d.conn.ContainerStart(ctx, ID, opt)
	return err
}
//StopContainer stop a container
func (d *Docker)StopContainer(ID string)error{
	if ID == "" {
		return fmt.Errorf("container name is empty")
	}
	// c, err := d.GetContainer(name, true)
	// if err != nil {
	// 	return fmt.Errorf("cannot get the container %s: %s", name, err.Error())
	// }
	ctx := context.Background()
	err := d.conn.ContainerStop(ctx, ID, nil)
	return err
}
//RestartContainer restart a container
func (d *Docker)RestartContainer(ID string)error{
	err := d.StopContainer(ID)
	if err != nil {
		return err
	}
	return d.StartContainer(ID)
}
//UpdateContainer update a container
func (d *Docker)UpdateContainer(ID string, opt UpdateOptions)error {
	ctx := context.Background()
	oomkilled := !opt.OomKilled
	conf := container.UpdateConfig{}
	conf.Memory = opt.Memory
	conf.MemorySwap = opt.Memory + 1
	// conf.MemorySwappiness = opt.
	conf.OomKillDisable = &oomkilled
	conf.RestartPolicy = container.RestartPolicy{
		Name:string(opt.RestartPolicy),
		// MaximumRetryCount: opt.RestartLimit,
	}
	if conf.RestartPolicy.IsOnFailure(){
		conf.RestartPolicy.MaximumRetryCount = opt.RestartLimit
	}
	_, err := d.conn.ContainerUpdate(ctx, ID, conf)
	if err != nil {
		return err
	}
	return nil
}
//CreateContainer Create a container
func (d *Docker)CreateContainer(opt *CreateContainerOptions)(string, error){
	var body container.ContainerCreateCreatedBody
	d.PullImage(opt.Image, opt.RegistryUser, opt.RegistryPassword)
	ctx := context.Background()
	var conf container.Config
	conf.Image = opt.Image
	conf.Env = opt.Env
	conf.Hostname = opt.HostName
	conf.AttachStdout = true
	conf.AttachStderr = true
	
	var hostConf container.HostConfig
	//volumes
	hostConf.Mounts = opt.Mounts
	//checkout volume src
	checkoutVolume(opt.Mounts)
	//expose
	hostConf.PortBindings = opt.PortBindings
	hostConf.LogConfig.Type = "journald"
	hostConf.IpcMode = ""
	// hostConf.Runtime = "docker-runc"
	hostConf.Devices = []container.DeviceMapping{}
	conf.ExposedPorts = opt.ExposedPorts
	// hostConf.Binds = opt.VolumeBinds
	//hosts
	hostConf.ExtraHosts = opt.ExtraHosts
	//resource
	hostConf.Memory = opt.Memory
	hostConf.OomKillDisable = &opt.OomKillDisable
	hostConf.MemorySwappiness = &opt.Swappiness
	hostConf.MemorySwap = opt.Memory + 1
	hostConf.RestartPolicy = container.RestartPolicy{Name:opt.RestartPolicy,}
		if hostConf.RestartPolicy.IsOnFailure(){
		hostConf.RestartPolicy.MaximumRetryCount = opt.RestartRetryCount
	}
	body, err := d.conn.ContainerCreate(ctx, &conf, &hostConf, nil, opt.Name)
	if err != nil{		
		return "", fmt.Errorf("create container error: %s", err.Error())
	}
	return body.ID, nil
}

//GetContainerByID get container by id
func (d *Docker)GetContainerByID(ID string, all bool)(*types.Container, error){
	if ID == "" {
		return nil, fmt.Errorf("container ID is empty")
	}
	containers, err := d.GetContainers(all)
	if err != nil {
		return nil, err
	}
	for _, c := range containers {
		if c.ID == ID {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
//GetContainerByImage get container by image
func (d *Docker)GetContainerByImage(img string, all bool)(*types.Container, error){
	if img == "" {
		return nil, fmt.Errorf("not found")
	}
	containers, err := d.GetContainers(all)
	if err != nil {
		return nil, err
	}
	for _, c := range containers {
		if c.Image == img  {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

//Close close connection
func (d *Docker)Close(){
	d.conn.Close()
}

//RemoveContainer remove a container
func (d *Docker)RemoveContainer(ID string, force bool)error{
	if ID == "" {
		return fmt.Errorf("id is empty")
	}
	opt := types.ContainerRemoveOptions{
		Force:force,
	}
	err := d.conn.ContainerRemove(context.Background(), ID, opt)
	return err
}

func checkoutVolume(mounts []mount.Mount){
	for _, m := range mounts {
		if _, err := os.Stat(m.Source); err != nil || os.IsNotExist(err) {
			os.MkdirAll(m.Source, os.ModePerm)
		}
	}
}

//Tag tag an image
func (d *Docker)Tag(old string, new string)error{
	err := d.conn.ImageTag(context.Background(), old, new)
	return err 
}

//Pull pull an image
func (d *Docker)PullImage(tag string, username string, password string)([]string, error){
	pullOpt := types.ImagePullOptions{}
	tags := strings.Split(tag, "/")
	if len(tags) > 1 && username != "" && password != ""{
		authStr, err := GetAuthString(tags[0], username, password)
		if err != nil {
			return []string{}, fmt.Errorf("get auth string error")
		}
		pullOpt.RegistryAuth = authStr
	}
	response, err := d.conn.ImagePull(context.Background(), tag,  pullOpt)
	if err != nil {
		return []string{}, err
	}
	defer response.Close()
	data, err := ioutil.ReadAll(response)
	return strings.Split(string(data), "\n"), nil
}

func GetAuthString(server string, username string, password string)(string, error){
	if server == "" {
		return "", nil
	}
	authConf := types.AuthConfig {}
	authConf.Username = username
	authConf.Password = password
	authConf.ServerAddress = server
	encodedJSON, err := json.Marshal(authConf)
	if err != nil {
		return "", fmt.Errorf("convert login string error")
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	return authStr, nil
}