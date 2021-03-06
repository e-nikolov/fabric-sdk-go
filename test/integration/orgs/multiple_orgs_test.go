/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orgs

import (
	"math"
	"path"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"

	selection "github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/dynamicselection"

	"os"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
)

const (
	pollRetries      = 5
	org1             = "Org1"
	org2             = "Org2"
	ordererAdminUser = "Admin"
	ordererOrgName   = "ordererorg"
	org1AdminUser    = "Admin"
	org2AdminUser    = "Admin"
	org1User         = "User1"
	org2User         = "User1"
	channelID        = "orgchannel"
	exampleCC        = "exampleCC"
)

var (
	// SDK
	sdk *fabsdk.FabricSDK

	// Org MSP clients
	org1MspClient *mspclient.Client
	org2MspClient *mspclient.Client
	// Peers
	orgTestPeer0 fab.Peer
	orgTestPeer1 fab.Peer

	isJoinedChannel bool
)

// used to create context for different tests in the orgs package
type multiorgContext struct {
	// client contexts
	ordererClientContext   contextAPI.ClientProvider
	org1AdminClientContext contextAPI.ClientProvider
	org2AdminClientContext contextAPI.ClientProvider
	org1ResMgmt            *resmgmt.Client
	org2ResMgmt            *resmgmt.Client
	ccName                 string
}

func TestMain(m *testing.M) {
	err := setup()
	defer teardown()
	var r int
	if err == nil {
		r = m.Run()
	}
	defer os.Exit(r)
	runtime.Goexit()
}

func setup() error {
	// Create SDK setup for the integration tests
	var err error
	sdk, err = fabsdk.New(integration.ConfigBackend)
	if err != nil {
		return errors.Wrap(err, "Failed to create new SDK")
	}

	org1MspClient, err = mspclient.New(sdk.Context(), mspclient.WithOrg(org1))
	if err != nil {
		return errors.Wrap(err, "failed to create org1MspClient, err")
	}

	org2MspClient, err = mspclient.New(sdk.Context(), mspclient.WithOrg(org2))
	if err != nil {
		return errors.Wrap(err, "failed to create org2MspClient, err")
	}

	return nil
}

func teardown() {
	if sdk != nil {
		sdk.Close()
	}
}

// TestOrgsEndToEnd creates a channel with two organisations, installs chaincode
// on each of them, and finally invokes a transaction on an org2 peer and queries
// the result from an org1 peer
func TestOrgsEndToEnd(t *testing.T) {

	// Delete all private keys from the crypto suite store
	// and users from the user store at the end
	integration.CleanupUserData(t, sdk)
	defer integration.CleanupUserData(t, sdk)

	// Load specific targets for move funds test
	loadOrgPeers(t, sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1)))

	//prepare contexts
	mc := multiorgContext{
		ordererClientContext:   sdk.Context(fabsdk.WithUser(ordererAdminUser), fabsdk.WithOrg(ordererOrgName)),
		org1AdminClientContext: sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1)),
		org2AdminClientContext: sdk.Context(fabsdk.WithUser(org2AdminUser), fabsdk.WithOrg(org2)),
		ccName:                 exampleCC, // basic multi orgs test uses exampleCC for testing
	}

	expectedValue := testWithOrg1(t, sdk, &mc)
	expectedValue = testWithOrg2(t, expectedValue, mc.ccName)
	verifyWithOrg1(t, sdk, expectedValue, mc.ccName)
}

