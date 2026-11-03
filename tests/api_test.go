package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"

	"github.com/phosphoriat/invest-pool-go/internal/database/sqlite"
	"github.com/phosphoriat/invest-pool-go/internal/handlers"
	"github.com/phosphoriat/invest-pool-go/internal/models"
	"github.com/phosphoriat/invest-pool-go/internal/repository"
	"github.com/phosphoriat/invest-pool-go/internal/routes"
	"github.com/phosphoriat/invest-pool-go/internal/services"
)

type fixture struct {
	app           *fiber.App
	repo          *repository.ProblemRepository
	events        *repository.EventRepository
	bus           *services.Bus
	investments   *services.InvestmentsService
	distributions *services.DistributionsService
}

// setup Create an isolated database and wire the application for API tests.
func setup(t *testing.T, path string) *fixture {
	t.Helper()
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	r := repository.NewProblemRepository(db)
	events := repository.NewEventRepository(db)
	bus := services.NewBus(events)
	bus.Subscribe(events.Audit)
	app := fiber.New(fiber.Config{ErrorHandler: handlers.ErrorHandler, BodyLimit: 16 * 1024})
	app.Use(recover.New())
	routes.SetupRoutes(app, r, events, bus)
	return &fixture{app: app, repo: r, events: events, bus: bus, investments: services.NewInvestmentsService(r, bus), distributions: services.NewDistributionsService(r, bus)}
}

// request Send an authenticated HTTP request and assert the expected response status.
func (f *fixture) request(t *testing.T, method, path, user, role string, body any, want int) []byte {
	t.Helper()
	var raw []byte
	switch v := body.(type) {
	case string:
		raw = []byte(v)
	case nil:
	default:
		raw, _ = json.Marshal(v)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", user)
	req.Header.Set("X-Role", role)
	res, err := f.app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	if res.StatusCode != want {
		t.Fatalf("%s %s got %d %s, want %d", method, path, res.StatusCode, data, want)
	}
	return data
}

// post Send a JSON command through the authenticated test client.
func (f *fixture) post(t *testing.T, path, user, role string, body any, want int) []byte {
	return f.request(t, "POST", path, user, role, body, want)
}

// project Publish a test problem and return its identifier.
func (f *fixture) project(t *testing.T) string {
	raw := f.post(t, "/api/problems", "region-tomsk", "region", map[string]any{"text": "Томская область, очистка воды"}, 200)
	var v map[string]string
	json.Unmarshal(raw, &v)
	return v["problem_id"]
}

// matched Propose a solution and confirm its match, including a safe confirmation retry.
func (f *fixture) matched(t *testing.T, id string) {
	path := "/api/problems/" + id
	proposed := result(t, f.post(t, path+"/solutions", "tech-water", "technologist", map[string]any{"text": "Очистка воды"}, 200))
	if proposed.ProblemID != id || proposed.SolutionID == "" {
		t.Fatalf("propose result %+v", proposed)
	}
	body := map[string]any{"solution_id": proposed.SolutionID}
	if r := result(t, f.post(t, path+"/match", "expert-1", "expert", body, 200)); r.Replayed || r.SolutionID != proposed.SolutionID {
		t.Fatalf("match result %+v", r)
	}
	if r := result(t, f.post(t, path+"/match", "expert-1", "expert", body, 200)); !r.Replayed {
		t.Fatalf("match retry must be marked as replayed: %+v", r)
	}
}

// result Decode a command response.
func result(t *testing.T, raw []byte) services.Result {
	t.Helper()
	var r services.Result
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatalf("bad command result %s: %v", raw, err)
	}
	return r
}

