package main

import (
    "os"
    "fmt"

    "github.com/fsouza/go-dockerclient"
)

func main() {
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
    cmd := []string{"git", "clone", "https://github.com/JIC-Image-Analysis/yeast_growth", "/test/yeast_growth"}

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

    logoptions := docker.LogsOptions{Container: container.ID, Stdout: true, Stderr: true, OutputStream: os.Stdout, ErrorStream: os.Stderr}
    err = client.Logs(logoptions)
    if err != nil {
        fmt.Println("Err:, ", err)
    }

    client.RemoveContainer(docker.RemoveContainerOptions{ID: container.ID})

    //fmt.Println("Created ID: ", container.ID)

}
