package main

import (
    "os"
    "fmt"
    "io/ioutil"
    "path/filepath"

    "gopkg.in/yaml.v2"

    "github.com/mitchellh/cli"
    "github.com/mitchellh/go-homedir"
    "github.com/fsouza/go-dockerclient"
)

type project struct {
    Name string
    URL string
    Version string
    Container string
}

var castleRoot = "/Users/hartleym/tmp/movingcastle"

var dirs = []string{"data", "output", "working"}

func (p project) dataPath() string {
    return filepath.Join(castleRoot, p.Name, "data")
}

func (p project) codePath() string {
    return filepath.Join(castleRoot, p.Name, "code")
}

func (p project) outputPath() string {
    return filepath.Join(castleRoot, p.Name, "output")
}

func init() {
    envCastleRoot := os.Getenv("MOVINGCASTLE")
    if envCastleRoot != "" {
        castleRoot = envCastleRoot
    } else {
        home, _ := homedir.Dir()
        castleRoot = filepath.Join(home, "movingcastle")
    }
}

func deploy() {

     //endpoint := "unix:///var/run/docker.sock"
    client, _ := docker.NewClientFromEnv()
    // imgs, _ := client.ListImages(docker.ListImagesOptions{All: false})
    // for _, img := range imgs {
    //     fmt.Println("ID: ", img.ID)
    //     fmt.Println("RepoTags: ", img.RepoTags)
    //     fmt.Println("Created: ", img.Created)
    //     fmt.Println("Size: ", img.Size)
    //     fmt.Println("VirtualSize: ", img.VirtualSize)
    //     fmt.Println("ParentId: ", img.ParentID)
    // }

    newmount := docker.Mount{Source: "/Users/hartleym/tmp", Destination: "/test"}
    mounts := []docker.Mount{newmount}

    binds := []string{"/Users/hartleym/tmp/:/test"}

    //cmd := []string{"ls", "/test"}
    cmd := []string{"git", "clone", "-c", "http.sslVerify=false", 
        "https://github.com/JIC-Image-Analysis/yeast_growth", "/test/yeast_growth"}

    config := docker.Config{Cmd: cmd, Image: "movingcastle", Mounts: mounts}
    hostconfig := docker.HostConfig{Binds: binds}
    container, err := client.CreateContainer(docker.CreateContainerOptions{Name: "executor", Config: &config})
    if err != nil {
        panic("arg")
    }
    client.StartContainer(container.ID, &hostconfig)

    ret, waiterr := client.WaitContainer(container.ID)
    if waiterr != nil {
        panic("waiterr")
    }
    fmt.Println("Returned: ", ret)

    logoptions := docker.LogsOptions{Container: container.ID, Stdout: true, 
        Stderr: true, OutputStream: os.Stdout, ErrorStream: os.Stderr}
    err = client.Logs(logoptions)
    if err != nil {
        fmt.Println("Err:, ", err)
    }

    client.RemoveContainer(docker.RemoveContainerOptions{ID: container.ID})
   
}

func mkProjectDirs(projectName string) {

    projectRoot := filepath.Join(castleRoot, projectName)
    
    for _, dir := range dirs {
         os.MkdirAll(filepath.Join(projectRoot, dir), 0777)
    }   
}

func initDockerClient() (*docker.Client, error) {
    var client *docker.Client
    var err error

    if os.Getenv("DOCKER_HOST") != "" {
        client, err = docker.NewClientFromEnv()
    } else {
        endpoint := "unix:///var/run/docker.sock"
        client, err = docker.NewClient(endpoint)
    }

    return client, err
}

func cloneProject(proj project) {

    projectRoot := filepath.Join(castleRoot, proj.Name)

    client, _ := initDockerClient()
    newmount := docker.Mount{Source: projectRoot, Destination: "/deploy"}
    mounts := []docker.Mount{newmount}

    bindString := fmt.Sprintf("%s:/deploy", projectRoot)
    binds := []string{bindString}

    cmd := []string{"git", "clone", proj.URL, "/deploy/code"}

    config := docker.Config{Cmd: cmd, Image: "movingcastle", Mounts: mounts}
    hostconfig := docker.HostConfig{Binds: binds}
    container, err := client.CreateContainer(docker.CreateContainerOptions{Name: "executor", Config: &config})
    if err != nil {
        panic("arg")
    }

    runContainer(client, hostconfig, container)
    // client.StartContainer(container.ID, &hostconfig)

    // ret, waiterr := client.WaitContainer(container.ID)
    // if waiterr != nil {
    //     panic("waiterr")
    // }
    // //fmt.Println("Returned: ", ret)

    // logoptions := docker.LogsOptions{Container: container.ID, Stdout: true, 
    //     Stderr: true, OutputStream: os.Stdout, ErrorStream: os.Stderr}
    // err = client.Logs(logoptions)
    // if err != nil {
    //     fmt.Println("Err:, ", err)
    // }

    // client.RemoveContainer(docker.RemoveContainerOptions{ID: container.ID})   
}

