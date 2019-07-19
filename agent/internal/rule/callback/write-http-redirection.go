// Copyright (c) 2016 - 2019 Sqreen. All Rights Reserved.
// Please refer to our terms for more information:
// https://www.sqreen.io/terms.html

package callback

import (
	"net/http"
	"net/url"

	"github.com/sqreen/go-agent/agent/internal/backend/api"
	"github.com/sqreen/go-agent/agent/sqlib/sqerrors"
	"github.com/sqreen/go-agent/agent/sqlib/sqhook"
)

// NewWriteHTTPRedirectionCallbacks returns the native prolog and epilog
// callbacks modifying the arguments of `httphandler.WriteResponse` in order to
// modify the http status code and headers in order to perform an HTTP
// redirection to the URL provided by the rule's data.
func NewWriteHTTPRedirectionCallbacks(rule Context, nextProlog, nextEpilog sqhook.Callback) (prolog, epilog sqhook.Callback, err error) {
	var redirectionURL string
	if cfg := rule.Config(); cfg != nil {
		cfg, ok := cfg.(*api.RedirectionRuleDataEntry)
		if !ok {
			err = sqerrors.Errorf("unexpected callback data type: got `%T` instead of `*api.CustomErrorPageRuleDataEntry`", cfg)
			return
		}
		redirectionURL = cfg.RedirectionURL
	}
	if redirectionURL == "" {
		err = sqerrors.New("unexpected empty redirection url")
		return
	}
	if _, err = url.ParseRequestURI(redirectionURL); err != nil {
		err = sqerrors.Wrap(err, "validation error of the redirection url")
		return
	}

	// Next callbacks to call
	actualNextProlog, ok := nextProlog.(WriteHTTPRedirectionPrologCallbackType)
	if nextProlog != nil && !ok {
		err = sqerrors.Errorf("unexpected next prolog type `%T`", nextProlog)
		return
	}
	// No epilog in this callback, so simply check and pass the given one
	if _, ok := nextEpilog.(WriteHTTPRedirectionEpilogCallbackType); nextEpilog != nil && !ok {
		err = sqerrors.Errorf("unexpected next epilog type `%T` instead of `%T`", nextEpilog, WriteHTTPRedirectionEpilogCallbackType(nil))
		return
	}
	return newWriteHTTPRedirectionPrologCallback(redirectionURL, actualNextProlog), nextEpilog, nil
}

type WriteHTTPRedirectionPrologCallbackType = func(*sqhook.Context, *http.ResponseWriter, **http.Request, *http.Header, *int, *[]byte) error
type WriteHTTPRedirectionEpilogCallbackType = func(*sqhook.Context)

// The prolog callback modifies the function arguments in order to perform an
// HTTP redirection.
func newWriteHTTPRedirectionPrologCallback(url string, next WriteHTTPRedirectionPrologCallbackType) WriteHTTPRedirectionPrologCallbackType {
	return func(ctx *sqhook.Context, callerWriter *http.ResponseWriter, callerRequest **http.Request, callerHeaders *http.Header, callerStatusCode *int, callerBody *[]byte) error {
		*callerStatusCode = http.StatusSeeOther
		if *callerHeaders == nil {
			*callerHeaders = make(http.Header)
		}
		callerHeaders.Set("Location", url)

		if next == nil {
			return nil
		}
		return next(ctx, callerWriter, callerRequest, callerHeaders, callerStatusCode, callerBody)
	}
}
