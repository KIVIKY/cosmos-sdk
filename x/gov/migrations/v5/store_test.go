package v5_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/x/bank"
	"cosmossdk.io/x/gov"
	v5 "cosmossdk.io/x/gov/migrations/v5"
	v1 "cosmossdk.io/x/gov/types/v1"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

func TestMigrateStore(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, gov.AppModule{}, bank.AppModule{}).Codec

	ctx := coretesting.Context()
	storeService := coretesting.KVStoreService(ctx, "gov")
	store := storeService.OpenKVStore(ctx)
	sb := collections.NewSchemaBuilder(storeService)
	constitutionCollection := collections.NewItem(sb, v5.ConstitutionKey, "constitution", collections.StringValue)

	var params v1.Params
	bz, err := store.Get(v5.ParamsKey)
	require.NoError(t, err)
	require.NoError(t, cdc.Unmarshal(bz, &params))
	require.NotNil(t, params)
	require.Equal(t, "", params.ExpeditedThreshold)
	require.Equal(t, (*time.Duration)(nil), params.ExpeditedVotingPeriod)

	// Run migrations.
	err = v5.MigrateStore(ctx, storeService, cdc, constitutionCollection)
	require.NoError(t, err)

	// Check params
	bz, err = store.Get(v5.ParamsKey)
	require.NoError(t, err)
	require.NoError(t, cdc.Unmarshal(bz, &params))
	require.NotNil(t, params)
	require.Equal(t, v1.DefaultParams().ExpeditedMinDeposit, params.ExpeditedMinDeposit)
	require.Equal(t, v1.DefaultParams().ExpeditedThreshold, params.ExpeditedThreshold)
	require.Equal(t, v1.DefaultParams().ExpeditedVotingPeriod, params.ExpeditedVotingPeriod)
	require.Equal(t, v1.DefaultParams().MinDepositRatio, params.MinDepositRatio)

	// Check constitution
	result, err := constitutionCollection.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, "This chain has no constitution.", result)
}