func runContainer(client *docker.Client, hostconfig docker.HostConfig, container *docker.Container) int {
    client.StartContainer(container.ID, &hostconfig)

    ret, waiterr := client.WaitContainer(container.ID)
    if waiterr != nil {
        panic("waiterr")
    }

    logoptions := docker.LogsOptions{Container: container.ID, Stdout: true, 
        Stderr: true, OutputStream: os.Stdout, ErrorStream: os.Stderr}
    err := client.Logs(logoptions)
    if err != nil {
        fmt.Println("Err:, ", err)
    }

    client.RemoveContainer(docker.RemoveContainerOptions{ID: container.ID}) 

    return ret
}

func runProjectCode(proj project) {

    client, _ := initDockerClient()

    cmd := []string{"python", "/code/scripts/nikonE800_annotate.py", "/data/image-01.jpg", "/output"}

    dataBind := fmt.Sprintf("%s:/data", proj.dataPath())
    codeBind := fmt.Sprintf("%s:/code", proj.codePath())
    outputBind := fmt.Sprintf("%s:/output", proj.outputPath())
    binds := []string{dataBind, codeBind, outputBind}

    //config := docker.Config{Cmd: cmd, Image: proj.Container, Mounts: mounts}
    config := docker.Config{Cmd: cmd, Image: proj.Container}
    hostconfig := docker.HostConfig{Binds: binds}
    container, err := client.CreateContainer(docker.CreateContainerOptions{Name: "executor", Config: &config})
    if err != nil {
        panic("arg")
    }

    runContainer(client, hostconfig, container)

    //bindString := fmt.Sprintf()

}

func deployProject(proj project) {
    mkProjectDirs(proj.Name)
    cloneProject(proj)   
}

type deployCommand struct {
    Ui      cli.Ui
}

func (c *deployCommand) Run(_ []string) int {
    var p project
    dat, _ := ioutil.ReadFile("project.yml")
    yaml.Unmarshal(dat, &p)

    deployProject(p)

    return 0
}

func (c *deployCommand) Help() string {
    return "Deploy project locally"
}

func (c *deployCommand) Synopsis() string {
    return "Deploy things"
}

func projFromYAML(yamlFile string) project {

    var p project
    dat, _ := ioutil.ReadFile(yamlFile)
    yaml.Unmarshal(dat, &p)

    return p
}


type runCommand struct {
    Ui      cli.Ui
}

func (c *runCommand) Run(_ []string) int {
    p := projFromYAML("project.yml")

    runProjectCode(p)

    return 0
}

func (c *runCommand) Help() string {
    return "Run analysis for project"
}

func (c *runCommand) Synopsis() string {
    return "Run things"
}

func main() {

    // myproj := project{Name: "yeast_growth", URL: "https://github.com/JIC-Image-Analysis/yeast_growth", version: "0.1.0"}
    // d, err := yaml.Marshal(&myproj)

    // if err != nil {
    //     panic(err)
    // }

    // fmt.Printf("%v", myproj)

    // fmt.Printf("%v", string(d))

    // os.Getenv("MOVINGCASTLE")

    c := cli.NewCLI("calcifer", "0.1.0")
    c.Args = os.Args[1:]
    c.Commands = map[string]cli.CommandFactory{
        "deploy" : func() (cli.Command, error) {
            return &deployCommand{}, nil
        },
        "run" : func() (cli.Command, error) {
            return &runCommand{}, nil
        },
    }

    exitStatus, err := c.Run()
    if err != nil {
        fmt.Fprintln(os.Stderr, err.Error())
    }

    os.Exit(exitStatus)

    //runProjectCode(p)

    //fmt.Println("Created ID: ", container.ID)

}
