package service

import (
	"strings"
	"testing"

	"ticket-katon-backend/internal/domain"
)

func TestBuildSeatPlanCounts(t *testing.T) {
	tests := []struct {
		name   string
		vtype  domain.TransportType
		layout map[string]any
		want   int
	}{
		{"bus 2x2", domain.TransportBus, map[string]any{"type": "2x2"}, 40},
		{"bus 2x1", domain.TransportBus, map[string]any{"type": "2x1"}, 30},
		{"launch", domain.TransportLaunch, map[string]any{"type": "cabin_and_deck"}, 54},
		{
			"train two coaches", domain.TransportTrain,
			map[string]any{"coaches": []any{
				map[string]any{"coach_name": "KA", "class": "SNIGDHA", "total_seats": float64(60)},
				map[string]any{"coach_name": "KHA", "class": "SHOVON_CHAIR", "total_seats": float64(60)},
			}},
			120,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			seats, err := buildSeatPlan(tc.vtype, tc.layout, 750)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(seats) != tc.want {
				t.Fatalf("got %d seats, want %d", len(seats), tc.want)
			}

			seen := map[string]bool{}
			for _, s := range seats {
				if seen[s.SeatNumber] {
					t.Fatalf("duplicate seat number %s", s.SeatNumber)
				}
				seen[s.SeatNumber] = true
				if s.Fare <= 0 {
					t.Fatalf("seat %s has fare %v", s.SeatNumber, s.Fare)
				}
			}
		})
	}
}

func TestBuildSeatPlanFares(t *testing.T) {
	seats, err := buildSeatPlan(domain.TransportTrain, map[string]any{"coaches": []any{
		map[string]any{"coach_name": "KA", "class": "SNIGDHA", "total_seats": float64(2)},
	}}, 400)
	if err != nil {
		t.Fatal(err)
	}
	if seats[0].SeatNumber != "KA-1" || seats[0].Fare != 560 {
		t.Fatalf("got %+v, want KA-1 at 560 (Snigdha is 1.4x)", seats[0])
	}
}

func TestBuildSeatPlanRejectsBadInput(t *testing.T) {
	if _, err := buildSeatPlan(domain.TransportTrain, map[string]any{}, 100); err == nil {
		t.Error("train without coaches should fail")
	}
	if _, err := buildSeatPlan(domain.TransportType("AIR"), map[string]any{}, 100); err == nil {
		t.Error("unsupported vehicle type should fail")
	}
}

func TestBuildLayoutConfig(t *testing.T) {
	good := &domain.CreateVehicleRequest{Type: domain.TransportBus}
	cfg, err := buildLayoutConfig(good)
	if err != nil || cfg["type"] != "2x2" {
		t.Fatalf("bus default should be 2x2, got %v, %v", cfg, err)
	}

	cases := []struct {
		name string
		req  domain.CreateVehicleRequest
		msg  string
	}{
		{"bad bus layout", domain.CreateVehicleRequest{Type: domain.TransportBus, BusLayout: "3x3"}, "2x2 or 2x1"},
		{"train with no coaches", domain.CreateVehicleRequest{Type: domain.TransportTrain}, "between 1 and 12"},
		{"bad coach name", domain.CreateVehicleRequest{Type: domain.TransportTrain, Coaches: []domain.CoachConfig{{CoachName: "K1", Class: "SNIGDHA", TotalSeats: 10}}}, "1 to 4 letters"},
		{"bad class", domain.CreateVehicleRequest{Type: domain.TransportTrain, Coaches: []domain.CoachConfig{{CoachName: "KA", Class: "FIRST", TotalSeats: 10}}}, "SNIGDHA or SHOVON_CHAIR"},
		{"too many seats", domain.CreateVehicleRequest{Type: domain.TransportTrain, Coaches: []domain.CoachConfig{{CoachName: "KA", Class: "SNIGDHA", TotalSeats: 500}}}, "1 and 120"},
		{"duplicate coach", domain.CreateVehicleRequest{Type: domain.TransportTrain, Coaches: []domain.CoachConfig{
			{CoachName: "KA", Class: "SNIGDHA", TotalSeats: 10}, {CoachName: "ka", Class: "SNIGDHA", TotalSeats: 10}}}, "listed twice"},
		{"unsupported type", domain.CreateVehicleRequest{Type: domain.TransportType("AIR")}, "not supported"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := buildLayoutConfig(&tc.req)
			if err == nil || !strings.Contains(err.Error(), tc.msg) {
				t.Fatalf("got %v, want error containing %q", err, tc.msg)
			}
		})
	}
}

func TestNameValidators(t *testing.T) {
	for _, ok := range []string{"Dhaka", "Cox's Bazar", "Chattogram", "Sylhet-Sadar"} {
		if !cityRe.MatchString(ok) {
			t.Errorf("city %q should be valid", ok)
		}
	}
	for _, bad := range []string{"", "1Dhaka", "Dhaka<script>", " "} {
		if cityRe.MatchString(bad) {
			t.Errorf("city %q should be invalid", bad)
		}
	}
	for _, ok := range []string{"DHAKA-METRO-BA-11-1234", "DEMO-BUS-101", "KA 12"} {
		if !regNoRe.MatchString(ok) {
			t.Errorf("registration %q should be valid", ok)
		}
	}
	if regNoRe.MatchString("A") || regNoRe.MatchString("bad!reg") {
		t.Error("short or unusual registration numbers should be invalid")
	}
	if got := cleanSpaces("  Cox's   Bazar "); got != "Cox's Bazar" {
		t.Errorf("cleanSpaces gave %q", got)
	}
}
