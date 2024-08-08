package keeper

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"

	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/x/upgrade/types"
)

type storedUpgrade struct {
	name   string
	height int64
}

func encodeOldDoneKey(upgrade storedUpgrade) []byte {
	return append([]byte{types.DoneByte}, []byte(upgrade.name)...)
}

func TestMigrateDoneUpgradeKeys(t *testing.T) {
	ctx := coretesting.Context()
	storeService := coretesting.KVStoreService(ctx, "upgrade")
	store := storeService.OpenKVStore(ctx)

	testCases := []struct {
		name     string
		upgrades []storedUpgrade
	}{
		{
			name: "valid upgrades",
			upgrades: []storedUpgrade{
				{name: "some-other-upgrade", height: 1},
				{name: "test02", height: 2},
				{name: "test01", height: 3},
			},
		},
	}

	for _, tc := range testCases {
		for _, upgrade := range tc.upgrades {
			bz := make([]byte, 8)
			binary.BigEndian.PutUint64(bz, uint64(upgrade.height))
			oldKey := encodeOldDoneKey(upgrade)
			require.NoError(t, store.Set(oldKey, bz))
		}

		err := migrateDoneUpgradeKeys(ctx, storeService)
		require.NoError(t, err)

		for _, upgrade := range tc.upgrades {
			newKey := encodeDoneKey(upgrade.name, upgrade.height)
			oldKey := encodeOldDoneKey(upgrade)
			v, err := store.Get(oldKey)
			require.Nil(t, v)
			require.NoError(t, err)

			nv, err := store.Get(newKey)
			require.Equal(t, []byte{1}, nv)
			require.NoError(t, err)
		}
	}
}
