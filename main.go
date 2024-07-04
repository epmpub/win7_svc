package main

import (
	"WinHelper/tools"
	"fmt"
	"log"
	"os"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

type myService struct{}

func (m *myService) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	s <- svc.Status{State: svc.StartPending}
	s <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	elog, err := eventlog.Open("MyService")
	if err != nil {
		return false, 1
	}
	defer elog.Close()

	elog.Info(1, "MyService started")

loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				s <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				break loop
			default:
				elog.Warning(1, "unexpected control request")
			}
		case <-time.After(1 * time.Second):
			// Perform your service's work here
			WriteTimeToFile()
			host, err := os.Hostname()
			if err != nil {
				fmt.Println("hostname get error")
			}
			msg := "Log,Windows7," + host + "," + time.Now().Format(time.RFC3339)
			tools.Http_POST(msg)
			time.Sleep(5 * time.Second)
		}
	}
	s <- svc.Status{State: svc.StopPending}
	return false, 0
}

func setServiceDescription(serviceName, description string) error {
	// Connect to the service manager
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	// Open the service
	s, err := m.OpenService(serviceName)
	if err != nil {
		return err
	}
	defer s.Close()

	// Set the service description
	sd := windows.SERVICE_DESCRIPTION{
		Description: windows.StringToUTF16Ptr(description),
	}

	err = windows.ChangeServiceConfig2(
		s.Handle,
		windows.SERVICE_CONFIG_DESCRIPTION,
		(*byte)(unsafe.Pointer(&sd)),
	)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	isInteractive, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	}
	if !isInteractive {
		runService("WinHelpService", false)
		return
	}

	if len(os.Args) < 2 {
		log.Println("Usage: WinHelpService <install | uninstall>")
		return
	}
	cmd := os.Args[1]
	switch cmd {
	case "install":
		installService("WinHelper", "Windows Helper Service")
		setServiceDescription("WinHelper", "Provide system security extension service for windows 7")

	case "uninstall":
		removeService("WinHelpService")
	default:
		log.Println("Usage: WinHelpService <install | uninstall>")
	}
}

func runService(name string, isDebug bool) {
	err := svc.Run(name, &myService{})
	if err != nil {
		log.Fatalf("failed to run service: %v", err)
	}
}

func installService(name, desc string) {
	m, err := mgr.Connect()
	if err != nil {
		log.Fatalf("could not connect to service manager: %v", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err == nil {
		s.Close()
		log.Fatalf("service %s already exists", name)
	}

	exepath, err := os.Executable()
	if err != nil {
		log.Fatalf("could not get executable path: %v", err)
	}

	s, err = m.CreateService(name, exepath, mgr.Config{DisplayName: desc, StartType: mgr.StartAutomatic})
	if err != nil {
		log.Fatalf("could not create service: %v", err)
	}
	defer s.Close()

	err = eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		log.Fatalf("could not setup event log source: %v", err)
	}

	log.Printf("service %s installed", name)
}

func removeService(name string) {
	m, err := mgr.Connect()
	if err != nil {
		log.Fatalf("could not connect to service manager: %v", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		log.Fatalf("could not access service: %v", err)
	}
	defer s.Close()

	err = s.Delete()
	if err != nil {
		log.Fatalf("could not delete service: %v", err)
	}

	err = eventlog.Remove(name)
	if err != nil {
		log.Fatalf("could not remove event log source: %v", err)
	}

	log.Printf("service %s removed", name)
}