func setupClientContextsAndChannel(t *testing.T, sdk *fabsdk.FabricSDK, mc *multiorgContext) {
	// Get signing identity that is used to sign create channel request
	org1AdminUser, err := org1MspClient.GetSigningIdentity(org1AdminUser)
	if err != nil {
		t.Fatalf("failed to get org1AdminUser, err : %v", err)
	}

	org2AdminUser, err := org2MspClient.GetSigningIdentity(org2AdminUser)
	if err != nil {
		t.Fatalf("failed to get org2AdminUser, err : %v", err)
	}

	// Org1 resource management client (Org1 is default org)
	org1RMgmt, err := resmgmt.New(mc.org1AdminClientContext)
	require.NoError(t, err, "failed to create org1 resource management client")

	mc.org1ResMgmt = org1RMgmt

	// Org2 resource management client
	org2RMgmt, err := resmgmt.New(mc.org2AdminClientContext)
	require.NoError(t, err, "failed to create org2 resource management client")

	mc.org2ResMgmt = org2RMgmt

	// create/join channel if was not already done
	if !isJoinedChannel {
		defer func() { isJoinedChannel = true }()
		createChannel(org1AdminUser, org2AdminUser, mc, t)
		// Org1 peers join channel
		err = mc.org1ResMgmt.JoinChannel(channelID, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com"))
		require.NoError(t, err, "Org1 peers failed to JoinChannel")

		// Org2 peers join channel
		err = mc.org2ResMgmt.JoinChannel(channelID, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com"))
		require.NoError(t, err, "Org2 peers failed to JoinChannel")
	}
}

func testWithOrg1(t *testing.T, sdk *fabsdk.FabricSDK, mc *multiorgContext) int {

	org1AdminChannelContext := sdk.ChannelContext(channelID, fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1))
	org1ChannelClientContext := sdk.ChannelContext(channelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1))
	org2ChannelClientContext := sdk.ChannelContext(channelID, fabsdk.WithUser(org2User), fabsdk.WithOrg(org2))

	setupClientContextsAndChannel(t, sdk, mc)

	ccPkg, err := packager.NewCCPackage("github.com/example_cc", "../../fixtures/testdata")
	if err != nil {
		t.Fatal(err)
	}

	// Create chaincode package for example cc
	createCC(t, mc.org1ResMgmt, mc.org2ResMgmt, ccPkg, mc.ccName)

	chClientOrg1User, chClientOrg2User := connectUserToOrgChannel(org1ChannelClientContext, t, org2ChannelClientContext)

	// Call with a dummy function and expect a fail with multiple errors
	verifyErrorFromCC(chClientOrg1User, t, mc.ccName)

	// Org1 user queries initial value on both peers
	value := queryCC(chClientOrg1User, t, mc.ccName)
	initial, _ := strconv.Atoi(string(value))

	ledgerClient, err := ledger.New(org1AdminChannelContext)
	if err != nil {
		t.Fatalf("Failed to create new ledger client: %s", err)
	}

	// Ledger client will verify blockchain info
	ledgerInfoBefore := getBlockchainInfo(ledgerClient, t)

	// Org2 user moves funds on org2 peer
	transactionID := moveFunds(chClientOrg2User, t, mc.ccName)

	// Assert that funds have changed value on org1 peer
	verifyValue(t, chClientOrg1User, initial+1, mc.ccName)

	// Get latest blockchain info
	checkLedgerInfo(ledgerClient, t, ledgerInfoBefore, transactionID)

	// Start chaincode upgrade process (install and instantiate new version of exampleCC)
	upgradeCC(ccPkg, mc.org1ResMgmt, t, mc.org2ResMgmt, mc.ccName)

	// Org2 user moves funds on org2 peer (cc policy fails since both Org1 and Org2 peers should participate)
	testCCPolicy(chClientOrg2User, t, mc.ccName)

	// Assert that funds have changed value on org1 peer
	beforeTxValue, _ := strconv.Atoi(integration.ExampleCCUpgradeB)
	expectedValue := beforeTxValue + 1
	verifyValue(t, chClientOrg1User, expectedValue, mc.ccName)

	return expectedValue
}

func connectUserToOrgChannel(org1ChannelClientContext contextAPI.ChannelProvider, t *testing.T, org2ChannelClientContext contextAPI.ChannelProvider) (*channel.Client, *channel.Client) {
	// Org1 user connects to 'orgchannel'
	chClientOrg1User, err := channel.New(org1ChannelClientContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client for Org1 user: %s", err)
	}
	// Org2 user connects to 'orgchannel'
	chClientOrg2User, err := channel.New(org2ChannelClientContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client for Org2 user: %s", err)
	}
	return chClientOrg1User, chClientOrg2User
}

