package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/frontend/gateway/grpcclient"
	"github.com/moby/buildkit/util/appcontext"
	"github.com/pkg/errors"
	mockerfile "github.com/r2d4/mockerfile/pkg/build"
	"github.com/r2d4/mockerfile/pkg/mocker/config"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := grpcclient.RunFromEnvironment(appcontext.Context(), mockerfile.Build); err != nil {
		logrus.Fatalf("fatal error: %s", err)
		panic(err)
	}
}

func run() error {
	c, err := config.NewFromFilename("docker.yaml")
	if err != nil {
		return errors.Wrap(err, "config")
	}
	s := solve(c)
	out := s.Run(llb.Shlex("ls -l /bin"))
	dt, err := out.Marshal(llb.LinuxAmd64)
	if err != nil {
		return errors.Wrap(err, "marshaling llb")
	}
	llb.WriteTo(dt, os.Stdout)
	return nil
}

func curl() llb.State {
	return llb.Image("docker.io/library/alpine:3.6").
		Run(llb.Shlex("apk add --no-cache curl")).Root()
}

func packages(base llb.State, p *config.Package) llb.State {
	if len(p.Repo) > 0 {
		cmds := []string{}
		for _, repo := range p.Repo {
			cmds = append(cmds, fmt.Sprintf("add-apt-repository %s", repo))
		}
		base = base.Run(llb.Shlex(strings.Join(cmds, " &&"))).Root()
	}
	for _, key := range p.Gpg {
		base = aptAddKey(base, key)
	}
	if len(p.Install) > 0 {
		packages := strings.Join(p.Install, " ")
		base = base.Run(llb.Shlex(fmt.Sprintf("apt-get install --no-install-recommends --no-install-suggests -y %s", packages))).Root()
	}
	return base
}

func solve(c *config.Config) llb.State {
	current := c.Images[1]
	s := llb.Image(current.From)
	if current.Package != nil {
		s = packages(s, current.Package)
	}
	for _, e := range current.ExternalFiles {
		downloaded := external(e)
		s = copy(s, e.Destination, downloaded, e.Destination)
	}
	return s
}

func aptAddKey(s llb.State, url string) llb.State {
	return s.Run(llb.Shlex(fmt.Sprintf("curl -fsSL %s | apt-key add -", url))).Root()
}

func external(e *config.ExternalFile) llb.State {
	cmd := fmt.Sprintf("curl -Lo %s %s && chmod +x %s", e.Source, e.Destination, e.Source)
	return curl().
		Run(llb.Shlex(cmd)).Root()
}

func copy(src llb.State, srcPath string, dest llb.State, destPath string) llb.State {
	cpImage := llb.Image("docker.io/library/alpine:3.6")
	cp := cpImage.Run(llb.Shlexf("cp -a /src%s /dest%s", srcPath, destPath))
	cp.AddMount("/src", src)
	return cp.AddMount("/dest", dest)
}
