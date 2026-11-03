package tests

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/phosphoriat/invest-pool-go/internal/models"
	"github.com/phosphoriat/invest-pool-go/internal/services"
)

// TestAllocate Check proportional payouts, rounding ties and large integer products.
func TestAllocate(t *testing.T) {
	tests := []struct {
		name   string
		shares []models.Investment
		profit int64
		want   []models.Payout
	}{
		{"60/40", []models.Investment{{InvestorID: "a", Amount: 600000}, {InvestorID: "b", Amount: 400000}}, 100000, []models.Payout{{InvestorID: "a", Amount: 60000}, {InvestorID: "b", Amount: 40000}}},
		{"tie sorted by ID", []models.Investment{{InvestorID: "c", Amount: 1}, {InvestorID: "b", Amount: 1}, {InvestorID: "a", Amount: 1}}, 2, []models.Payout{{InvestorID: "a", Amount: 1}, {InvestorID: "b", Amount: 1}, {InvestorID: "c", Amount: 0}}},
		{"largest remainder", []models.Investment{{InvestorID: "a", Amount: 1}, {InvestorID: "b", Amount: 2}}, 2, []models.Payout{{InvestorID: "a", Amount: 1}, {InvestorID: "b", Amount: 1}}},
		{"aggregate investor", []models.Investment{{InvestorID: "a", Amount: 1}, {InvestorID: "b", Amount: 2}, {InvestorID: "a", Amount: 1}}, 10, []models.Payout{{InvestorID: "a", Amount: 5}, {InvestorID: "b", Amount: 5}}},
		{"large product", []models.Investment{{InvestorID: "a", Amount: services.MaxMoney / 2}, {InvestorID: "b", Amount: services.MaxMoney / 2}}, services.MaxMoney, []models.Payout{{InvestorID: "a", Amount: services.MaxMoney / 2}, {InvestorID: "b", Amount: services.MaxMoney / 2}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := services.Allocate(tt.shares, tt.profit)
			if err != nil || !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v %v, want %v", got, err, tt.want)
			}
		})
	}
}

// TestInvalidAllocation Reject invalid profits, empty pools and invalid investment amounts.
func TestInvalidAllocation(t *testing.T) {
	for _, profit := range []int64{-1, 0, services.MaxMoney + 1} {
		if _, err := services.Allocate([]models.Investment{{InvestorID: "a", Amount: 1}}, profit); err == nil {
			t.Fatal("accepted invalid profit")
		}
	}
	for _, shares := range [][]models.Investment{nil, {{InvestorID: "a", Amount: 0}}, {{InvestorID: "a", Amount: -1}}, {{InvestorID: "a", Amount: services.MaxMoney}, {InvestorID: "b", Amount: 1}}} {
		if _, err := services.Allocate(shares, 1); err == nil {
			t.Fatal("accepted invalid shares")
		}
	}
}

// TestConservationAndOrderIndependence Check profit conservation and stable payouts under reordered shares.
func TestConservationAndOrderIndependence(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	for n := 0; n < 1000; n++ {
		shares := []models.Investment{{InvestorID: "a", Amount: r.Int63n(1000000) + 1}, {InvestorID: "b", Amount: r.Int63n(1000000) + 1}, {InvestorID: "c", Amount: r.Int63n(1000000) + 1}}
		profit := r.Int63n(services.MaxMoney) + 1
		a, err := services.Allocate(shares, profit)
		if err != nil {
			t.Fatal(err)
		}
		var total int64
		for _, p := range a {
			total += p.Amount
		}
		if total != profit {
			t.Fatalf("%d != %d", total, profit)
		}
		r.Shuffle(len(shares), func(i, j int) { shares[i], shares[j] = shares[j], shares[i] })
		b, _ := services.Allocate(shares, profit)
		if !reflect.DeepEqual(a, b) {
			t.Fatal("order dependent")
		}
	}
}
