package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed orca-icon.icns
var iconData []byte

func iconPath(app string) (string, error) {
	out, err := exec.Command("/usr/bin/defaults", "read", filepath.Join(app, "Contents", "Info"), "CFBundleIconFile").Output()
	if err != nil {
		return "", fmt.Errorf("read app icon name: %w", err)
	}
	name := strings.TrimSpace(string(out))
	if name == "" || filepath.Base(name) != name {
		return "", fmt.Errorf("invalid app icon name %q", name)
	}
	if filepath.Ext(name) == "" {
		name += ".icns"
	}
	return filepath.Join(app, "Contents", "Resources", name), nil
}
func iconStatus(app string) string {
	path, err := iconPath(app)
	if err != nil {
		return err.Error()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err.Error()
	}
	if bytes.Equal(data, iconData) {
		return "Avari icon installed"
	}
	return "original or other icon"
}
func changeIcon(app string, revert bool) error {
	path, err := iconPath(app)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	backupPath := path + ".avari-original"
	var next []byte
	if revert {
		if !bytes.Equal(data, iconData) {
			return fmt.Errorf("current icon is not the Avari icon; left unchanged")
		}
		next, err = os.ReadFile(backupPath)
		if err != nil {
			return fmt.Errorf("read original icon: %w", err)
		}
	} else {
		if bytes.Equal(data, iconData) {
			return nil
		}
		f, e := os.OpenFile(backupPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm())
		if e != nil {
			return fmt.Errorf("create icon backup (restore or remove existing backup first): %w", e)
		}
		if _, e = f.Write(data); e != nil {
			f.Close()
			os.Remove(backupPath)
			return e
		}
		if e = f.Close(); e != nil {
			os.Remove(backupPath)
			return e
		}
		next = iconData
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".avari-icon-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err = f.Chmod(info.Mode().Perm()); err == nil {
		_, err = f.Write(next)
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err = os.Rename(tmp, path); err != nil {
		return err
	}
	if revert {
		if err = os.Remove(backupPath); err != nil {
			return err
		}
	}
	return nil
}
