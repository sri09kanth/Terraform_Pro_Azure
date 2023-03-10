package virtualmachines

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/go-azure-helpers/polling"
)

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See NOTICE.txt in the project root for license information.

type ClaimOperationResponse struct {
	Poller       polling.LongRunningPoller
	HttpResponse *http.Response
}

// Claim ...
func (c VirtualMachinesClient) Claim(ctx context.Context, id VirtualMachineId) (result ClaimOperationResponse, err error) {
	req, err := c.preparerForClaim(ctx, id)
	if err != nil {
		err = autorest.NewErrorWithError(err, "virtualmachines.VirtualMachinesClient", "Claim", nil, "Failure preparing request")
		return
	}

	result, err = c.senderForClaim(ctx, req)
	if err != nil {
		err = autorest.NewErrorWithError(err, "virtualmachines.VirtualMachinesClient", "Claim", result.HttpResponse, "Failure sending request")
		return
	}

	return
}

// ClaimThenPoll performs Claim then polls until it's completed
func (c VirtualMachinesClient) ClaimThenPoll(ctx context.Context, id VirtualMachineId) error {
	result, err := c.Claim(ctx, id)
	if err != nil {
		return fmt.Errorf("performing Claim: %+v", err)
	}

	if err := result.Poller.PollUntilDone(); err != nil {
		return fmt.Errorf("polling after Claim: %+v", err)
	}

	return nil
}

// preparerForClaim prepares the Claim request.
func (c VirtualMachinesClient) preparerForClaim(ctx context.Context, id VirtualMachineId) (*http.Request, error) {
	queryParameters := map[string]interface{}{
		"api-version": defaultApiVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsPost(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(fmt.Sprintf("%s/claim", id.ID())),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// senderForClaim sends the Claim request. The method will close the
// http.Response Body if it receives an error.
func (c VirtualMachinesClient) senderForClaim(ctx context.Context, req *http.Request) (future ClaimOperationResponse, err error) {
	var resp *http.Response
	resp, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		return
	}

	future.Poller, err = polling.NewPollerFromResponse(ctx, resp, c.Client, req.Method)
	return
}
