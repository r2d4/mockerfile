package main

import (
	"flag"
	"io"
	"os"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/frontend/gateway/grpcclient"
	"github.com/moby/buildkit/util/appcontext"
	"github.com/pkg/errors"
	mockerfile "github.com/r2d4/mockerfile/pkg/build"
	"github.com/r2d4/mockerfile/pkg/mocker/config"
	"github.com/r2d4/mockerfile/pkg/mockerfile2llb"
	"github.com/sirupsen/logrus"
)

var graph bool
var filename string

func main() {
	flag.BoolVar(&graph, "graph", false, "output a graph and exit")
	flag.StringVar(&filename, "filename", "Mockerfile.yaml", "the file to read from")
	flag.Parse()

	if graph {
		if err := printGraph(filename, os.Stdout); err != nil {
			logrus.Fatalf("fatal error: %s", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if err := grpcclient.RunFromEnvironment(appcontext.Context(), mockerfile.Build); err != nil {
		logrus.Fatalf("fatal error: %s", err)
		panic(err)
	}
}

func printGraph(filename string, out io.Writer) error {
	c, err := config.NewFromFilename(filename)
	if err != nil {
		return errors.Wrap(err, "opening config file")
	}
	st, _ := mockerfile2llb.Mockerfile2LLB(c)
	dt, err := st.Marshal()
	if err != nil {
		return errors.Wrap(err, "marshaling llb state")
	}

	return llb.WriteTo(dt, out)
}
