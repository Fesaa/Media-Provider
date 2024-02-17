package mount

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

var mounted bool = false

// Tries mounting the network drive.
// If reset is true, it will try unmounting the drive first.
// If the (un)mount fails, the program will exit.
func Mount(reset bool) {
	if !doMount {
		return
	}
	if mounted && reset {
		Unmount()
	}

	switch stdout, err := exec.CommandContext(context.TODO(),
		"mount",
		"-t",
		"cifs",
		"-o",
		"username="+user+",password="+pass+",domain="+domain,
		"//"+url,
		"/app/mount").Output(); err.(type) {
	case nil:
		mounted = true
		slog.Info("Mount successful" + string(stdout))
		break
	case (*exec.ExitError):
		err := err.(*exec.ExitError)
		slog.Error(fmt.Sprintf("Mount failed with exit code: %d and message\n %s", err.ExitCode(), string(err.Stderr)))
		slog.Info("No mount breaks the program. Exiting")
		os.Exit(1)
	default:
		slog.Error(fmt.Sprintf("Mount failed with unknown error: %s", err.Error()))
		slog.Info("No mount breaks the program. Exiting")
		os.Exit(1)
	}

	mounted = true
}

// Tries unmounting the network drive.
// If the unmount fails, the program will exit.
func Unmount() {
	if !mounted || !doMount {
		return
	}

	switch stdout, err := exec.CommandContext(context.TODO(),
		"umount",
		"/app/mount").Output(); err.(type) {
	case nil:
		mounted = true
		slog.Info("Umount successful" + string(stdout))
		break
	case (*exec.ExitError):
		err := err.(*exec.ExitError)
		slog.Error(fmt.Sprintf("Umount failed with exit code: %d and message\n %s", err.ExitCode(), string(err.Stderr)))
		slog.Info("This leaves the program in a potentially broken state. Exiting")
		os.Exit(1)
	default:
		slog.Error(fmt.Sprintf("Umount failed with unknown error: %s", err.Error()))
		slog.Info("This leaves the program in a potentially broken state. Exiting")
		os.Exit(1)
	}
}
