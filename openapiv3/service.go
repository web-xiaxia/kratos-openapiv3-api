package openapiv3

import (
	"context"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"google.golang.org/protobuf/compiler/protogen"
	"strings"

	"github.com/go-kratos/kratos/v2/api/metadata"
	dpb "google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/google/gnostic/cmd/protoc-gen-openapi/generator"
)

func ToPtr[T any](x T) *T {
	return &x
}

// Service is service
type Service struct {
	ser  *metadata.Server
	conf *generator.Configuration
}

// New service
func New(opts ...Option) *Service {
	o := &options{
		conf: func(c *generator.Configuration) {

		},
	}
	for _, opt := range opts {
		opt(o)
	}
	conf := &generator.Configuration{
		// Version:         (*)"0.0.1",
		Title:           ToPtr(""),
		Description:     ToPtr(""),
		Naming:          ToPtr("json"),
		FQSchemaNaming:  ToPtr(false),
		EnumType:        ToPtr("integer"),
		CircularDepth:   ToPtr(2),
		DefaultResponse: ToPtr(false),
		OutputMode:      ToPtr("merged"),
	}
	o.conf(conf)
	return &Service{
		ser:  metadata.NewServer(nil),
		conf: conf,
	}
}

// ListServices list services
func (s *Service) ListServices(ctx context.Context) (*metadata.ListServicesReply, error) {
	return s.ser.ListServices(ctx, &metadata.ListServicesRequest{})
}

// GetServiceOpenAPI get service open api
func (s *Service) GetServiceOpenAPI(_ context.Context, name string) (string, error) {
	services, err := s.ser.GetServiceDesc(nil, &metadata.GetServiceDescRequest{
		Name: name,
	})
	if err != nil {
		return "", err
	}

	files1 := services.GetFileDescSet().File
	files := make([]*dpb.FileDescriptorProto, 0, len(files1))
	xx := mapset.NewSet[string]()
	for _, ff := range files1 {
		if xx.ContainsAny(*ff.Name) {
			continue
		}
		xx.Add(*ff.Name)
		files = append(files, ff)
	}
	return s.generated(files, []string{*files[len(files)-1].Name})
}

func (s *Service) generated(files []*dpb.FileDescriptorProto, target []string) (string, error) {

	req := new(pluginpb.CodeGeneratorRequest)
	req.FileToGenerate = target
	var para = ""
	req.Parameter = &para
	req.ProtoFile = files

	opts := protogen.Options{}
	plugin, err := opts.New(req)
	if err != nil {
		return "", err
	}

	targetSet := mapset.NewSet[string](target...)
	for _, ff := range plugin.Files {
		packageArr := strings.Split(*ff.Proto.Package, ".")
		packageName := strings.ToLower(packageArr[len(packageArr)-1])
		if !targetSet.ContainsOne(*ff.Proto.Name) {
			continue
		}
		for _, ss := range ff.Services {
			if strings.ToLower(ss.GoName) != packageName {
				ss.GoName = fmt.Sprintf("%s.%s", packageName, ss.GoName)
			}
		}
	}

	iv3Generator := generator.NewOpenAPIv3Generator(plugin, *s.conf, plugin.Files)

	outputFile := plugin.NewGeneratedFile("", "")
	outputFile.Skip()
	err = iv3Generator.Run(outputFile)
	if err != nil {
		return "", err
	}
	content, err := outputFile.Content()
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (s *Service) GetServiceGroupOpenAPI(ctx context.Context, name string) (string, error) {
	servicesReply, err := s.ListServices(ctx)
	if err != nil {
		return "", err
	}

	files := make([]*dpb.FileDescriptorProto, 0)
	xx := mapset.NewSet[string]()
	target := make([]string, 0)
	for _, sName := range servicesReply.Services {
		if strings.HasPrefix(sName, name) {
			services, err := s.ser.GetServiceDesc(nil, &metadata.GetServiceDescRequest{
				Name: sName,
			})
			if err != nil {
				return "", err
			}

			files1 := services.GetFileDescSet().File
			for _, ff := range files1 {
				if xx.ContainsAny(*ff.Name) {
					continue
				}
				xx.Add(*ff.Name)
				files = append(files, ff)
			}
			target = append(target, *files[len(files)-1].Name)
		}
	}

	return s.generated(files, target)
}
