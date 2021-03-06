// Copyright (c) 2016 - 2019 Sqreen. All Rights Reserved.
// Please refer to our terms for more information:
// https://www.sqreen.io/terms.html

package backend_test

import (
	"net/http"
	"os"

	fuzz "github.com/google/gofuzz"
	"github.com/onsi/gomega/ghttp"
	"github.com/sqreen/go-agent/internal/backend"
	"github.com/sqreen/go-agent/internal/backend/api"
	"github.com/sqreen/go-agent/internal/config"
	"github.com/sqreen/go-agent/internal/plog"
	"github.com/sqreen/go-agent/tools/testlib"
)

var (
	logger = plog.NewLogger(plog.Debug, os.Stderr, nil)
	fuzzer = fuzz.New().Funcs(FuzzStruct, FuzzCommandRequest, FuzzRuleDataValue, FuzzRuleSignature)
)

// TODO: fuzzing is a mess for such a dynamic API - copy past real-world
//  examples instead
//func TestClient(t *testing.T) {
//	RegisterTestingT(t)
//
//	//t.Run("AppLogin", func(t *testing.T) {
//	//	g := NewGomegaWithT(t)
//	//
//	//	token := testlib.RandHTTPHeaderValue(2, 50)
//	//	appName := testlib.RandHTTPHeaderValue(2, 50)
//	//
//	//	statusCode := http.StatusOK
//	//
//	//	endpointCfg := &config.BackendHTTPAPIEndpoint.AppLogin
//	//
//	//	response := NewRandomAppLoginResponse()
//	//	request := NewRandomAppLoginRequest()
//	//
//	//	headers := http.Header{
//	//		config.BackendHTTPAPIHeaderToken:   []string{token},
//	//		config.BackendHTTPAPIHeaderAppName: []string{appName},
//	//	}
//	//
//	//	server := initFakeServer(endpointCfg, request, response, statusCode, headers)
//	//	defer server.Close()
//	//
//	//	client, err := backend.NewClient(server.URL(), "", logger)
//	//	require.NoError(t, err)
//	//
//	//	res, err := client.AppLogin(request, token, appName, false)
//	//	g.Expect(err).NotTo(HaveOccurred())
//	//	// A request has been received
//	//	g.Expect(len(server.ReceivedRequests())).ToNot(Equal(0))
//	//	g.Expect(res).Should(Equal(response))
//	//})
//
//	//t.Run("AppBeat", func(t *testing.T) {
//	//	g := NewGomegaWithT(t)
//	//
//	//	statusCode := http.StatusOK
//	//
//	//	endpointCfg := &config.BackendHTTPAPIEndpoint.AppBeat
//	//
//	//	response := NewRandomAppBeatResponse()
//	//	request := NewRandomAppBeatRequest()
//	//
//	//	client, server := initFakeServerSession(endpointCfg, request, response, statusCode, nil)
//	//	defer server.Close()
//	//
//	//	res, err := client.AppBeat(context.Background(), request)
//	//	g.Expect(err).NotTo(HaveOccurred())
//	//	// A request has been received
//	//	g.Expect(len(server.ReceivedRequests())).ToNot(Equal(0))
//	//	g.Expect(res).Should(Equal(response))
//	//})
//
//	//t.Run("Batch", func(t *testing.T) {
//	//	g := NewGomegaWithT(t)
//	//
//	//	statusCode := http.StatusOK
//	//
//	//	endpointCfg := &config.BackendHTTPAPIEndpoint.Batch
//	//
//	//	request := NewRandomBatchRequest()
//	//	t.Logf("%#v", request)
//	//
//	//	client, server := initFakeServerSession(endpointCfg, request, nil, statusCode, nil)
//	//	defer server.Close()
//	//
//	//	err := client.Batch(context.Background(), request)
//	//	g.Expect(err).NotTo(HaveOccurred())
//	//	// A request has been received
//	//	g.Expect(len(server.ReceivedRequests())).ToNot(Equal(0))
//	//})
//
//	t.Run("ActionsPack", func(t *testing.T) {
//		g := NewGomegaWithT(t)
//
//		statusCode := http.StatusOK
//
//		endpointCfg := &config.BackendHTTPAPIEndpoint.ActionsPack
//
//		response := NewRandomActionsPackResponse()
//
//		client, server := initFakeServerSession(endpointCfg, nil, response, statusCode, nil)
//		defer server.Close()
//
//		res, err := client.ActionsPack()
//		g.Expect(err).NotTo(HaveOccurred())
//		// A request has been received
//		g.Expect(len(server.ReceivedRequests())).ToNot(Equal(0))
//		g.Expect(res).Should(Equal(response))
//	})
//
//	t.Run("AppLogout", func(t *testing.T) {
//		g := NewGomegaWithT(t)
//
//		statusCode := http.StatusOK
//
//		endpointCfg := &config.BackendHTTPAPIEndpoint.AppLogout
//
//		client, server := initFakeServerSession(endpointCfg, nil, nil, statusCode, nil)
//		defer server.Close()
//
//		err := client.AppLogout()
//		g.Expect(err).NotTo(HaveOccurred())
//		// A request has been received
//		g.Expect(len(server.ReceivedRequests())).ToNot(Equal(0))
//	})
//}

