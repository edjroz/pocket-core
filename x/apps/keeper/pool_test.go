package keeper

import (
	"fmt"
	"github.com/pokt-network/pocket-core/x/apps/types"
	sdk "github.com/pokt-network/posmint/types"
	"github.com/pokt-network/posmint/x/supply/exported"
	"strings"
	"testing"
)

func TestPool_CoinsFromUnstakedToStaked(t *testing.T) {
	application := getBondedApplication()
	applicationAddress := application.Address

	tests := []struct {
		name        string
		want    string
		application types.Application
		amount      sdk.Int
		panics      bool
	}{
		{
			name:        "stake coins on pool",
			application: types.Application{Address: applicationAddress},
			amount:      sdk.NewInt(10),
			panics:      false,
		},
		{
			name:        "panics if negative ammount",
			application: types.Application{Address: applicationAddress},
			amount:      sdk.NewInt(-1),
			want:    fmt.Sprintf("negative coin amount: -1"),
			panics:      true,
		},
		{name: "panics if no supply is set",
			application: types.Application{Address: applicationAddress},
			want:    fmt.Sprintf("insufficient account funds"),
			amount:      sdk.NewInt(10),
			panics:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)

			switch tt.panics {
			case true:
				defer func() {
					if err := recover().(error); !strings.Contains(err.Error(), tt.want) {
						t.Errorf("KeeperCoins.FromUnstakedToStaked()= %v, want %v", err.Error(), tt.want)
					}
				}()
				if strings.Contains(tt.name, "setup") {
					addMintedCoinsToModule(t, context, &keeper, types.StakedPoolName)
					sendFromModuleToAccount(t, context, &keeper, types.StakedPoolName, tt.application.Address, sdk.NewInt(100000000000))
				}
				keeper.coinsFromUnstakedToStaked(context, tt.application, tt.amount)
			default:
				addMintedCoinsToModule(t, context, &keeper, types.StakedPoolName)
				sendFromModuleToAccount(t, context, &keeper, types.StakedPoolName, tt.application.Address, sdk.NewInt(100000000000))
				keeper.coinsFromUnstakedToStaked(context, tt.application, tt.amount)
				if got := keeper.GetStakedTokens(context);!tt.amount.Add(sdk.NewInt(100000000000)).Equal(got) {
					t.Errorf("KeeperCoins.FromUnstakedToStaked()= %v, want %v", got, tt.amount.Add(sdk.NewInt(100000000000)))
				}
			}
		})
	}
}

func TestPool_CoinsFromStakedToUnstaked(t *testing.T) {
	application := getBondedApplication()
	applicationAddress := application.Address

	tests := []struct {
		name        string
		amount      sdk.Int
		want    string
		application types.Application
		panics      bool
	}{
		{
			name:        "unstake coins from pool",
			application: types.Application{Address: applicationAddress, StakedTokens: sdk.NewInt(10)},
			amount:      sdk.NewInt(110),
			panics:      false,
		},
		{
			name:        "panics if negative ammount",
			application: types.Application{Address: applicationAddress, StakedTokens: sdk.NewInt(-1)},
			amount:      sdk.NewInt(-1),
			want:    fmt.Sprintf("negative coin amount: -1"),
			panics:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)

			switch tt.panics {
			case true:
				defer func() {
					if err := recover().(error); !strings.Contains(err.Error(), tt.want) {
						t.Errorf("KeeperCoins.FromStakedToUnstaked()= %v, want %v", err.Error(), tt.want)
					}
				}()
				if strings.Contains(tt.name, "setup") {
					addMintedCoinsToModule(t, context, &keeper, types.StakedPoolName)
					sendFromModuleToAccount(t, context, &keeper, types.StakedPoolName, tt.application.Address, sdk.NewInt(100))
				}
				keeper.coinsFromStakedToUnstaked(context, tt.application)
			default:
				addMintedCoinsToModule(t, context, &keeper, types.StakedPoolName)
				sendFromModuleToAccount(t, context, &keeper, types.StakedPoolName, tt.application.Address, sdk.NewInt(100))
				keeper.coinsFromStakedToUnstaked(context, tt.application)
				if got := keeper.GetUnstakedTokens(context); !tt.amount.Equal(got) {
					t.Errorf("KeeperCoins.FromStakedToUnstaked()= %v, want %v", got, tt.amount)
				}
			}
		})
	}
}

func TestPool_BurnStakedTokens(t *testing.T) {
	application := getBondedApplication()
	applicationAddress := application.Address

	supplySize := sdk.NewInt(100000000000)
	tests := []struct {
		name        string
		expected    string
		application types.Application
		burnAmount  sdk.Int
		amount      sdk.Int
		errs        bool
	}{
		{
			name:        "burn coins from pool",
			application: types.Application{Address: applicationAddress},
			burnAmount:  sdk.NewInt(5),
			amount:      sdk.NewInt(10),
			errs:        false,
		},
		{
			name:        "errs trying to burn from pool",
			application: types.Application{Address: applicationAddress},
			burnAmount:  sdk.NewInt(-1),
			amount:      sdk.NewInt(10),
			errs:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)

			switch tt.errs {
			case true:
				addMintedCoinsToModule(t, context, &keeper, types.StakedPoolName)
				sendFromModuleToAccount(t, context, &keeper, types.StakedPoolName, tt.application.Address, supplySize)
				keeper.coinsFromUnstakedToStaked(context, tt.application, tt.amount)
				if err := keeper.burnStakedTokens(context, tt.burnAmount); err != nil {
					t.Errorf("KeeperCoins.BurnStakedTokens()= %v, want nil", err)
				}
			default:
				addMintedCoinsToModule(t, context, &keeper, types.StakedPoolName)
				sendFromModuleToAccount(t, context, &keeper, types.StakedPoolName, tt.application.Address, supplySize)
				keeper.coinsFromUnstakedToStaked(context, tt.application, tt.amount)
				err := keeper.burnStakedTokens(context, tt.burnAmount)
				if err != nil {
					t.Fail()
				}
				if got := keeper.GetStakedTokens(context); !tt.amount.Sub(tt.burnAmount).Add(supplySize).Equal(got) {
					t.Errorf("KeeperCoins.BurnStakedTokens()= %v, want %v", got, tt.amount.Sub(tt.burnAmount).Add(supplySize))
				}
			}
		})
	}
}

func TestPool_GetFeePool(t *testing.T) {
	tests := []struct{
		name string
	}{
		{
			"gets fee pool",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T){
			context, _, keeper := createTestInput(t, true)
			got := keeper.getFeePool(context)

			if _, ok := got.(exported.ModuleAccountI); !ok {
				t.Errorf("KeeperPool.getFeePool()= %v", ok)
			}
		})
	}
}

func TestPool_StakedRatio(t *testing.T) {
	application := getBondedApplication()
	applicationAddress := application.Address

	tests := []struct{
		name string
		amount sdk.Dec
		address sdk.ValAddress
	}{
		{"return 0 if stake supply is lower than 0", sdk.ZeroDec(), applicationAddress},
		{"return supply", sdk.NewDec(1), applicationAddress},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T){
			context, _, keeper := createTestInput(t, true)
			if !tt.amount.Equal(sdk.ZeroDec()) {
				addMintedCoinsToModule(t, context, &keeper, types.StakedPoolName)
			}

			if got := keeper.StakedRatio(context); !got.Equal(tt.amount) {
				t.Errorf("KeeperPool.StakedRatio()= %v, %v", got.String(), tt.amount.String())
			}
		})
	}
}