// TestHappyPathAndRejections Exercise the complete scenario, rejected commands and role-filtered data.
func TestHappyPathAndRejections(t *testing.T) {
	f := setup(t, filepath.Join(t.TempDir(), "test.db"))
	id := f.project(t)
	base := "/api/problems/" + id
	f.post(t, "/api/problems", "tech-water", "technologist", `{"text":"bad"}`, 403)
	f.post(t, base+"/solutions", "investor-a", "investor", `{"text":"bad"}`, 403)
	f.post(t, base+"/match", "region-tomsk", "region", `{"solution_id":"bad"}`, 403)
	f.post(t, base+"/pool", "region-tomsk", "region", `{"amount":1000000}`, 409)
	f.post(t, base+"/investments", "investor-a", "investor", `{"amount":1,"key":"pre-match"}`, 409)
	f.post(t, "/api/problems/missing/investments", "investor-a", "investor", `{"amount":1,"key":"missing"}`, 404)
	f.post(t, "/api/problems", "expert-1", "region", `{"text":"bad"}`, 403)
	f.post(t, "/api/problems", "unknown", "region", `{"text":"bad"}`, 401)
	f.matched(t, id)
	f.post(t, base+"/pool", "region-other", "region", `{"amount":1000000}`, 403)
	if r := result(t, f.post(t, base+"/pool", "region-tomsk", "region", `{"amount":1000000}`, 200)); r.PoolID == "" {
		t.Fatalf("pool result %+v", r)
	}
	f.post(t, base+"/profit", "region-tomsk", "region", `{"amount":100000}`, 409)
	f.post(t, base+"/investments", "investor-a", "investor", `{"amount":1,"key":""}`, 422)
	f.post(t, base+"/investments", "investor-a", "investor", `{"amount":1.5,"key":"float"}`, 400)
	f.post(t, base+"/investments", "investor-a", "investor", `{"amount":0,"key":"zero"}`, 422)
	f.post(t, base+"/investments", "investor-a", "investor", `{"amount":-1,"key":"negative"}`, 422)
	first := result(t, f.post(t, base+"/investments", "investor-a", "investor", `{"amount":600000,"key":"a-1"}`, 200))
	retry := result(t, f.post(t, base+"/investments", "investor-a", "investor", `{"amount":600000,"key":"a-1"}`, 200))
	if first.InvestmentID == "" || first.Total != 600000 || first.Replayed || retry.InvestmentID != first.InvestmentID || !retry.Replayed {
		t.Fatalf("investment results %+v %+v", first, retry)
	}
	f.post(t, base+"/investments", "investor-a", "investor", `{"amount":1,"key":"a-1"}`, 409)
	f.post(t, base+"/investments", "investor-b", "investor", `{"amount":400001,"key":"too-much"}`, 409)
	f.post(t, base+"/investments", "investor-b", "investor", `{"amount":400000,"key":"b-1"}`, 200)
	f.post(t, base+"/investments", "investor-c", "investor", `{"amount":1,"key":"late"}`, 409)
	for i := 0; i < 2; i++ {
		r := result(t, f.post(t, base+"/profit", "region-tomsk", "region", `{"amount":100000}`, 200))
		if r.Replayed != (i == 1) || r.Distribution == nil || len(r.Distribution.Payouts) != 2 ||
			r.Distribution.Payouts[0] != (models.Payout{InvestorID: "investor-a", Amount: 60000}) ||
			r.Distribution.Payouts[1] != (models.Payout{InvestorID: "investor-b", Amount: 40000}) {
			t.Fatalf("profit result #%d %+v", i, r)
		}
	}
	f.post(t, base+"/profit", "region-tomsk", "region", `{"amount":100001}`, 409)
	f.post(t, base+"/profit", "region-tomsk", "region", `{"amount":100000,"key":"k"}`, 400)
	ps, _ := f.repo.Problems(context.Background())
	p := ps[0]
	if p.Pool.Total != 1000000 || p.Pool.Status != "distributed" || len(p.Pool.Investments) != 2 {
		t.Fatalf("bad pool %+v", p.Pool)
	}
	pays := p.Pool.Distribution.Payouts
	if len(pays) != 2 || pays[0].Amount != 60000 || pays[1].Amount != 40000 {
		t.Fatalf("bad payouts %+v", pays)
	}
	if err := f.bus.Drain(context.Background()); err != nil {
		t.Fatal(err)
	}
	events, _ := f.events.Events(context.Background())
	counts := map[string]int{}
	for _, e := range events {
		counts[e.Name]++
		if strings.Contains(string(e.Payload), `"key"`) {
			t.Fatalf("operation key in %s payload: %s", e.Name, e.Payload)
		}
		if err := f.events.Audit(context.Background(), e); err != nil {
			t.Fatal(err)
		}
	}
	again, _ := f.events.Events(context.Background())
	if len(again) != len(events) {
		t.Fatal("duplicate audit")
	}
	if counts["ProblemPublished"] != 1 || counts["SolutionMatched"] != 1 || counts["InvestmentAccepted"] != 2 || counts["ProfitDistributed"] != 1 {
		t.Fatalf("events %+v", counts)
	}
	for _, u := range []string{"investor-a", "investor-b"} {
		raw := f.request(t, "GET", "/api/state", u, "investor", nil, 200)
		other := "investor-b"
		if u == other {
			other = "investor-a"
		}
		if strings.Contains(string(raw), other) {
			t.Fatalf("leaked %s: %s", other, raw)
		}
	}
	for _, pair := range [][2]string{{"region-other", "region"}, {"investor-c", "investor"}} {
		raw := f.request(t, "GET", "/api/state", pair[0], pair[1], nil, 200)
		if strings.Contains(string(raw), id) {
			t.Fatalf("foreign pool leaked: %s", raw)
		}
	}
	raw := f.request(t, "GET", "/api/state", "tech-other", "technologist", nil, 200)
	if strings.Contains(string(raw), p.Pool.ID) {
		t.Fatal("pool leaked to other technologist")
	}
	var techState struct {
		Projects []models.ProblemView `json:"projects"`
	}
	if err := json.Unmarshal(raw, &techState); err != nil || len(techState.Projects) != 1 {
		t.Fatalf("other technologist state %s", raw)
	}
	if v := techState.Projects[0]; !v.Matched || v.Match != nil || v.Pool != nil {
		t.Fatalf("other technologist must see only that the problem is matched: %s", raw)
	}
	f.post(t, base+"/solutions", "tech-other", "technologist", `{"text":"Поздно"}`, 409)
	for _, pair := range [][2]string{{"region-tomsk", "region"}, {"expert-1", "expert"}, {"investor-a", "investor"}, {"tech-water", "technologist"}} {
		raw := f.request(t, "GET", "/api/state", pair[0], pair[1], nil, 200)
		if strings.Contains(string(raw), `"key"`) {
			t.Fatalf("operation key leaked to %s: %s", pair[0], raw)
		}
	}
	f.request(t, "GET", "/api/events", "investor-a", "investor", nil, 403)
	raw = f.request(t, "GET", "/api/events", "region-other", "region", nil, 200)
	if string(raw) != "[]" {
		t.Fatalf("audit leak %s", raw)
	}
}

