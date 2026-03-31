package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type authResponse struct {
	User struct {
		ID       string  `json:"id"`
		Phone    *string `json:"phone"`
		Nickname string  `json:"nickname"`
	} `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type petResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type deviceResponse struct {
	ID string `json:"id"`
}

type deviceStatusResponse struct {
	LatestDataPoints []json.RawMessage `json:"latest_data_points"`
}

type deviceStreamMessage struct {
	Type   string                `json:"type"`
	Status *deviceStatusResponse `json:"status"`
}

type healthSummaryResponse struct {
	Score int `json:"score"`
}

type healthAlertResponse struct {
	ID string `json:"id"`
}

type serviceProviderResponse struct {
	ID string `json:"id"`
}

type serviceAvailabilityResponse struct {
	Slots []string `json:"slots"`
}

type bookingResponse struct {
	ID string `json:"id"`
}

type postResponse struct {
	ID string `json:"id"`
}

type commentResponse struct {
	ID string `json:"id"`
}

type communityAIResponse struct {
	Answer string `json:"answer"`
}

type trainingPlanResponse struct {
	ID          string           `json:"id"`
	AIGenerated bool             `json:"ai_generated"`
	Progress    int              `json:"progress"`
	Steps       []map[string]any `json:"steps"`
}

type productResponse struct {
	ID                string `json:"id"`
	RecommendedReason string `json:"recommended_reason"`
}

type notificationResponse struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	IsRead bool   `json:"is_read"`
}

type notificationReadResponse struct {
	ID          string `json:"id"`
	Read        bool   `json:"read"`
	UnreadCount int64  `json:"unread_count"`
}

type notificationReadAllResponse struct {
	Updated     int64 `json:"updated"`
	UnreadCount int64 `json:"unread_count"`
}

type notificationStreamMessage struct {
	Type          string                 `json:"type"`
	Notification  *notificationResponse  `json:"notification"`
	Notifications []notificationResponse `json:"notifications"`
	UnreadCount   int64                  `json:"unread_count"`
}

type smokeRunner struct {
	baseURL string
	rootURL string
	wsRoot  string
	client  *http.Client
	skipWS  bool
}

func main() {
	baseURL := flag.String("base-url", "http://127.0.0.1:8080/api/v1", "API base URL")
	waitDuration := flag.Duration("wait", 0, "wait for API readiness before running")
	skipWS := flag.Bool("skip-ws", false, "skip websocket checks")
	flag.Parse()

	rootURL := strings.TrimSuffix(*baseURL, "/api/v1")
	runner := &smokeRunner{
		baseURL: strings.TrimRight(*baseURL, "/"),
		rootURL: rootURL,
		wsRoot:  strings.Replace(rootURL, "http://", "ws://", 1),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		skipWS: *skipWS,
	}

	ctx := context.Background()
	if *waitDuration > 0 {
		if err := runner.waitForReady(ctx, *waitDuration); err != nil {
			log.Fatalf("wait for ready: %v", err)
		}
	}

	if err := runner.run(ctx); err != nil {
		log.Fatalf("smoke failed: %v", err)
	}

	log.Printf("smoke passed against %s", runner.baseURL)
}

func (r *smokeRunner) run(ctx context.Context) error {
	phone := fmt.Sprintf("13%09d", time.Now().UnixNano()%1_000_000_000)

	log.Printf("register/login with phone %s", phone)
	if err := r.request(ctx, http.MethodPost, "/auth/register", map[string]any{
		"phone":    phone,
		"password": "strong-pass",
		"nickname": "SmokeTester",
	}, "", http.StatusCreated, nil); err != nil {
		return fmt.Errorf("register: %w", err)
	}

	var login authResponse
	if err := r.request(ctx, http.MethodPost, "/auth/login", map[string]any{
		"phone":    phone,
		"password": "strong-pass",
	}, "", http.StatusOK, &login); err != nil {
		return fmt.Errorf("login: %w", err)
	}
	if login.AccessToken == "" || login.RefreshToken == "" {
		return fmt.Errorf("login returned empty tokens")
	}

	var refreshed authResponse
	if err := r.request(ctx, http.MethodPost, "/auth/refresh", map[string]any{
		"refresh_token": login.RefreshToken,
	}, "", http.StatusOK, &refreshed); err != nil {
		return fmt.Errorf("refresh: %w", err)
	}

	var me struct {
		ID    string  `json:"id"`
		Phone *string `json:"phone"`
	}
	if err := r.request(ctx, http.MethodGet, "/users/me", nil, login.AccessToken, http.StatusOK, &me); err != nil {
		return fmt.Errorf("get me: %w", err)
	}
	if me.Phone == nil || *me.Phone != phone {
		return fmt.Errorf("unexpected /users/me phone: %+v", me.Phone)
	}

	var pet petResponse
	if err := r.request(ctx, http.MethodPost, "/pets", map[string]any{
		"name":        "DouDou",
		"species":     "dog",
		"breed":       "corgi",
		"gender":      "female",
		"is_neutered": true,
		"allergies":   []string{"beef"},
		"notes":       "smoke test pet",
	}, login.AccessToken, http.StatusCreated, &pet); err != nil {
		return fmt.Errorf("create pet: %w", err)
	}
	if pet.ID == "" {
		return fmt.Errorf("create pet returned empty id")
	}

	var pets []petResponse
	if err := r.request(ctx, http.MethodGet, "/pets", nil, login.AccessToken, http.StatusOK, &pets); err != nil {
		return fmt.Errorf("list pets: %w", err)
	}
	if len(pets) == 0 {
		return fmt.Errorf("list pets returned no items")
	}

	serialNumber := fmt.Sprintf("SMOKE-%d", time.Now().UnixNano())
	var device deviceResponse
	if err := r.request(ctx, http.MethodPost, "/devices", map[string]any{
		"pet_id":        pet.ID,
		"device_type":   "feeder",
		"brand":         "PetVerse",
		"model":         "F-SMOKE",
		"nickname":      "Smoke Feeder",
		"serial_number": serialNumber,
	}, login.AccessToken, http.StatusCreated, &device); err != nil {
		return fmt.Errorf("create device: %w", err)
	}

	var deviceConn *websocket.Conn
	if !r.skipWS {
		conn, err := r.connectWS("/devices/"+device.ID+"/stream", login.AccessToken)
		if err != nil {
			return fmt.Errorf("connect device websocket: %w", err)
		}
		deviceConn = conn
		defer deviceConn.Close()

		var initialDeviceMessage deviceStreamMessage
		if err := readWSJSON(deviceConn, &initialDeviceMessage); err != nil {
			return fmt.Errorf("read initial device websocket message: %w", err)
		}
		if initialDeviceMessage.Type != "device_status" || initialDeviceMessage.Status == nil {
			return fmt.Errorf("unexpected device websocket message: %+v", initialDeviceMessage)
		}
	}

	var deviceStatus deviceStatusResponse
	if err := r.request(ctx, http.MethodGet, "/devices/"+device.ID+"/status", nil, login.AccessToken, http.StatusOK, &deviceStatus); err != nil {
		return fmt.Errorf("device status: %w", err)
	}
	if len(deviceStatus.LatestDataPoints) == 0 {
		return fmt.Errorf("device status returned no datapoints")
	}

	recordedAt := time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339)
	dueDate := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	if err := r.request(ctx, http.MethodPost, "/pets/"+pet.ID+"/health", map[string]any{
		"type":        "symptom",
		"title":       "夜间异常",
		"description": "出现呕吐和腹泻，需要继续观察",
		"recorded_at": recordedAt,
		"due_date":    dueDate,
		"data": map[string]any{
			"temperature": 39.2,
		},
	}, login.AccessToken, http.StatusCreated, nil); err != nil {
		return fmt.Errorf("create health record: %w", err)
	}

	var alerts []healthAlertResponse
	if err := r.request(ctx, http.MethodGet, "/pets/"+pet.ID+"/health/alerts", nil, login.AccessToken, http.StatusOK, &alerts); err != nil {
		return fmt.Errorf("health alerts: %w", err)
	}
	if len(alerts) == 0 {
		return fmt.Errorf("health alerts returned no items")
	}

	var healthSummary healthSummaryResponse
	if err := r.request(ctx, http.MethodGet, "/pets/"+pet.ID+"/health/summary", nil, login.AccessToken, http.StatusOK, &healthSummary); err != nil {
		return fmt.Errorf("health summary: %w", err)
	}
	if healthSummary.Score <= 0 {
		return fmt.Errorf("invalid health summary score: %d", healthSummary.Score)
	}

	if err := r.request(ctx, http.MethodPost, "/devices/"+device.ID+"/command", map[string]any{
		"command": "feed_now",
		"params": map[string]any{
			"value": 42,
		},
	}, login.AccessToken, http.StatusOK, nil); err != nil {
		return fmt.Errorf("device command: %w", err)
	}

	if deviceConn != nil {
		var updatedDeviceMessage deviceStreamMessage
		if err := readWSJSON(deviceConn, &updatedDeviceMessage); err != nil {
			return fmt.Errorf("read updated device websocket message: %w", err)
		}
		if updatedDeviceMessage.Status == nil || len(updatedDeviceMessage.Status.LatestDataPoints) == 0 {
			return fmt.Errorf("unexpected updated device websocket message: %+v", updatedDeviceMessage)
		}
	}

	var providers []serviceProviderResponse
	if err := r.request(ctx, http.MethodGet, "/services?lat=31.23&lng=121.47", nil, login.AccessToken, http.StatusOK, &providers); err != nil {
		return fmt.Errorf("list services: %w", err)
	}
	if len(providers) == 0 {
		return fmt.Errorf("list services returned no items")
	}

	var availability serviceAvailabilityResponse
	if err := r.request(ctx, http.MethodGet, "/services/"+providers[0].ID+"/availability", nil, login.AccessToken, http.StatusOK, &availability); err != nil {
		return fmt.Errorf("service availability: %w", err)
	}
	if len(availability.Slots) == 0 {
		return fmt.Errorf("service availability returned no slots")
	}

	var booking bookingResponse
	if err := r.request(ctx, http.MethodPost, "/bookings", map[string]any{
		"provider_id":  providers[0].ID,
		"pet_id":       pet.ID,
		"service_name": "年度体检",
		"start_time":   availability.Slots[0],
		"price":        199,
	}, login.AccessToken, http.StatusCreated, &booking); err != nil {
		return fmt.Errorf("create booking: %w", err)
	}

	if err := r.request(ctx, http.MethodPut, "/bookings/"+booking.ID+"/review", map[string]any{
		"rating": 5,
		"review": "smoke review",
	}, login.AccessToken, http.StatusOK, nil); err != nil {
		return fmt.Errorf("review booking: %w", err)
	}

	var post postResponse
	if err := r.request(ctx, http.MethodPost, "/posts", map[string]any{
		"pet_id":  pet.ID,
		"title":   "Smoke 帖子",
		"content": "社区流程验证",
		"tags":    []string{"smoke", "test"},
	}, login.AccessToken, http.StatusCreated, &post); err != nil {
		return fmt.Errorf("create post: %w", err)
	}

	if err := r.request(ctx, http.MethodPost, "/posts/"+post.ID+"/like", nil, login.AccessToken, http.StatusOK, nil); err != nil {
		return fmt.Errorf("like post: %w", err)
	}

	var comment commentResponse
	if err := r.request(ctx, http.MethodPost, "/posts/"+post.ID+"/comments", map[string]any{
		"content": "smoke comment",
	}, login.AccessToken, http.StatusCreated, &comment); err != nil {
		return fmt.Errorf("create comment: %w", err)
	}
	if comment.ID == "" {
		return fmt.Errorf("create comment returned empty id")
	}

	var communityAI communityAIResponse
	if err := r.request(ctx, http.MethodPost, "/community/ask-ai", map[string]any{
		"question": "给狗狗选寄养机构最该确认什么？",
	}, login.AccessToken, http.StatusOK, &communityAI); err != nil {
		return fmt.Errorf("community ask ai: %w", err)
	}
	if communityAI.Answer == "" {
		return fmt.Errorf("community ask ai returned empty answer")
	}

	var notificationConn *websocket.Conn
	if !r.skipWS {
		conn, err := r.connectWS("/ws", login.AccessToken)
		if err != nil {
			return fmt.Errorf("connect notifications websocket: %w", err)
		}
		notificationConn = conn
		defer notificationConn.Close()

		var syncMessage notificationStreamMessage
		if err := readWSJSON(notificationConn, &syncMessage); err != nil {
			return fmt.Errorf("read initial notification websocket message: %w", err)
		}
		if syncMessage.Type != "notification_sync" {
			return fmt.Errorf("unexpected notification sync message: %+v", syncMessage)
		}
	}

	var generatedPlan trainingPlanResponse
	if err := r.request(ctx, http.MethodPost, "/pets/"+pet.ID+"/training/generate", map[string]any{
		"goal":       "学会坐下等待",
		"difficulty": "beginner",
		"category":   "obedience",
	}, login.AccessToken, http.StatusCreated, &generatedPlan); err != nil {
		return fmt.Errorf("generate training plan: %w", err)
	}
	if !generatedPlan.AIGenerated || len(generatedPlan.Steps) == 0 {
		return fmt.Errorf("generated training plan is incomplete: %+v", generatedPlan)
	}

	if notificationConn != nil {
		var notificationMessage notificationStreamMessage
		if err := readWSJSON(notificationConn, &notificationMessage); err != nil {
			return fmt.Errorf("read generated training notification: %w", err)
		}
		if notificationMessage.Type != "notification" || notificationMessage.Notification == nil {
			return fmt.Errorf("unexpected training notification message: %+v", notificationMessage)
		}
	}

	var plans []trainingPlanResponse
	if err := r.request(ctx, http.MethodGet, "/pets/"+pet.ID+"/training", nil, login.AccessToken, http.StatusOK, &plans); err != nil {
		return fmt.Errorf("list training plans: %w", err)
	}
	if len(plans) == 0 {
		return fmt.Errorf("list training plans returned no items")
	}

	if err := r.request(ctx, http.MethodPut, "/training/"+generatedPlan.ID, map[string]any{
		"progress": 50,
	}, login.AccessToken, http.StatusOK, nil); err != nil {
		return fmt.Errorf("update training plan: %w", err)
	}

	var planDetail trainingPlanResponse
	if err := r.request(ctx, http.MethodGet, "/training/"+generatedPlan.ID, nil, login.AccessToken, http.StatusOK, &planDetail); err != nil {
		return fmt.Errorf("get training plan: %w", err)
	}
	if planDetail.Progress != 50 {
		return fmt.Errorf("unexpected training progress: %d", planDetail.Progress)
	}

	var products []productResponse
	if err := r.request(ctx, http.MethodGet, "/shop/products?category=food", nil, login.AccessToken, http.StatusOK, &products); err != nil {
		return fmt.Errorf("list products: %w", err)
	}
	if len(products) == 0 {
		return fmt.Errorf("list products returned no items")
	}

	if err := r.request(ctx, http.MethodGet, "/shop/products/"+products[0].ID, nil, login.AccessToken, http.StatusOK, nil); err != nil {
		return fmt.Errorf("get product detail: %w", err)
	}

	var recommendations []productResponse
	if err := r.request(ctx, http.MethodGet, "/shop/recommendations/"+pet.ID, nil, login.AccessToken, http.StatusOK, &recommendations); err != nil {
		return fmt.Errorf("get product recommendations: %w", err)
	}
	if len(recommendations) == 0 || recommendations[0].RecommendedReason == "" {
		return fmt.Errorf("recommendations missing reason")
	}

	var notifications []notificationResponse
	if err := r.request(ctx, http.MethodGet, "/notifications?page=1&page_size=20", nil, login.AccessToken, http.StatusOK, &notifications); err != nil {
		return fmt.Errorf("list notifications: %w", err)
	}
	if len(notifications) == 0 {
		return fmt.Errorf("list notifications returned no items")
	}

	var markRead notificationReadResponse
	if err := r.request(ctx, http.MethodPut, "/notifications/"+notifications[0].ID+"/read", nil, login.AccessToken, http.StatusOK, &markRead); err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	if !markRead.Read {
		return fmt.Errorf("notification was not marked read")
	}

	var markAll notificationReadAllResponse
	if err := r.request(ctx, http.MethodPut, "/notifications/read-all", nil, login.AccessToken, http.StatusOK, &markAll); err != nil {
		return fmt.Errorf("mark all notifications read: %w", err)
	}

	log.Printf("smoke flow completed successfully")
	return nil
}

func (r *smokeRunner) waitForReady(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.rootURL+"/healthz", nil)
		if err != nil {
			return err
		}
		resp, err := r.client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			_ = resp.Body.Close()
			return nil
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		lastErr = err
		time.Sleep(1 * time.Second)
	}

	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("healthz did not become ready within %s", timeout)
}

func (r *smokeRunner) request(ctx context.Context, method, path string, payload any, token string, expectedStatus int, out any) error {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, r.baseURL+path, body)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != expectedStatus {
		return fmt.Errorf("%s %s returned %d want %d body=%s", method, path, resp.StatusCode, expectedStatus, string(rawBody))
	}

	if out == nil {
		return nil
	}

	var env envelope
	if err := json.Unmarshal(rawBody, &env); err != nil {
		return fmt.Errorf("decode envelope: %w body=%s", err, string(rawBody))
	}
	if len(env.Data) == 0 || string(env.Data) == "null" {
		return nil
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("decode envelope data: %w body=%s", err, string(rawBody))
	}
	return nil
}

func (r *smokeRunner) connectWS(path, accessToken string) (*websocket.Conn, error) {
	wsURL, err := url.Parse(r.wsRoot + "/api/v1" + path)
	if err != nil {
		return nil, err
	}
	query := wsURL.Query()
	query.Set("access_token", accessToken)
	wsURL.RawQuery = query.Encode()

	conn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func readWSJSON[T any](conn *websocket.Conn, target *T) error {
	_ = conn.SetReadDeadline(time.Now().Add(4 * time.Second))
	return conn.ReadJSON(target)
}
