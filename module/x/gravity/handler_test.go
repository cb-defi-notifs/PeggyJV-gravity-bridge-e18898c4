package gravity

import (
	"bytes"
	"testing"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/gravity-bridge/module/x/gravity/keeper"
	"github.com/cosmos/gravity-bridge/module/x/gravity/types"
)

func TestHandleMsgSendToEth(t *testing.T) {
	var (
		userCosmosAddr, _               = sdk.AccAddressFromBech32("cosmos1990z7dqsvh8gthw9pa5sn4wuy2xrsd80mg5z6y")
		blockTime                       = time.Date(2020, 9, 14, 15, 20, 10, 0, time.UTC)
		blockHeight           int64     = 200
		denom                           = "gravity0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e"
		startingCoinAmount, _           = sdk.NewIntFromString("150000000000000000000") // 150 ETH worth, required to reach above u64 limit (which is about 18 ETH)
		sendAmount, _                   = sdk.NewIntFromString("50000000000000000000")  // 50 ETH
		feeAmount, _                    = sdk.NewIntFromString("5000000000000000000")   // 5 ETH
		startingCoins         sdk.Coins = sdk.Coins{sdk.NewCoin(denom, startingCoinAmount)}
		sendingCoin           sdk.Coin  = sdk.NewCoin(denom, sendAmount)
		feeCoin               sdk.Coin  = sdk.NewCoin(denom, feeAmount)
		ethDestination                  = "0x3c9289da00b02dC623d0D8D907619890301D26d4"
	)

	// we start by depositing some funds into the users balance to send
	input := keeper.CreateTestEnv(t)
	ctx := input.Context
	handler := NewHandler(input.GravityKeeper)
	input.BankKeeper.MintCoins(ctx, types.ModuleName, startingCoins)
	input.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, userCosmosAddr, startingCoins)
	balance1 := input.BankKeeper.GetAllBalances(ctx, userCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin(denom, startingCoinAmount)}, balance1)

	// send some coins
	msg := &types.MsgSendToEth{
		Sender:    userCosmosAddr.String(),
		EthDest:   ethDestination,
		Amount:    sendingCoin,
		BridgeFee: feeCoin}
	ctx = ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
	_, err := handler(ctx, msg)
	require.NoError(t, err)
	balance2 := input.BankKeeper.GetAllBalances(ctx, userCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin(denom, startingCoinAmount.Sub(sendAmount).Sub(feeAmount))}, balance2)

	// do the same thing again and make sure it works twice
	msg1 := &types.MsgSendToEth{
		Sender:    userCosmosAddr.String(),
		EthDest:   ethDestination,
		Amount:    sendingCoin,
		BridgeFee: feeCoin}
	ctx = ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
	_, err1 := handler(ctx, msg1)
	require.NoError(t, err1)
	balance3 := input.BankKeeper.GetAllBalances(ctx, userCosmosAddr)
	finalAmount3 := startingCoinAmount.Sub(sendAmount).Sub(sendAmount).Sub(feeAmount).Sub(feeAmount)
	assert.Equal(t, sdk.Coins{sdk.NewCoin(denom, finalAmount3)}, balance3)

	// now we should be out of coins and error
	msg2 := &types.MsgSendToEth{
		Sender:    userCosmosAddr.String(),
		EthDest:   ethDestination,
		Amount:    sendingCoin,
		BridgeFee: feeCoin}
	ctx = ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
	_, err2 := handler(ctx, msg2)
	require.Error(t, err2)
	balance4 := input.BankKeeper.GetAllBalances(ctx, userCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin(denom, finalAmount3)}, balance4)
}

