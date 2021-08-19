package main

import (
	"context"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"

	"github.com/go-openapi/inflect"
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/proto/spiffe/workload"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spire/proto/spire/api/node"
	"github.com/spiffe/spire/proto/spire/api/server/agent/v1"
	"github.com/spiffe/spire/proto/spire/api/server/bundle/v1"
	"github.com/spiffe/spire/proto/spire/api/server/entry/v1"
	"github.com/spiffe/spire/proto/spire/api/server/svid/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	apiNameRuleset *inflect.Ruleset
	zeroValue      reflect.Value
)

func init() {
	apiNameRuleset = inflect.NewDefaultRuleset()
	apiNameRuleset.AddAcronym("SVID")
	apiNameRuleset.AddAcronym("ID")
	apiNameRuleset.AddAcronym("CA")
	apiNameRuleset.AddAcronym("JWT")
	apiNameRuleset.AddAcronym("X509")
}

func dasherizeAPIName(s string) string {
	return apiNameRuleset.Dasherize(s)
}

func RPCCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "rpc"}
	cmd.AddCommand(makeServerAPICommands("Node", node.NewNodeClient))
	cmd.AddCommand(makeServerAPICommands("Agent", agent.NewAgentClient))
	cmd.AddCommand(makeServerAPICommands("Entry", entry.NewEntryClient))
	cmd.AddCommand(makeServerAPICommands("Bundle", bundle.NewBundleClient))
	cmd.AddCommand(makeServerAPICommands("SVID", svid.NewSVIDClient))
	cmd.AddCommand(makeAgentAPICommands("Workload", workload.NewSpiffeWorkloadAPIClient))
	return cmd
}

type rpcConfig struct {
	tcpAddr         string
	udsAddr         string
	useTCP          bool
	useWorkloadAPI  bool
	svidPath        string
	workloadAPIAddr string
	metadataPairs   []string
}

func makeServerAPICommands(groupName string, clientFn interface{}) *cobra.Command {
	config := new(rpcConfig)
	cmd := &cobra.Command{
		Use:   dasherizeAPIName(groupName),
		Short: fmt.Sprintf("%s API RPCs", groupName),
	}
	cmd.PersistentFlags().StringVarP(&config.tcpAddr, "tcp-addr", "", "localhost:8081", "server TCP address")
	cmd.PersistentFlags().StringVarP(&config.udsAddr, "uds-addr", "", "unix:///tmp/spire-server/private/api.sock", "server UDS address")
	cmd.PersistentFlags().BoolVarP(&config.useTCP, "use-tcp", "", false, "Issue RPC via TCP")
	cmd.PersistentFlags().StringVarP(&config.svidPath, "svid-path", "", "", "SVID to use to issue the RPC (implies --use-tcp)")
	cmd.PersistentFlags().BoolVarP(&config.useWorkloadAPI, "use-workload-api", "", false, "Use the Workload API to obtain an SVID used to issue the RPC (implies --use-tcp)")
	cmd.PersistentFlags().StringVarP(&config.workloadAPIAddr, "workload-api-addr", "", "unix:///tmp/spire-agent/public/api.sock", "Address to the Workload API socket")
	addRPCCommands(cmd, groupName, clientFn, config)
	return cmd
}

func makeAgentAPICommands(groupName string, clientFn interface{}) *cobra.Command {
	config := new(rpcConfig)
	cmd := &cobra.Command{
		Use:   dasherizeAPIName(groupName),
		Short: fmt.Sprintf("%s API RPCs", groupName),
	}
	cmd.Flags().StringVarP(&config.udsAddr, "uds-addr", "", "unix:///tmp/spire-agent/public/api.sock", "agent UDS address")
	addRPCCommands(cmd, groupName, clientFn, config, setWorkloadAPIHeader)
	return cmd
}