func checkLedgerInfo(ledgerClient *ledger.Client, t *testing.T, ledgerInfoBefore *fab.BlockchainInfoResponse, transactionID fab.TransactionID) {
	ledgerInfoAfter, err := ledgerClient.QueryInfo(ledger.WithTargets(orgTestPeer0.(fab.Peer), orgTestPeer1.(fab.Peer)), ledger.WithMinTargets(2))
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}
	if ledgerInfoAfter.BCI.Height-ledgerInfoBefore.BCI.Height <= 0 {
		t.Fatalf("Block size did not increase after transaction")
	}
	// Test Query Block by Hash - retrieve current block by number
	block, err := ledgerClient.QueryBlock(ledgerInfoAfter.BCI.Height-1, ledger.WithTargets(orgTestPeer0.(fab.Peer), orgTestPeer1.(fab.Peer)), ledger.WithMinTargets(2))
	if err != nil {
		t.Fatalf("QueryBlock return error: %v", err)
	}
	if block == nil {
		t.Fatalf("Block info not available")
	}
	// Get transaction info
	transactionInfo, err := ledgerClient.QueryTransaction(transactionID, ledger.WithTargets(orgTestPeer0.(fab.Peer), orgTestPeer1.(fab.Peer)), ledger.WithMinTargets(2))
	if err != nil {
		t.Fatalf("QueryTransaction return error: %v", err)
	}
	if transactionInfo.TransactionEnvelope == nil {
		t.Fatalf("Transaction info missing")
	}
}

func createChannel(org1AdminUser msp.SigningIdentity, org2AdminUser msp.SigningIdentity, mc *multiorgContext, t *testing.T) {
	// Channel management client is responsible for managing channels (create/update channel)
	chMgmtClient, err := resmgmt.New(mc.ordererClientContext)
	require.NoError(t, err, "failed to get a new channel management client")

	// create a channel for orgchannel.tx
	req := resmgmt.SaveChannelRequest{ChannelID: channelID,
		ChannelConfigPath: path.Join("../../../", metadata.ChannelConfigPath, "orgchannel.tx"),
		SigningIdentities: []msp.SigningIdentity{org1AdminUser, org2AdminUser}}
	txID, err := chMgmtClient.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.Nil(t, err, "error should be nil for SaveChannel of orgchannel")
	require.NotEmpty(t, txID, "transaction ID should be populated")

	//do the same get ch client and create channel for each anchor peer as well (first for Org1MSP)
	chMgmtClient, err = resmgmt.New(mc.org1AdminClientContext)
	require.NoError(t, err, "failed to get a new channel management client for org1Admin")
	req = resmgmt.SaveChannelRequest{ChannelID: channelID,
		ChannelConfigPath: path.Join("../../../", metadata.ChannelConfigPath, "orgchannelOrg1MSPanchors.tx"),
		SigningIdentities: []msp.SigningIdentity{org1AdminUser}}
	txID, err = chMgmtClient.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.Nil(t, err, "error should be nil for SaveChannel for anchor peer 1")
	require.NotEmpty(t, txID, "transaction ID should be populated for anchor peer 1")

	// lastly create channel for Org2MSP anchor peer
	chMgmtClient, err = resmgmt.New(mc.org2AdminClientContext)
	require.NoError(t, err, "failed to get a new channel management client for org2Admin")
	req = resmgmt.SaveChannelRequest{ChannelID: channelID,
		ChannelConfigPath: path.Join("../../../", metadata.ChannelConfigPath, "orgchannelOrg2MSPanchors.tx"),
		SigningIdentities: []msp.SigningIdentity{org2AdminUser}}
	txID, err = chMgmtClient.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.Nil(t, err, "error should be nil for SaveChannel for anchor peer 2")
	require.NotEmpty(t, txID, "transaction ID should be populated for anchor peer 2")
}

