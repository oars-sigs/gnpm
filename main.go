package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	pkgFile = "gnpm.json"
	modDir  = "src/gcomponents"
)

func main() {
	dir := "src/gcomponents"
	os.MkdirAll(dir, 0755)
	_, err := os.Stat(pkgFile)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.WriteFile(pkgFile, []byte("{}"), 0664)
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}
		} else {
			fmt.Println(err)
			os.Exit(-1)
		}
	}
	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "install":
			err = installAll()
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}
		default:
			fmt.Println("无效命令")
			os.Exit(-1)
		}
		return
	}
	if len(os.Args) == 3 {
		err = installOne(os.Args[2], "")
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		return
	}
	if len(os.Args) == 4 {
		err = installOne(os.Args[2], os.Args[3])
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		return
	}
	fmt.Println("无效命令")
	os.Exit(-1)
}

type pkgDef struct {
	Pkgs []pkgItem `json:"pkgs"`
}

type pkgItem struct {
	Repo   string `json:"repo"`
	Target string `json:"target"`
}

func parsePkg() (*pkgDef, error) {
	data, err := os.ReadFile(pkgFile)
	if err != nil {
		return nil, err
	}
	var p pkgDef
	err = json.Unmarshal(data, &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func genPkg(p *pkgDef) error {
	b, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(pkgFile, b, 0664)
}

func installAll() error {
	p, err := parsePkg()
	if err != nil {
		return err
	}
	for _, pkg := range p.Pkgs {
		err = installOne(pkg.Repo, pkg.Target)
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func installOne(pkg, target string) error {
	os.MkdirAll(modDir, 0755)
	tmpDir := os.TempDir() + "/" + time.Now().Format("20060102150405")
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)
	ss := strings.Split(pkg, "@")
	if len(ss) != 2 {
		return errors.New(pkg + "格式错误")
	}
	pss := strings.Split(ss[0], "/")
	if len(pss) < 5 {
		return errors.New(pkg + "格式错误")
	}
	reg := strings.Join(pss[:5], "/")
	fmt.Println(reg)
	cmd := exec.Command("git", "clone", "--depth", "1", "-b", ss[1], "--single-branch", reg+".git")
	cmd.Dir = tmpDir
	cmdRes, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(cmdRes))
	}
	dpaths := pss[4:]
	pdir := strings.Join(dpaths, "/")
	fmt.Println(pdir)
	dpath := modDir + "/" + pdir
	if len(target) == 0 {
		target = pdir
	} else {

		dpath = modDir + "/" + target
	}
	os.RemoveAll(dpath)
	fmt.Println(tmpDir+"/"+pdir, dpath, 22222)

	err = os.Rename(tmpDir+"/"+pdir, dpath)
	if err != nil {
		return err
	}
	p, err := parsePkg()
	if err != nil {
		return err
	}
	exist := false
	for i, pp := range p.Pkgs {
		vss := strings.Split(pp.Repo, "@")
		if vss[0] == ss[0] {
			p.Pkgs[i] = pkgItem{
				Repo:   pkg,
				Target: target,
			}
			exist = true
			break
		}
	}
	if !exist {
		p.Pkgs = append(p.Pkgs, pkgItem{
			Repo:   pkg,
			Target: target,
		})
	}
	err = genPkg(p)
	if err != nil {
		return err
	}
	return nil
}
