package wireguard

import (
	"bytes"
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl"
	"k8s.io/klog/v2"
	"os"
	"os/exec"
)

func ShutdownWgQuickLink(netifName string) error {
	cmd := exec.Command("wg-quick", "down", netifName+".conf")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	if err != nil {
		klog.Warningf("wg-quick down %s failed with err: %v", netifName, err)
	}
	klog.Infof("wg-quick down %s exec stdout: %s", netifName, stdout.String())
	klog.Infof("wg-quick down %s exec stdout: %s", netifName, stdout.String())

	if err != nil {
		return err
	}

	wgCli, err := wgctrl.New()
	if err != nil {
		klog.Warning("try to ensure wg link has been shutdown but failed with err: %s", err)
	}
	_, err = wgCli.Device(netifName)
	if err == os.ErrNotExist {
		return nil
	}
	if err == nil {
		return fmt.Errorf("wg link %s should down but failed", netifName)
	}
	klog.Warning("try to ensure wg link has been shutdown but failed with err: %s", err)
	return err
}