func testCCPolicy(chClientOrg2User *channel.Client, t *testing.T, ccName string) {
	_, err := chClientOrg2User.Execute(channel.Request{ChaincodeID: ccName, Fcn: "invoke", Args: integration.ExampleCCTxArgs()}, channel.WithTargets(orgTestPeer1),
		channel.WithRetry(retry.DefaultChannelOpts))
	if err == nil {
		t.Fatalf("Should have failed to move funds due to cc policy")
	}
	// Org2 user moves funds (cc policy ok since we have provided peers for both Orgs)
	_, err = chClientOrg2User.Execute(channel.Request{ChaincodeID: ccName, Fcn: "invoke", Args: integration.ExampleCCTxArgs()}, channel.WithTargets(orgTestPeer0, orgTestPeer1),
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}
}

func upgradeCC(ccPkg *resource.CCPackage, org1ResMgmt *resmgmt.Client, t *testing.T, org2ResMgmt *resmgmt.Client, ccName string) {
	installCCReq := resmgmt.InstallCCRequest{Name: ccName, Path: "github.com/example_cc", Version: "1", Package: ccPkg}
	// Install example cc version '1' to Org1 peers
	_, err := org1ResMgmt.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	require.Nil(t, err, "error should be nil for InstallCC version '1' or Org1 peers")

	// Install example cc version '1' to Org2 peers
	_, err = org2ResMgmt.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	require.Nil(t, err, "error should be nil for InstallCC version '1' or Org2 peers")

	// New chaincode policy (both orgs have to approve)
	org1Andorg2Policy, err := cauthdsl.FromString("AND ('Org1MSP.member','Org2MSP.member')")
	require.Nil(t, err, "error should be nil for getting cc policy with both orgs to approve")

	// Org1 resource manager will instantiate 'example_cc' version 1 on 'orgchannel'
	upgradeResp, err := org1ResMgmt.UpgradeCC(channelID, resmgmt.UpgradeCCRequest{Name: ccName, Path: "github.com/example_cc", Version: "1", Args: integration.ExampleCCUpgradeArgs(), Policy: org1Andorg2Policy})
	require.Nil(t, err, "error should be nil for UpgradeCC version '1' on 'orgchannel'")
	require.NotEmpty(t, upgradeResp, "transaction response should be populated")
}

func moveFunds(chClientOrgUser *channel.Client, t *testing.T, ccName string) fab.TransactionID {
	response, err := chClientOrgUser.Execute(channel.Request{ChaincodeID: ccName, Fcn: "invoke", Args: integration.ExampleCCTxArgs()}, channel.WithTargets(orgTestPeer1),
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}
	if response.ChaincodeStatus == 0 {
		t.Fatalf("Expected ChaincodeStatus")
	}
	if response.Responses[0].ChaincodeStatus != response.ChaincodeStatus {
		t.Fatalf("Expected the chaincode status returned by successful Peer Endorsement to be same as Chaincode status for client response")
	}
	return response.TransactionID
}

func getBlockchainInfo(ledgerClient *ledger.Client, t *testing.T) *fab.BlockchainInfoResponse {
	channelCfg, err := ledgerClient.QueryConfig(ledger.WithTargets(orgTestPeer0, orgTestPeer1), ledger.WithMinTargets(2))
	if err != nil {
		t.Fatalf("QueryConfig return error: %v", err)
	}
	if len(channelCfg.Orderers()) == 0 {
		t.Fatalf("Failed to retrieve channel orderers")
	}
	expectedOrderer := "orderer.example.com"
	if !strings.Contains(channelCfg.Orderers()[0], expectedOrderer) {
		t.Fatalf("Expecting %s, got %s", expectedOrderer, channelCfg.Orderers()[0])
	}
	ledgerInfoBefore, err := ledgerClient.QueryInfo(ledger.WithTargets(orgTestPeer0, orgTestPeer1), ledger.WithMinTargets(2), ledger.WithMaxTargets(3))
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}
	// Test Query Block by Hash - retrieve current block by hash
	block, err := ledgerClient.QueryBlockByHash(ledgerInfoBefore.BCI.CurrentBlockHash, ledger.WithTargets(orgTestPeer0.(fab.Peer), orgTestPeer1.(fab.Peer)), ledger.WithMinTargets(2))
	if err != nil {
		t.Fatalf("QueryBlockByHash return error: %v", err)
	}
	if block == nil {
		t.Fatalf("Block info not available")
	}
	return ledgerInfoBefore
}

