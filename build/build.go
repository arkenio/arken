	package build

import (
	"html/template"
	"os"

	"archive/tar"
	"bufio"
	"bytes"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"golang.org/x/net/context"
	"io/ioutil"
	"log"
)

type BuildService struct {
	dockerHost string
}

type BuildDescriptor struct {
	LocalTag string
	Params   map[string]string
}

func NewBuildService() *BuildService {
	return &BuildService{
		dockerHost: "unix:///var/run/docker.sock",
	}
}

func (bs *BuildService) Build(ctx context.Context, desc *BuildDescriptor) {
	tarfile := compileTarArchive(desc)
	bs.launchDockerBuild(ctx, tarfile, desc.LocalTag)
}

func generateDockerFile(desc *BuildDescriptor) bytes.Buffer {


	// Build the Docker file
	tpl := FSMustString(false, "/Dockerfile")
	tmpl, _ := template.New("dockerfile").Parse(tpl)
	var b bytes.Buffer
	out := bufio.NewWriter(&b)
	err := tmpl.Execute(out, desc.Params)
	if err != nil {
		panic(err)
	}
	out.Flush()
	return b
}

func compileTarArchive(desc *BuildDescriptor) *os.File {


	// Create a new tar archive.
	tarfile, err := ioutil.TempFile("/tmp", "arken-build")
	if err != nil {
		panic(err)
	}
	tw := tar.NewWriter(tarfile)
	defer tw.Close()

	b := generateDockerFile(desc)
	if err := tw.WriteHeader(&tar.Header{Name: "Dockerfile", Size: int64(b.Len())}); err != nil {
		log.Fatalln(err)
	}
	tw.Write(b.Bytes())

	tw.Close()
	return tarfile
}

func (bs *BuildService) launchDockerBuild(ctx context.Context, tarfile *os.File, destImageName string) {
	// Launch the build
	defaultHeaders := map[string]string{"User-Agent": "docker-client"}
	cli, err := client.NewClient(bs.dockerHost, "v1.22", nil, defaultHeaders)
	if err != nil {
		panic(err)
	}

	dockerBuildContext, err := os.Open(tarfile.Name())
	defer dockerBuildContext.Close()

	resp, err := cli.ImageBuild(ctx, dockerBuildContext, types.ImageBuildOptions{Tags: []string{destImageName}})
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	err = jsonmessage.DisplayJSONMessagesStream(resp.Body, buf, 0, true, nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf(buf.String())
}
