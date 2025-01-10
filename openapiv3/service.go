package openapiv3

import (
	"context"
	"flag"
	mapset "github.com/deckarep/golang-set/v2"
	"google.golang.org/protobuf/compiler/protogen"
	"strings"

	"github.com/go-kratos/kratos/v2/api/metadata"
	"google.golang.org/grpc"
	dpb "google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/google/gnostic/cmd/protoc-gen-openapi/generator"
)

// Service is service
type Service struct {
	ser *metadata.Server
}

// New service
func New(srv *grpc.Server) *Service {
	return &Service{
		ser: metadata.NewServer(srv),
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

	return s.xxx(files)
}

func (s *Service) xxx(files []*dpb.FileDescriptorProto) (string, error) {
	var target string
	target = *files[len(files)-1].Name

	req := new(pluginpb.CodeGeneratorRequest)
	req.FileToGenerate = []string{target}
	var para = ""
	req.Parameter = &para
	req.ProtoFile = files

	opts := protogen.Options{}
	plugin, err := opts.New(req)
	if err != nil {
		return "", err
	}
	var flags flag.FlagSet
	conf := generator.Configuration{
		Version:         flags.String("version", "0.0.1", "version number text, e.g. 1.2.3"),
		Title:           flags.String("title", "", "name of the API"),
		Description:     flags.String("description", "", "description of the API"),
		Naming:          flags.String("naming", "json", `naming convention. Use "proto" for passing names directly from the proto files`),
		FQSchemaNaming:  flags.Bool("fq_schema_naming", false, `schema naming convention. If "true", generates fully-qualified schema names by prefixing them with the proto message package name`),
		EnumType:        flags.String("enum_type", "integer", `type for enum serialization. Use "string" for string-based serialization`),
		CircularDepth:   flags.Int("depth", 2, "depth of recursion for circular messages"),
		DefaultResponse: flags.Bool("default_response", true, `add default response. If "true", automatically adds a default response to operations which use the google.rpc.Status message. Useful if you use envoy or grpc-gateway to transcode as they use this type for their default error responses.`),
		OutputMode:      flags.String("output_mode", "merged", `output generation mode. By default, a single openapi.yaml is generated at the out folder. Use "source_relative' to generate a separate '[inputfile].openapi.yaml' next to each '[inputfile].proto'.`),
	}

	iv3Generator := generator.NewOpenAPIv3Generator(plugin, conf, plugin.Files)

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
	for _, ss := range servicesReply.Services {
		if strings.HasPrefix(ss, name) {
			services, err := s.ser.GetServiceDesc(nil, &metadata.GetServiceDescRequest{
				Name: name,
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

		}
	}

	return s.xxx(files)
}