func queryCC(chClientOrg1User *channel.Client, t *testing.T, ccName string) []byte {
	response, err := chClientOrg1User.Query(channel.Request{ChaincodeID: ccName, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()},
		channel.WithRetry(retry.DefaultChannelOpts))

	require.NoError(t, err, "Failed to query funds")

	require.NotZero(t, response.ChaincodeStatus, "Expected ChaincodeStatus")

	require.Equal(t, response.ChaincodeStatus, response.Responses[0].ChaincodeStatus, "Expected the chaincode status returned by successful Peer Endorsement to be same as Chaincode status for client response")

	return response.Payload
}

func verifyErrorFromCC(chClientOrg1User *channel.Client, t *testing.T, ccName string) {
	r, err := chClientOrg1User.Query(channel.Request{ChaincodeID: ccName, Fcn: "DUMMY_FUNCTION", Args: integration.ExampleCCQueryArgs()},
		channel.WithRetry(retry.DefaultChannelOpts))
	t.Logf("verifyErrorFromCC err: %s ***** responses: %v", err, r)

	require.Error(t, err, "Should have failed with dummy function")
	s, ok := status.FromError(err)
	t.Logf("verifyErrorFromCC status.FromError s: %s, ok: %t", s, ok)

	require.True(t, ok, "expected status error")
	require.Equal(t, s.Code, int32(status.MultipleErrors))

	for _, err := range err.(multi.Errors) {
		s, ok := status.FromError(err)
		require.True(t, ok, "expected status error")
		require.EqualValues(t, int32(500), s.Code)
		require.Equal(t, status.ChaincodeStatus, s.Group)
	}
}

func createCC(t *testing.T, org1ResMgmt *resmgmt.Client, org2ResMgmt *resmgmt.Client, ccPkg *resource.CCPackage, ccName string) {
	installCCReq := resmgmt.InstallCCRequest{Name: ccName, Path: "github.com/example_cc", Version: "0", Package: ccPkg}

	// Install example cc to Org1 peers
	_, err := org1ResMgmt.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	require.NoError(t, err, "InstallCC for Org1 failed")

	// Install example cc to Org2 peers
	_, err = org2ResMgmt.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	require.NoError(t, err, "InstallCC for Org2 failed")

	instantiateCC(t, org1ResMgmt, []string{"peer0.org1.example.com"}, ccName)
	// instantiateCC above will propagate the call to the other peers through gossip, so no need to call it on the other peers.

	// Verify that example CC is instantiated on Org1 peer
	chaincodeQueryResponse, err := org1ResMgmt.QueryInstantiatedChaincodes(channelID, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	require.NoError(t, err, "QueryInstantiatedChaincodes return error")

	found := false
	for _, chaincode := range chaincodeQueryResponse.Chaincodes {
		if chaincode.Name == ccName {
			found = true
		}
	}
	require.True(t, found, "QueryInstantiatedChaincodes failed to find instantiated '%s' chaincode", ccName)

}

func instantiateCC(t *testing.T, resMgmt *resmgmt.Client, targets []string, ccName string) {
	// Set up chaincode policy to 'any of two msps'
	ccPolicy := cauthdsl.SignedByAnyMember([]string{"Org1MSP", "Org2MSP"})
	// Org1 resource manager will instantiate 'example_cc' on 'orgchannel'
	instantiateResp, err := resMgmt.InstantiateCC(
		channelID,
		resmgmt.InstantiateCCRequest{
			Name:    ccName,
			Path:    "github.com/example_cc",
			Version: "0",
			Args:    integration.ExampleCCInitArgs(),
			Policy:  ccPolicy,
		},
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		resmgmt.WithTargetEndpoints(targets...),
	)

	require.Nil(t, err, "error should be nil for instantiateCC")
	require.NotEmpty(t, instantiateResp, "transaction response should be populated for instantateCC")
}

func testWithOrg2(t *testing.T, expectedValue int, ccName string) int {
	// Specify user that will be used by dynamic selection service (to retrieve chanincode policy information)
	// This user has to have privileges to query lscc for chaincode data
	mychannelUser := selection.ChannelUser{ChannelID: channelID, Username: "User1", OrgName: "Org1"}

	// Create SDK setup for channel client with dynamic selection
	sdk, err := fabsdk.New(integration.ConfigBackend,
		fabsdk.WithServicePkg(&DynamicSelectionProviderFactory{ChannelUsers: []selection.ChannelUser{mychannelUser}}))
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}
	defer sdk.Close()

	//prepare contexts
	org2ChannelClientContext := sdk.ChannelContext(channelID, fabsdk.WithUser(org2User), fabsdk.WithOrg(org2))

	// Create new client that will use dynamic selection
	chClientOrg2User, err := channel.New(org2ChannelClientContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client for Org2 user: %s", err)
	}

	// Org2 user moves funds (dynamic selection will inspect chaincode policy to determine endorsers)
	_, err = chClientOrg2User.Execute(channel.Request{ChaincodeID: ccName, Fcn: "invoke", Args: integration.ExampleCCTxArgs()},
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	expectedValue++
	return expectedValue
}

func verifyWithOrg1(t *testing.T, sdk *fabsdk.FabricSDK, expectedValue int, ccName string) {
	//prepare context
	org1ChannelClientContext := sdk.ChannelContext(channelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1))

	// Org1 user connects to 'orgchannel'
	chClientOrg1User, err := channel.New(org1ChannelClientContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client for Org1 user: %s", err)
	}

	verifyValue(t, chClientOrg1User, expectedValue, ccName)
}