func TestMsgDepositClaimSingleValidator(t *testing.T) {
	var (
		myOrchestratorAddr sdk.AccAddress = make([]byte, sdk.AddrLen)
		myCosmosAddr, _                   = sdk.AccAddressFromBech32("cosmos16ahjkfqxpp6lvfy9fpfnfjg39xr96qett0alj5")
		myValAddr                         = sdk.ValAddress(myOrchestratorAddr) // revisit when proper mapping is impl in keeper
		myNonce                           = uint64(1)
		anyETHAddr                        = "0xf9613b532673Cc223aBa451dFA8539B87e1F666D"
		tokenETHAddr                      = "0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e"
		myBlockTime                       = time.Date(2020, 9, 14, 15, 20, 10, 0, time.UTC)
		amountA, _                        = sdk.NewIntFromString("50000000000000000000")  // 50 ETH
		amountB, _                        = sdk.NewIntFromString("100000000000000000000") // 100 ETH
	)
	input := keeper.CreateTestEnv(t)
	ctx := input.Context
	input.GravityKeeper.StakingKeeper = keeper.NewStakingKeeperMock(myValAddr)
	input.GravityKeeper.SetOrchestratorValidator(ctx, myValAddr, myOrchestratorAddr)
	handler := NewHandler(input.GravityKeeper)

	myErc20 := types.ERC20Token{
		Amount:   amountA,
		Contract: tokenETHAddr,
	}

	msg := types.DepositClaim{
		EventNonce:          myNonce,
		TokenContract:       myErc20.Contract,
		Amount:              myErc20.Amount,
		EthereumSender:      anyETHAddr,
		CosmosReceiver:      myCosmosAddr.String(),
		OrchestratorAddress: myOrchestratorAddr.String(),
	}

	any, err := codectypes.NewAnyWithValue(&msg)
	require.NoError(t, err)

	ethClaim := types.MsgSubmitClaim{
		ClaimType: types.ClaimType_DEPOSIT,
		Claim:     any,
	}

	// when
	ctx = ctx.WithBlockTime(myBlockTime)
	_, err = handler(ctx, &ethClaim)
	EndBlocker(ctx, input.GravityKeeper)
	require.NoError(t, err)

	// and attestation persisted
	a := input.GravityKeeper.GetAttestation(ctx, myNonce, msg.ClaimHash())
	require.NotNil(t, a)
	// and vouchers added to the account
	balance := input.BankKeeper.GetAllBalances(ctx, myCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin("gravity0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e", amountA)}, balance)

	// Test to reject duplicate deposit
	// when
	ctx = ctx.WithBlockTime(myBlockTime)
	_, err = handler(ctx, &ethClaim)
	EndBlocker(ctx, input.GravityKeeper)
	// then
	require.Error(t, err)
	balance = input.BankKeeper.GetAllBalances(ctx, myCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin("gravity0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e", amountA)}, balance)

	// Test to reject skipped nonce
	msg = types.DepositClaim{
		EventNonce:          uint64(3),
		TokenContract:       tokenETHAddr,
		Amount:              amountA,
		EthereumSender:      anyETHAddr,
		CosmosReceiver:      myCosmosAddr.String(),
		OrchestratorAddress: myOrchestratorAddr.String(),
	}

	any, err = codectypes.NewAnyWithValue(&msg)
	require.NoError(t, err)

	ethClaim = types.MsgSubmitClaim{
		ClaimType: types.ClaimType_DEPOSIT,
		Claim:     any,
	}

	// when
	ctx = ctx.WithBlockTime(myBlockTime)
	_, err = handler(ctx, &ethClaim)
	EndBlocker(ctx, input.GravityKeeper)
	// then
	require.Error(t, err)
	balance = input.BankKeeper.GetAllBalances(ctx, myCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin("gravity0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e", amountA)}, balance)

	// Test to finally accept consecutive nonce
	msg = types.DepositClaim{
		EventNonce:          uint64(2),
		Amount:              amountA,
		TokenContract:       tokenETHAddr,
		EthereumSender:      anyETHAddr,
		CosmosReceiver:      myCosmosAddr.String(),
		OrchestratorAddress: myOrchestratorAddr.String(),
	}

	any, err = codectypes.NewAnyWithValue(&msg)
	require.NoError(t, err)

	ethClaim = types.MsgSubmitClaim{
		ClaimType: types.ClaimType_DEPOSIT,
		Claim:     any,
	}

	// when
	ctx = ctx.WithBlockTime(myBlockTime)
	_, err = handler(ctx, &ethClaim)
	EndBlocker(ctx, input.GravityKeeper)

	// then
	require.NoError(t, err)
	balance = input.BankKeeper.GetAllBalances(ctx, myCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewCoin("gravity0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e", amountB)}, balance)
}