// TestConcurrentInvestmentsAndProfit Verify concurrent requests cannot overfund a pool or duplicate profit.
func TestConcurrentInvestmentsAndProfit(t *testing.T) {
	f := setup(t, filepath.Join(t.TempDir(), "race.db"))
	id := f.project(t)
	f.matched(t, id)
	f.post(t, "/api/problems/"+id+"/pool", "region-tomsk", "region", `{"amount":10}`, 200)
	var wg sync.WaitGroup
	var mu sync.Mutex
	accepted := 0
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := f.investments.Invest(context.Background(), models.User{ID: "investor-a", Role: "investor"}, id, 1, fmt.Sprint(i))
			if err == nil {
				mu.Lock()
				accepted++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	if accepted != 10 {
		t.Fatalf("accepted %d", accepted)
	}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := f.distributions.Distribute(context.Background(), models.User{ID: "region-tomsk", Role: "region"}, id, 100)
			if err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	events, _ := f.events.Pending(context.Background())
	n := 0
	for _, e := range events {
		if e.Name == "ProfitDistributed" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("distributed %d times", n)
	}
}

// TestRestartAndOutboxRecovery Verify saved payouts and undelivered events survive database reopening.
func TestRestartAndOutboxRecovery(t *testing.T) {
	path := filepath.Join(t.TempDir(), "persist.db")
	f := setup(t, path)
	id := f.project(t)
	f.matched(t, id)
	base := "/api/problems/" + id
	f.post(t, base+"/pool", "region-tomsk", "region", `{"amount":1}`, 200)
	f.post(t, base+"/investments", "investor-a", "investor", `{"amount":1,"key":"a"}`, 200)
	f.post(t, base+"/profit", "region-tomsk", "region", `{"amount":3}`, 200)
	f.repo.DB.Close()
	g := setup(t, path)
	g.post(t, base+"/profit", "region-tomsk", "region", `{"amount":3}`, 200)
	if err := g.bus.Drain(context.Background()); err != nil {
		t.Fatal(err)
	}
	events, _ := g.events.Events(context.Background())
	if len(events) != 6 {
		t.Fatalf("events lost: %d", len(events))
	}
	pending, _ := g.events.Pending(context.Background())
	if len(pending) != 0 {
		t.Fatal("outbox not drained")
	}
}

// TestRoleMatrix Reject every command issued by an unauthorized role.
func TestRoleMatrix(t *testing.T) {
	f := setup(t, ":memory:")
	id := f.project(t)
	actions := map[string]string{"": "region", "/solutions": "technologist", "/match": "expert", "/pool": "region", "/investments": "investor", "/profit": "region"}
	for _, u := range models.Users {
		for suffix, allowed := range actions {
			if u.Role == allowed {
				continue
			}
			path := "/api/problems/" + id + suffix
			if suffix == "" {
				path = "/api/problems"
			}
			f.post(t, path, u.ID, u.Role, `{}`, 403)
		}
	}
}

// TestAPIOnly Verify the service does not serve pages or frontend assets.
func TestAPIOnly(t *testing.T) {
	f := setup(t, ":memory:")
	for _, path := range []string{"/", "/index.html", "/src/main.jsx", "/assets/app.js"} {
		f.request(t, "GET", path, "region-tomsk", "region", nil, 404)
	}
	f.request(t, "GET", "/health", "", "", nil, 200)
	f.request(t, "GET", "/api/users", "", "", nil, 200)
}