func verifyValue(t *testing.T, chClient *channel.Client, expected int, ccName string) {
	// Assert that funds have changed value on org1 peer
	var valueInt int
	for i := 0; i < pollRetries; i++ {
		// Query final value on org1 peer
		response, err := chClient.Query(channel.Request{ChaincodeID: ccName, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()}, channel.WithTargets(orgTestPeer0),
			channel.WithRetry(retry.DefaultChannelOpts))
		if err != nil {
			t.Fatalf("Failed to query funds after transaction: %s", err)
		}
		// If value has not propogated sleep with exponential backoff
		valueInt, _ = strconv.Atoi(string(response.Payload))
		if expected != valueInt {
			backoffFactor := math.Pow(2, float64(i))
			time.Sleep(time.Millisecond * 50 * time.Duration(backoffFactor))
		} else {
			break
		}
	}
	if expected != valueInt {
		t.Fatalf("Org2 'move funds' transaction result was not propagated to Org1. Expected %d, got: %d",
			(expected), valueInt)
	}

}

func loadOrgPeers(t *testing.T, ctxProvider contextAPI.ClientProvider) {

	ctx, err := ctxProvider()
	if err != nil {
		t.Fatalf("context creation failed: %s", err)
	}

	org1Peers, err := ctx.EndpointConfig().PeersConfig(org1)
	if err != nil {
		t.Fatal(err)
	}

	org2Peers, err := ctx.EndpointConfig().PeersConfig(org2)
	if err != nil {
		t.Fatal(err)
	}

	orgTestPeer0, err = ctx.InfraProvider().CreatePeerFromConfig(&fab.NetworkPeer{PeerConfig: org1Peers[0]})
	if err != nil {
		t.Fatal(err)
	}

	orgTestPeer1, err = ctx.InfraProvider().CreatePeerFromConfig(&fab.NetworkPeer{PeerConfig: org2Peers[0]})
	if err != nil {
		t.Fatal(err)
	}
}

// DynamicSelectionProviderFactory is configured with dynamic (endorser) selection provider
type DynamicSelectionProviderFactory struct {
	defsvc.ProviderFactory
	ChannelUsers []selection.ChannelUser
}

// CreateSelectionProvider returns a new implementation of dynamic selection provider
func (f *DynamicSelectionProviderFactory) CreateSelectionProvider(config fab.EndpointConfig) (fab.SelectionProvider, error) {
	return selection.New(config, f.ChannelUsers)
}