func TestMsgDepositClaimsMultiValidator(t *testing.T) {
	var (
		orchestratorAddr1, _ = sdk.AccAddressFromBech32("cosmos1dg55rtevlfxh46w88yjpdd08sqhh5cc3xhkcej")
		orchestratorAddr2, _ = sdk.AccAddressFromBech32("cosmos164knshrzuuurf05qxf3q5ewpfnwzl4gj4m4dfy")
		orchestratorAddr3, _ = sdk.AccAddressFromBech32("cosmos193fw83ynn76328pty4yl7473vg9x86alq2cft7")
		myCosmosAddr, _      = sdk.AccAddressFromBech32("cosmos16ahjkfqxpp6lvfy9fpfnfjg39xr96qett0alj5")
		valAddr1             = sdk.ValAddress(orchestratorAddr1) // revisit when proper mapping is impl in keeper
		valAddr2             = sdk.ValAddress(orchestratorAddr2) // revisit when proper mapping is impl in keeper
		valAddr3             = sdk.ValAddress(orchestratorAddr3) // revisit when proper mapping is impl in keeper
		myNonce              = uint64(1)
		anyETHAddr           = "0xf9613b532673Cc223aBa451dFA8539B87e1F666D"
		tokenETHAddr         = "0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e"
		myBlockTime          = time.Date(2020, 9, 14, 15, 20, 10, 0, time.UTC)
	)
	input := keeper.CreateTestEnv(t)
	ctx := input.Context
	input.GravityKeeper.StakingKeeper = keeper.NewStakingKeeperMock(valAddr1, valAddr2, valAddr3)
	input.GravityKeeper.SetOrchestratorValidator(ctx, valAddr1, orchestratorAddr1)
	input.GravityKeeper.SetOrchestratorValidator(ctx, valAddr2, orchestratorAddr2)
	input.GravityKeeper.SetOrchestratorValidator(ctx, valAddr3, orchestratorAddr3)
	h := NewHandler(input.GravityKeeper)

	myErc20 := types.ERC20Token{
		Amount:   sdk.NewInt(12),
		Contract: tokenETHAddr,
	}

	msg := types.DepositClaim{
		EventNonce:          myNonce,
		TokenContract:       myErc20.Contract,
		Amount:              myErc20.Amount,
		EthereumSender:      anyETHAddr,
		CosmosReceiver:      myCosmosAddr.String(),
		OrchestratorAddress: orchestratorAddr1.String(),
	}
	any1, err := codectypes.NewAnyWithValue(&msg)
	require.NoError(t, err)

	ethClaim1 := types.MsgSubmitClaim{
		ClaimType: types.ClaimType_DEPOSIT,
		Claim:     any1,
	}

	msg2 := types.DepositClaim{
		EventNonce:          myNonce,
		TokenContract:       myErc20.Contract,
		Amount:              myErc20.Amount,
		EthereumSender:      anyETHAddr,
		CosmosReceiver:      myCosmosAddr.String(),
		OrchestratorAddress: orchestratorAddr2.String(),
	}

	any2, err := codectypes.NewAnyWithValue(&msg2)
	require.NoError(t, err)

	ethClaim2 := types.MsgSubmitClaim{
		ClaimType: types.ClaimType_DEPOSIT,
		Claim:     any2,
	}

	msg3 := types.DepositClaim{
		EventNonce:          myNonce,
		TokenContract:       myErc20.Contract,
		Amount:              myErc20.Amount,
		EthereumSender:      anyETHAddr,
		CosmosReceiver:      myCosmosAddr.String(),
		OrchestratorAddress: orchestratorAddr3.String(),
	}

	any3, err := codectypes.NewAnyWithValue(&msg3)
	require.NoError(t, err)

	ethClaim3 := types.MsgSubmitClaim{
		ClaimType: types.ClaimType_DEPOSIT,
		Claim:     any3,
	}

	// when
	ctx = ctx.WithBlockTime(myBlockTime)
	_, err = h(ctx, &ethClaim1)
	EndBlocker(ctx, input.GravityKeeper)
	require.NoError(t, err)
	// and attestation persisted
	a1 := input.GravityKeeper.GetAttestation(ctx, myNonce, msg.ClaimHash())
	require.NotNil(t, a1)
	// and vouchers not yet added to the account
	balance1 := input.BankKeeper.GetAllBalances(ctx, myCosmosAddr)
	assert.NotEqual(t, sdk.Coins{sdk.NewInt64Coin("gravity0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e", 12)}, balance1)

	// when
	ctx = ctx.WithBlockTime(myBlockTime)
	_, err = h(ctx, &ethClaim2)
	EndBlocker(ctx, input.GravityKeeper)
	require.NoError(t, err)

	// and attestation persisted
	a2 := input.GravityKeeper.GetAttestation(ctx, myNonce, msg.ClaimHash())
	require.NotNil(t, a2)
	// and vouchers now added to the account
	balance2 := input.BankKeeper.GetAllBalances(ctx, myCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewInt64Coin("gravity0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e", 12)}, balance2)

	// when
	ctx = ctx.WithBlockTime(myBlockTime)
	_, err = h(ctx, &ethClaim3)
	EndBlocker(ctx, input.GravityKeeper)
	require.NoError(t, err)

	// and attestation persisted
	a3 := input.GravityKeeper.GetAttestation(ctx, myNonce, msg.ClaimHash())
	require.NotNil(t, a3)
	// and no additional added to the account
	balance3 := input.BankKeeper.GetAllBalances(ctx, myCosmosAddr)
	assert.Equal(t, sdk.Coins{sdk.NewInt64Coin("gravity0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e", 12)}, balance3)
}

func TestMsgSetOrchestratorAddresses(t *testing.T) {
	var (
		ethAddress                   = "0xb462864E395d88d6bc7C5dd5F3F5eb4cc2599255"
		cosmosAddress sdk.AccAddress = bytes.Repeat([]byte{0x1}, sdk.AddrLen)
		valAddress    sdk.ValAddress = bytes.Repeat([]byte{0x2}, sdk.AddrLen)
		blockTime                    = time.Date(2020, 9, 14, 15, 20, 10, 0, time.UTC)
		blockHeight   int64          = 200
	)
	input := keeper.CreateTestEnv(t)
	input.GravityKeeper.StakingKeeper = keeper.NewStakingKeeperMock(valAddress)
	ctx := input.Context
	handler := NewHandler(input.GravityKeeper)
	ctx = ctx.WithBlockTime(blockTime)

	msg := types.NewMsgSetDelegateKeys(valAddress, cosmosAddress, ethAddress)
	ctx = ctx.WithBlockTime(blockTime).WithBlockHeight(blockHeight)
	_, err := handler(ctx, msg)
	require.NoError(t, err)

	assert.Equal(t, input.GravityKeeper.GetEthAddress(ctx, valAddress), ethAddress)

	assert.Equal(t, input.GravityKeeper.GetOrchestratorValidator(ctx, cosmosAddress), valAddress)
}