func initFakeServer(endpointCfg *config.HTTPAPIEndpoint, request, response interface{}, statusCode int, headers http.Header) *ghttp.Server {
	handlers := []http.HandlerFunc{
		ghttp.VerifyRequest(endpointCfg.Method, endpointCfg.URL),
		ghttp.VerifyHeader(headers),
	}

	if request != nil {
		handlers = append(handlers, ghttp.VerifyJSONRepresenting(request))
	}

	if response != nil {
		handlers = append(handlers, ghttp.RespondWithJSONEncoded(statusCode, response))
	} else {
		handlers = append(handlers, ghttp.RespondWith(statusCode, nil))
	}

	server := ghttp.NewServer()
	server.AppendHandlers(ghttp.CombineHandlers(handlers...))
	return server
}

func initFakeServerSession(endpointCfg *config.HTTPAPIEndpoint, request, response interface{}, statusCode int, headers http.Header) (client *backend.Client, server *ghttp.Server) {
	server = ghttp.NewServer()

	loginReq := NewRandomAppLoginRequest()
	loginRes := NewRandomAppLoginResponse()
	loginRes.SessionId = testlib.RandHTTPHeaderValue(2, 50)
	loginRes.Status = true
	server.AppendHandlers(ghttp.RespondWithJSONEncoded(http.StatusOK, loginRes))

	client, err := backend.NewClient(server.URL(), "", logger)
	if err != nil {
		panic(err)
	}

	token := testlib.RandHTTPHeaderValue(2, 50)
	appName := testlib.RandHTTPHeaderValue(2, 50)
	_, err = client.AppLogin(loginReq, token, appName, false)
	if err != nil {
		panic(err)
	}

	if headers != nil {
		headers.Add(config.BackendHTTPAPIHeaderSession, loginRes.SessionId)
	} else {
		headers = http.Header{
			config.BackendHTTPAPIHeaderSession: []string{loginRes.SessionId},
		}
	}

	handlers := []http.HandlerFunc{
		ghttp.VerifyRequest(endpointCfg.Method, endpointCfg.URL),
		ghttp.VerifyHeader(headers),
	}

	if request != nil {
		handlers = append(handlers, ghttp.VerifyJSONRepresenting(request))
	}

	if response != nil {
		handlers = append(handlers, ghttp.RespondWithJSONEncoded(statusCode, response))
	} else {
		handlers = append(handlers, ghttp.RespondWith(statusCode, nil))
	}

	server.AppendHandlers(ghttp.CombineHandlers(handlers...))

	return client, server
}

func NewRandomAppLoginResponse() *api.AppLoginResponse {
	pb := new(api.AppLoginResponse)
	fuzzer.Fuzz(pb)
	// We don't want to use signals in these tests.
	pb.Features.UseSignals = false
	return pb
}

func NewRandomAppLoginRequest() *api.AppLoginRequest {
	pb := new(api.AppLoginRequest)
	fuzzer.Fuzz(pb)
	return pb
}

func NewRandomAppBeatResponse() *api.AppBeatResponse {
	pb := new(api.AppBeatResponse)
	fuzzer.Fuzz(pb)
	return pb
}

func NewRandomAppBeatRequest() *api.AppBeatRequest {
	pb := new(api.AppBeatRequest)
	fuzzer.Fuzz(pb)
	return pb
}

func NewRandomBatchRequest() *api.BatchRequest {
	pb := new(api.BatchRequest)
	fuzzer.Fuzz(pb)
	return pb
}

func NewRandomActionsPackResponse() *api.ActionsPackResponse {
	pb := new(api.ActionsPackResponse)
	fuzzer.Fuzz(pb)
	return pb
}

func NewRandomRulesPackResponse() *api.RulesPackResponse {
	pb := new(api.RulesPackResponse)
	fuzzer.Fuzz(pb)
	return pb
}

func FuzzStruct(e *api.Struct, c fuzz.Continue) {
	v := struct {
		A string
		B int
		C float64
		D bool
		F []byte
		G struct {
			A string
			B int
			C float64
			D bool
			F []byte
		}
	}{}
	c.Fuzz(&v)
	e.Value = v
}

func FuzzCommandRequest(e *api.CommandRequest, c fuzz.Continue) {
	c.Fuzz(&e.Name)
	c.Fuzz(&e.Uuid)
}

func FuzzRuleDataValue(e *api.RuleDataEntry, c fuzz.Continue) {
	v := &api.CustomErrorPageRuleDataEntry{}
	c.Fuzz(&v.StatusCode)
	e.Value = v
}

func FuzzRuleSignature(s *api.RuleSignature, c fuzz.Continue) {
	*s = api.RuleSignature{ECDSASignature: api.ECDSASignature{Message: []byte(`{}`)}}
}