func addRPCCommands(cmd *cobra.Command, groupName string, clientFn interface{}, config *rpcConfig, options ...rpcOption) {
	for _, option := range options {
		option(config)
	}
	fnv := reflect.ValueOf(clientFn)
	fnt := reflect.TypeOf(clientFn)

	outt := fnt.Out(0)
	for i := 0; i < outt.NumMethod(); i++ {
		mt := outt.Method(i)
		cmd.AddCommand(makeRPCCommand(fnv, groupName, mt.Name, config))
	}
}

func makeRPCCommand(newClientFn reflect.Value, groupName, methodName string, config *rpcConfig) *cobra.Command {
	impl := &rpcCommand{
		newClientFn: newClientFn,
		methodName:  methodName,
		config:      config,
	}
	cmd := &cobra.Command{
		Use:   dasherizeAPIName(methodName),
		Short: fmt.Sprintf("Invoke the %s %s RPC", groupName, methodName),
		RunE:  runInOut(impl),
	}
	return cmd
}

type rpcOption func(*rpcConfig)

type rpcCommand struct {
	newClientFn reflect.Value
	methodName  string
	useTCP      bool
	svidPath    string
	config      *rpcConfig
}

func (cmd *rpcCommand) Run(ctx context.Context, jsonIn []byte, args []string) ([]byte, error) {
	makeReq := func(t reflect.Type) (reflect.Value, error) {
		req := reflect.New(t.Elem())
		if len(jsonIn) == 0 {
			return zeroValue, fmt.Errorf("stdin was empty. Did you forget to pipe?")
		}
		if err := protojson.Unmarshal(jsonIn, req.Interface().(proto.Message)); err != nil {
			return zeroValue, fmt.Errorf("unmarshaling request: %v", err)
		}
		return req, nil
	}

	callErr := func(v reflect.Value) error {
		if e := v.Interface(); e != nil {
			st := status.Convert(e.(error))
			return fmt.Errorf("rpc %s: %s: %s", cmd.methodName, st.Code(), st.Message())
		}
		return nil
	}

	var conn *grpc.ClientConn
	var err error
	switch {
	case cmd.config.svidPath != "":
		conn, err = dialTCPWithSVID(cmd.config.tcpAddr, cmd.svidPath)
	case cmd.config.useWorkloadAPI:
		conn, err = dialTCPWithSVIDFromWorkloadAPI(ctx, cmd.config.tcpAddr, cmd.config.workloadAPIAddr)
	case cmd.config.useTCP:
		conn, err = dialInsecureTCP(cmd.config.tcpAddr)
	default:
		conn, err = dialUDS(cmd.config.udsAddr)
	}
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if len(cmd.config.metadataPairs) > 0 {
		ctx = metadata.AppendToOutgoingContext(ctx, cmd.config.metadataPairs...)
	}

	cv := cmd.newClientFn.Call(asValues(conn))[0]
	mv := cv.MethodByName(cmd.methodName)
	mt := mv.Type()

	var out []reflect.Value
	if mt.NumIn() == 2 {
		// Input doesn't take in request message. We'll have to send
		// that via Send()/SendAndClose().
		out = mv.Call(asValues(ctx))
		if err := callErr(out[1]); err != nil {
			return nil, err
		}

		send := out[0].MethodByName("Send")
		if send == zeroValue {
			send = out[0].MethodByName("SendAndClose")
		}

		req, err := makeReq(send.Type().In(0))
		if err != nil {
			return nil, err
		}
		if err := callErr(send.Call(asValues(req))[0]); err != nil {
			return nil, err
		}
	} else {
		// Input does take request message.
		req, err := makeReq(mt.In(1))
		if err != nil {
			return nil, err
		}
		out = mv.Call(asValues(ctx, req))
		if err := callErr(out[1]); err != nil {
			return nil, err
		}
	}

	if recv := out[0].MethodByName("Recv"); recv != zeroValue {
		out = recv.Call(asValues())
		if err := callErr(out[1]); err != nil {
			return nil, err
		}
	} else if recv := out[0].MethodByName("CloseAndRecv"); recv != zeroValue {
		out = recv.Call(asValues())
		if err := callErr(out[1]); err != nil {
			return nil, err
		}
	}

	var jsonOut []byte
	if !out[0].IsNil() {
		jsonOut = marshalProtoJSON(out[0].Interface().(proto.Message))
	}
	return jsonOut, nil
}

