package netrun

import (
	"fmt"
	"github.com/ingenierias-lentas/netrun/server"
	"github.com/opencontainers/runc/libcontainer"
	_ "github.com/opencontainers/runc/libcontainer/nsenter"
	log "github.com/sirupsen/logrus"
	"os"
	"runtime"
)

func init() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		runtime.GOMAXPROCS(1)
		runtime.LockOSThread()
		factory, _ := libcontainer.New("")
		if err := factory.StartInitialization(); err != nil {
			log.Fatal(err)
		}
		panic("--this line should never have been executed, congratulations--")
	}
}

/* Ways to interact with a running container
// return all the pids for all processes running inside the container
processes, err := container.Processes()

// get detailed cpu, memory, io, and network statistics for the container and
// it's processes
stats, err := container.Stats()

// pause all processes inside the container,
container.Pause()

// resume all paused processes.
container.Resume()

// send signal to container's init process.
container.Signal(signal)

// update container resource constraints
container.Set(config)

// get current status of the container
status, err := container.Status()

// get current container's state information
state, err := container.State()
*/
func main() {
	portString := ":8021"
	fmt.Printf("Initializing server at port %s\n", portString)
	router := server.InitServer()
	server.RunServer(portString, router)
	/*
		fmt.Printf("Running container for netrun-test\n")

		containerRoot := "/var/lib/container"
		factory, err := libcontainer.New(
			containerRoot,
			libcontainer.Cgroupfs,
			libcontainer.InitArgs(os.Args[0], "init"),
		)
		if err != nil {
			log.Fatal(err)
			return
		}

		config := DefaultContainerConfig

		containerId := "test-container-id"
		var container libcontainer.Container
		if _, err = os.Stat(containerRoot + "/" + containerId); err == nil {
			fmt.Printf("Loading existing container with id: %s\n", containerId)
			container, err = factory.Load(containerId)
		} else {
			fmt.Printf("Creating new container with id: %s\n", containerId)
			container, err = factory.Create(containerId, config)
		}

		if err != nil {
			log.Fatal(err)
			return
		}

		process := &libcontainer.Process{
			Args:   []string{"/bin/bash", "-it"},
			Env:    []string{"PATH=/bin"},
			User:   "daemon",
			Stdin:  os.Stdin,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Init:   true,
		}

		err = container.Run(process)
		if err != nil {
			container.Destroy()
			log.Fatal(err)
			return
		}

		// wait for the process to finish
		_, err = process.Wait()
		if err != nil {
			log.Fatal(err)
		}

		// destroy the container
		container.Destroy()
	*/
}
