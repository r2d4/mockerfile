package mockerfile2llb

import (
	"fmt"
	"strings"

	"github.com/moby/buildkit/client/llb"
	"github.com/r2d4/mockerfile/pkg/mocker/config"
)

func curl() llb.State {
	return llb.Image("docker.io/library/alpine:3.6").
		Run(llb.Shlex("apk add --no-cache curl")).Root()
}

func packages(base llb.State, p *config.Package) llb.State {
	base = base.Run(shf("apt-get update && apt-get install apt-transport-https -y")).Root()
	for _, repo := range p.Repo {
		base = base.Run(shf("echo \"%s\" >> /etc/apt/sources.list", repo)).Root()
	}
	for _, key := range p.Gpg {
		base = aptAddKey(base, key)
	}
	if len(p.Install) > 0 {
		packages := strings.Join(p.Install, " ")
		base = base.Run(shf("apt-get update && apt-get install --no-install-recommends --no-install-suggests -y %s", packages)).Root()
	}
	return base
}

func Mockerfile2LLB(c *config.Config) (llb.State, *Image) {
	current := c.Images[1]
	s := llb.Image(current.From)
	if current.Package != nil {
		s = packages(s, current.Package)
	}
	for _, e := range current.ExternalFiles {
		downloaded := external(e)
		s = copy(downloaded, e.Destination, s, e.Destination)
	}
	imageCfg := NewImageConfig(c)
	return s, imageCfg
}

func aptAddKey(dst llb.State, url string) llb.State {
	downloadSt := curl().
		Run(llb.Shlexf("curl -Lo /key.gpg %s", url)).Root()
	dst = copy(downloadSt, "/key.gpg", dst, "/key.gpg")
	return dst.
		Run(sh("apt-key add /key.gpg && rm /key.gpg")).
		Root()
}

func external(e *config.ExternalFile) llb.State {
	s := curl().
		Run(shf("curl -Lo %s %s && chmod +x %s", e.Destination, e.Source, e.Destination)).Root()
	if e.Sha256 != "" {
		//TODO r2d4: get piping and echo to work with /bin/sh -c
		s = s.Run(shf("echo \"%s  %s\" | sha256sum -c -", e.Sha256, e.Destination)).Root()
	}
	return s
}

func shf(cmd string, v ...interface{}) llb.RunOption {
	return llb.Args([]string{"/bin/sh", "-c", fmt.Sprintf(cmd, v...)})
}

func sh(cmd string) llb.RunOption {
	return llb.Args([]string{"/bin/sh", "-c", cmd})
}

func copy(src llb.State, srcPath string, dest llb.State, destPath string) llb.State {
	cpImage := llb.Image("docker.io/library/alpine:3.6")
	cp := cpImage.Run(llb.Shlexf("cp -a /src%s /dest%s", srcPath, destPath))
	cp.AddMount("/src", src, llb.Readonly)
	return cp.AddMount("/dest", dest)
}