func setWorkloadAPIHeader(c *rpcConfig) {
	c.metadataPairs = append(c.metadataPairs, "workload.spiffe.io", "true")
}

func asValues(args ...interface{}) []reflect.Value {
	var out []reflect.Value
	for _, arg := range args {
		if v, ok := arg.(reflect.Value); ok {
			out = append(out, v)
		} else {
			out = append(out, reflect.ValueOf(arg))
		}
	}
	return out
}

func dialOptions(options ...grpc.DialOption) []grpc.DialOption {
	return append(options, grpc.WithBlock(), grpc.FailOnNonTempDialError(true), grpc.WithReturnConnectionError())
}

func dialUDS(addr string) (*grpc.ClientConn, error) {
	return grpc.Dial(addr, dialOptions(grpc.WithInsecure())...)
}

func dialTCP(addr string, tlsConfig *tls.Config) (*grpc.ClientConn, error) {
	return grpc.Dial(addr, dialOptions(grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))...)
}

func dialInsecureTCP(addr string) (*grpc.ClientConn, error) {
	return dialTCP(addr, &tls.Config{InsecureSkipVerify: true})
}

func dialTCPWithSVID(addr, svidPath string) (*grpc.ClientConn, error) {
	svid, key, err := loadSVID(svidPath)
	if err != nil {
		return nil, err
	}
	return dialTCP(addr, tlsConfigForSVID(svid, key))
}

func dialTCPWithSVIDFromWorkloadAPI(ctx context.Context, addr, workloadAPIPath string) (*grpc.ClientConn, error) {
	var opts []workloadapi.X509SourceOption
	if workloadAPIPath != "" {
		opts = append(opts, workloadapi.WithClientOptions(workloadapi.WithAddr(workloadAPIPath)))
	}
	source, err := workloadapi.NewX509Source(ctx, opts...)
	if err != nil {
		return nil, err
	}

	// TODO: stricter authorizer?
	return dialTCP(addr, tlsconfig.MTLSClientConfig(source, source, tlsconfig.AuthorizeAny()))
}

func tlsConfigForSVID(svid []*x509.Certificate, key crypto.Signer) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: rawCertsFromCertificates(svid),
				PrivateKey:  key,
			},
		},
		// TODO: should probably verify the server certificate
		InsecureSkipVerify: true,
	}
}

func rawCertsFromCertificates(certs []*x509.Certificate) [][]byte {
	rawCerts := make([][]byte, 0, len(certs))
	for _, cert := range certs {
		rawCerts = append(rawCerts, cert.Raw)
	}
	return rawCerts
}

func loadSVID(path string) (_ []*x509.Certificate, _ crypto.Signer, err error) {
	pemBytes, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("unable to load SVID: %v", err)
	}

	var key crypto.PrivateKey
	var certs []*x509.Certificate
	rest := pemBytes
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		switch block.Type {
		case "CERTIFICATE":
			if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
				certs = append(certs, cert)
			} else {
				return nil, nil, fmt.Errorf("bad certificate in PEM block: %v", err)
			}
		case "PRIVATE KEY":
			key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, nil, fmt.Errorf("bad key in PEM block: %v", err)
			}
		default:
			log.Panicf("Unexpected block type %q in PEM", block.Type)
		}
	}

	if len(certs) == 0 {
		return nil, nil, errors.New("no certificates in PEM file")
	}
	if key == nil {
		return nil, nil, errors.New("no key in PEM file")
	}
	return certs, key.(crypto.Signer), nil
}

func marshalProtoJSON(m proto.Message) []byte {
	options := protojson.MarshalOptions{
		Multiline: true,
	}
	return []byte(options.Format(m))
}
