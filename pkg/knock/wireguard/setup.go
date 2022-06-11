package wireguard

import (
	"bytes"
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl"
	"io/ioutil"
	"k8s.io/klog/v2"
	"os"
	"os/exec"
)

func SetUpWireguardLink(netifName string, conf *WgQuickConf) error {
	err := ioutil.WriteFile(netifName+".conf", conf.Parse(), 0644)
	if err != nil {
		klog.Errorf("write %s.conf failed with err: %v", netifName, err)
		return err
	}

	cmd := exec.Command("wg-quick", "up", netifName+".conf")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Run()
	if err != nil {
		klog.Warningf("wg-quick up %s failed with err: %v", netifName, err)
	}
	klog.Infof("wg-quick up %s exec stdout: %s", netifName, stdout.String())
	klog.Infof("wg-quick up %s exec stdout: %s", netifName, stdout.String())

	if err != nil {
		return err
	}

	wgCli, err := wgctrl.New()
	if err != nil {
		klog.Warningf("try to ensure wg link has been setup but failed with err: %s", err)
	}
	_, err = wgCli.Device(netifName)
	if err == nil {
		return nil
	}

	if err == os.ErrNotExist {
		return fmt.Errorf("wg link %s should up but failed", netifName)
	}
	klog.Warningf("try to ensure wg link has been shutdown but failed with err: %s", err)
	return err
}
