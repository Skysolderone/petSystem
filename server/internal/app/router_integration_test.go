package app

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"log/slog"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"petverse/server/internal/dto"
	"petverse/server/internal/handler"
	"petverse/server/internal/pkg/ai"
	petjwt "petverse/server/internal/pkg/jwt"
	"petverse/server/internal/pkg/upload"
	"petverse/server/internal/repository"
	"petverse/server/internal/service"
	"petverse/server/internal/ws"
)

func TestPhaseOneAuthAndPetFlow(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t)

	loginBody := registerAndLogin(t, router, "13800138000")

	meResponse := performJSONRequest(t, router, http.MethodGet, "/api/v1/users/me", nil, loginBody.AccessToken)
	if meResponse.Code != http.StatusOK {
		t.Fatalf("me status = %d want %d body=%s", meResponse.Code, http.StatusOK, meResponse.Body.String())
	}

	var meBody dto.UserResponse
	decodeEnvelopeData(t, meResponse.Body.Bytes(), &meBody)
	if meBody.Phone == nil || *meBody.Phone != "13800138000" {
		t.Fatalf("unexpected me body: %+v", meBody)
	}

	createPetResponse := performJSONRequest(t, router, http.MethodPost, "/api/v1/pets", map[string]any{
		"name":        "DouDou",
		"species":     "dog",
		"breed":       "corgi",
		"gender":      "female",
		"is_neutered": true,
		"allergies":   []string{"beef"},
		"notes":       "friendly",
	}, loginBody.AccessToken)
	if createPetResponse.Code != http.StatusCreated {
		t.Fatalf("create pet status = %d want %d body=%s", createPetResponse.Code, http.StatusCreated, createPetResponse.Body.String())
	}

	var petBody dto.PetResponse
	decodeEnvelopeData(t, createPetResponse.Body.Bytes(), &petBody)
	if petBody.ID == "" {
		t.Fatalf("create pet returned empty id: %+v", petBody)
	}
	if _, err := uuid.Parse(petBody.ID); err != nil {
		t.Fatalf("create pet returned invalid uuid: %s", petBody.ID)
	}

	listPetsResponse := performJSONRequest(t, router, http.MethodGet, "/api/v1/pets", nil, loginBody.AccessToken)
	if listPetsResponse.Code != http.StatusOK {
		t.Fatalf("list pets status = %d want %d body=%s", listPetsResponse.Code, http.StatusOK, listPetsResponse.Body.String())
	}

	var listBody []dto.PetResponse
	decodeEnvelopeData(t, listPetsResponse.Body.Bytes(), &listBody)
	if len(listBody) != 1 {
		t.Fatalf("list pets returned %d items want 1: %+v", len(listBody), listBody)
	}
	if listBody[0].Name != "DouDou" {
		t.Fatalf("unexpected pet in list: %+v", listBody[0])
	}
}

func TestSocialLoginFlow(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t)

	var wechatLogin dto.AuthResponse
	wechatResponse := performJSONRequest(t, router, http.MethodPost, "/api/v1/auth/login/wechat", map[string]any{
		"open_id":    "wx-smoke-user",
		"nickname":   "微信测试用户",
		"avatar_url": "https://example.com/wechat.png",
	}, "")
	if wechatResponse.Code != http.StatusOK {
		t.Fatalf("wechat login status = %d want %d body=%s", wechatResponse.Code, http.StatusOK, wechatResponse.Body.String())
	}
	decodeEnvelopeData(t, wechatResponse.Body.Bytes(), &wechatLogin)
	if wechatLogin.User.ID == "" || wechatLogin.User.Nickname != "微信测试用户" {
		t.Fatalf("unexpected wechat login payload: %+v", wechatLogin)
	}

	meResponse := performJSONRequest(t, router, http.MethodGet, "/api/v1/users/me", nil, wechatLogin.AccessToken)
	if meResponse.Code != http.StatusOK {
		t.Fatalf("wechat /users/me status = %d want %d body=%s", meResponse.Code, http.StatusOK, meResponse.Body.String())
	}

	var appleLogin dto.AuthResponse
	appleResponse := performJSONRequest(t, router, http.MethodPost, "/api/v1/auth/login/apple", map[string]any{
		"apple_id":   "apple-smoke-user",
		"email":      "apple@example.com",
		"nickname":   "Apple 测试用户",
		"avatar_url": "https://example.com/apple.png",
	}, "")
	if appleResponse.Code != http.StatusOK {
		t.Fatalf("apple login status = %d want %d body=%s", appleResponse.Code, http.StatusOK, appleResponse.Body.String())
	}
	decodeEnvelopeData(t, appleResponse.Body.Bytes(), &appleLogin)
	if appleLogin.User.Email == nil || *appleLogin.User.Email != "apple@example.com" {
		t.Fatalf("unexpected apple login payload: %+v", appleLogin)
	}

	var googleLogin dto.AuthResponse
	googleResponse := performJSONRequest(t, router, http.MethodPost, "/api/v1/auth/login/google", map[string]any{
		"google_id":  "google-smoke-user",
		"email":      "google@example.com",
		"nickname":   "Google 测试用户",
		"avatar_url": "https://example.com/google.png",
	}, "")
	if googleResponse.Code != http.StatusOK {
		t.Fatalf("google login status = %d want %d body=%s", googleResponse.Code, http.StatusOK, googleResponse.Body.String())
	}
	decodeEnvelopeData(t, googleResponse.Body.Bytes(), &googleLogin)
	if googleLogin.User.Email == nil || *googleLogin.User.Email != "google@example.com" {
		t.Fatalf("unexpected google login payload: %+v", googleLogin)
	}
}

func TestPhaseTwoHealthAndDeviceFlow(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t)
	loginBody := registerAndLogin(t, router, "13800138010")
	petBody := createPet(t, router, loginBody.AccessToken)

	createDeviceResponse := performJSONRequest(t, router, http.MethodPost, "/api/v1/devices", map[string]any{
		"pet_id":        petBody.ID,
		"device_type":   "feeder",
		"brand":         "PetVerse",
		"model":         "F1",
		"nickname":      "客厅喂食器",
		"serial_number": "SER-1001",
	}, loginBody.AccessToken)
	if createDeviceResponse.Code != http.StatusCreated {
		t.Fatalf("create device status = %d want %d body=%s", createDeviceResponse.Code, http.StatusCreated, createDeviceResponse.Body.String())
	}

	var deviceBody dto.DeviceResponse
	decodeEnvelopeData(t, createDeviceResponse.Body.Bytes(), &deviceBody)

	statusResponse := performJSONRequest(t, router, http.MethodGet, "/api/v1/devices/"+deviceBody.ID+"/status", nil, loginBody.AccessToken)
	if statusResponse.Code != http.StatusOK {
		t.Fatalf("device status status = %d want %d body=%s", statusResponse.Code, http.StatusOK, statusResponse.Body.String())
	}

	var statusBody dto.DeviceStatusResponse
	decodeEnvelopeData(t, statusResponse.Body.Bytes(), &statusBody)
	if len(statusBody.LatestDataPoints) == 0 {
		t.Fatal("device status should include seeded data points")
	}

	recordedAt := time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339)
	dueDate := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	createHealthResponse := performJSONRequest(t, router, http.MethodPost, "/api/v1/pets/"+petBody.ID+"/health", map[string]any{
		"type":        "symptom",
		"title":       "夜间异常",
		"description": "出现呕吐和腹泻，需要继续观察",
		"recorded_at": recordedAt,
		"due_date":    dueDate,
		"data": map[string]any{
			"temperature": 39.2,
		},
	}, loginBody.AccessToken)
	if createHealthResponse.Code != http.StatusCreated {
		t.Fatalf("create health status = %d want %d body=%s", createHealthResponse.Code, http.StatusCreated, createHealthResponse.Body.String())
	}

	alertsResponse := performJSONRequest(t, router, http.MethodGet, "/api/v1/pets/"+petBody.ID+"/health/alerts", nil, loginBody.AccessToken)
	if alertsResponse.Code != http.StatusOK {
		t.Fatalf("alerts status = %d want %d body=%s", alertsResponse.Code, http.StatusOK, alertsResponse.Body.String())
	}
	var alertsBody []dto.HealthAlertResponse
	decodeEnvelopeData(t, alertsResponse.Body.Bytes(), &alertsBody)
	if len(alertsBody) == 0 {
		t.Fatal("health alerts should contain an auto-generated alert")
	}

	summaryResponse := performJSONRequest(t, router, http.MethodGet, "/api/v1/pets/"+petBody.ID+"/health/summary", nil, loginBody.AccessToken)
	if summaryResponse.Code != http.StatusOK {
		t.Fatalf("summary status = %d want %d body=%s", summaryResponse.Code, http.StatusOK, summaryResponse.Body.String())
	}
	var summaryBody dto.HealthSummaryResponse
	decodeEnvelopeData(t, summaryResponse.Body.Bytes(), &summaryBody)
	if summaryBody.Score <= 0 || summaryBody.GeneratedAt.IsZero() {
		t.Fatalf("unexpected summary payload: %+v", summaryBody)
	}

	askAIResponse := performJSONRequest(t, router, http.MethodPost, "/api/v1/pets/"+petBody.ID+"/health/ask-ai", map[string]any{
		"question": "这种情况要不要尽快去医院？",
	}, loginBody.AccessToken)
	if askAIResponse.Code != http.StatusOK {
		t.Fatalf("ask ai status = %d want %d body=%s", askAIResponse.Code, http.StatusOK, askAIResponse.Body.String())
	}
	var askAIBody dto.HealthAIAnswerResponse
	decodeEnvelopeData(t, askAIResponse.Body.Bytes(), &askAIBody)
	if askAIBody.Answer == "" {
		t.Fatalf("empty ai answer: %+v", askAIBody)
	}

	commandResponse := performJSONRequest(t, router, http.MethodPost, "/api/v1/devices/"+deviceBody.ID+"/command", map[string]any{
		"command": "feed_now",
		"params": map[string]any{
			"value": 50,
		},
	}, loginBody.AccessToken)
	if commandResponse.Code != http.StatusOK {
		t.Fatalf("device command status = %d want %d body=%s", commandResponse.Code, http.StatusOK, commandResponse.Body.String())
	}

	dataResponse := performJSONRequest(t, router, http.MethodGet, "/api/v1/devices/"+deviceBody.ID+"/data?metric=feeding_amount&hours=24&limit=10", nil, loginBody.AccessToken)
	if dataResponse.Code != http.StatusOK {
		t.Fatalf("device data status = %d want %d body=%s", dataResponse.Code, http.StatusOK, dataResponse.Body.String())
	}
	var dataBody []dto.DeviceDataPointResponse
	decodeEnvelopeData(t, dataResponse.Body.Bytes(), &dataBody)
	if len(dataBody) == 0 {
		t.Fatal("device data query returned no datapoints")
	}
}

func TestUserProfileAndUploadFlow(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t)
	loginBody := registerAndLogin(t, router, "13800138005")
	petBody := createPet(t, router, loginBody.AccessToken)

	updateMeResponse := performJSONRequest(t, router, http.MethodPut, "/api/v1/users/me", map[string]any{
		"nickname": "Updated Tester",
		"email":    "tester@example.com",
	}, loginBody.AccessToken)
	if updateMeResponse.Code != http.StatusOK {
		t.Fatalf("update me status = %d want %d body=%s", updateMeResponse.Code, http.StatusOK, updateMeResponse.Body.String())
	}

	var updatedUser dto.UserResponse
	decodeEnvelopeData(t, updateMeResponse.Body.Bytes(), &updatedUser)
	if updatedUser.Nickname != "Updated Tester" || updatedUser.Email == nil || *updatedUser.Email != "tester@example.com" {
		t.Fatalf("unexpected updated user: %+v", updatedUser)
	}

	updateLocationResponse := performJSONRequest(t, router, http.MethodPut, "/api/v1/users/me/location", map[string]any{
		"latitude":  31.2304,
		"longitude": 121.4737,
	}, loginBody.AccessToken)
	if updateLocationResponse.Code != http.StatusOK {
		t.Fatalf("update location status = %d want %d body=%s", updateLocationResponse.Code, http.StatusOK, updateLocationResponse.Body.String())
	}

	var locatedUser dto.UserResponse
	decodeEnvelopeData(t, updateLocationResponse.Body.Bytes(), &locatedUser)
	if locatedUser.Latitude == nil || locatedUser.Longitude == nil {
		t.Fatalf("expected location to be set: %+v", locatedUser)
	}

	imagePayload := mustPNG(t)

	uploadImageResponse := performMultipartRequest(t, router, http.MethodPost, "/api/v1/upload/image", "file", "avatar.png", imagePayload, "image/png", loginBody.AccessToken)
	if uploadImageResponse.Code != http.StatusCreated {
		t.Fatalf("upload image status = %d want %d body=%s", uploadImageResponse.Code, http.StatusCreated, uploadImageResponse.Body.String())
	}
	var uploadedImage dto.UploadResponse
	decodeEnvelopeData(t, uploadImageResponse.Body.Bytes(), &uploadedImage)
	if uploadedImage.URL == "" || uploadedImage.MIMEType == "" {
		t.Fatalf("unexpected upload image response: %+v", uploadedImage)
	}

	uploadFileResponse := performMultipartRequest(t, router, http.MethodPost, "/api/v1/upload/file", "file", "notes.txt", []byte("smoke file"), "text/plain", loginBody.AccessToken)
	if uploadFileResponse.Code != http.StatusCreated {
		t.Fatalf("upload file status = %d want %d body=%s", uploadFileResponse.Code, http.StatusCreated, uploadFileResponse.Body.String())
	}

	updateAvatarResponse := performMultipartRequest(t, router, http.MethodPut, "/api/v1/users/me/avatar", "file", "avatar.png", imagePayload, "image/png", loginBody.AccessToken)
	if updateAvatarResponse.Code != http.StatusOK {
		t.Fatalf("user avatar upload status = %d want %d body=%s", updateAvatarResponse.Code, http.StatusOK, updateAvatarResponse.Body.String())
	}
	var avatarUser dto.UserResponse
	decodeEnvelopeData(t, updateAvatarResponse.Body.Bytes(), &avatarUser)
	if avatarUser.AvatarURL == "" {
		t.Fatalf("expected avatar url to be set: %+v", avatarUser)
	}

	updatePetAvatarResponse := performMultipartRequest(t, router, http.MethodPut, "/api/v1/pets/"+petBody.ID+"/avatar", "file", "pet.png", imagePayload, "image/png", loginBody.AccessToken)
	if updatePetAvatarResponse.Code != http.StatusOK {
		t.Fatalf("pet avatar upload status = %d want %d body=%s", updatePetAvatarResponse.Code, http.StatusOK, updatePetAvatarResponse.Body.String())
	}
	var avatarPet dto.PetResponse
	decodeEnvelopeData(t, updatePetAvatarResponse.Body.Bytes(), &avatarPet)
	if avatarPet.AvatarURL == "" {
		t.Fatalf("expected pet avatar url to be set: %+v", avatarPet)
	}
}

func TestPushTokenRegistrationFlow(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t)
	loginBody := registerAndLogin(t, router, "13800138006")

	registerResponse := performJSONRequest(t, router, http.MethodPost, "/api/v1/notifications/push-token", map[string]any{
		"token":    "ExponentPushToken[test-token]",
		"provider": "expo",
		"platform": "ios",
	}, loginBody.AccessToken)
	if registerResponse.Code != http.StatusOK {
		t.Fatalf("register push token status = %d want %d body=%s", registerResponse.Code, http.StatusOK, registerResponse.Body.String())
	}

	var pushTokenBody dto.PushTokenResponse
	decodeEnvelopeData(t, registerResponse.Body.Bytes(), &pushTokenBody)
	if pushTokenBody.Token != "ExponentPushToken[test-token]" || !pushTokenBody.IsActive {
		t.Fatalf("unexpected push token payload: %+v", pushTokenBody)
	}

	deleteResponse := performJSONRequest(t, router, http.MethodDelete, "/api/v1/notifications/push-token", map[string]any{
		"token": "ExponentPushToken[test-token]",
	}, loginBody.AccessToken)
	if deleteResponse.Code != http.StatusOK {
		t.Fatalf("unregister push token status = %d want %d body=%s", deleteResponse.Code, http.StatusOK, deleteResponse.Body.String())
	}
}

func TestDeviceWebSocketStreamBroadcastsUpdates(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t)
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Skipf("skipping websocket stream test: cannot open local listener in this environment: %v", err)
	}

	server := httptest.NewUnstartedServer(router)
	server.Listener = listener
	server.Start()
	defer server.Close()

	loginBody := registerAndLogin(t, router, "13800138020")
	petBody := createPet(t, router, loginBody.AccessToken)

	createDeviceResponse := performJSONRequest(t, router, http.MethodPost, "/api/v1/devices", map[string]any{
		"pet_id":        petBody.ID,
		"device_type":   "feeder",
		"brand":         "PetVerse",
		"model":         "F2",
		"nickname":      "书房喂食器",
		"serial_number": "SER-2001",
	}, loginBody.AccessToken)
	if createDeviceResponse.Code != http.StatusCreated {
		t.Fatalf("create device status = %d want %d body=%s", createDeviceResponse.Code, http.StatusCreated, createDeviceResponse.Body.String())
	}

	var deviceBody dto.DeviceResponse
	decodeEnvelopeData(t, createDeviceResponse.Body.Bytes(), &deviceBody)

	dialer := websocket.Dialer{}
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+loginBody.AccessToken)
	wsURL := "ws" + server.URL[len("http"):] + "/api/v1/devices/" + deviceBody.ID + "/stream"

	conn, _, err := dialer.Dial(wsURL, headers)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var initialMessage dto.DeviceStreamMessage
	if err := conn.ReadJSON(&initialMessage); err != nil {
		t.Fatalf("read initial websocket message: %v", err)
	}
	if initialMessage.Type != "device_status" || initialMessage.Status == nil {
		t.Fatalf("unexpected initial message: %+v", initialMessage)
	}

	commandResponse := performJSONRequest(t, router, http.MethodPost, "/api/v1/devices/"+deviceBody.ID+"/command", map[string]any{
		"command": "feed_now",
	}, loginBody.AccessToken)
	if commandResponse.Code != http.StatusOK {
		t.Fatalf("device command status = %d want %d body=%s", commandResponse.Code, http.StatusOK, commandResponse.Body.String())
	}

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var updateMessage dto.DeviceStreamMessage
	if err := conn.ReadJSON(&updateMessage); err != nil {
		t.Fatalf("read websocket update: %v", err)
	}
	if updateMessage.Status == nil || len(updateMessage.Status.LatestDataPoints) == 0 {
		t.Fatalf("unexpected update message: %+v", updateMessage)
	}
}

func TestPhaseThreeServicesBookingAndCommunityFlow(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t)
	loginBody := registerAndLogin(t, router, "13800138030")
	petBody := createPet(t, router, loginBody.AccessToken)

	servicesResponse := performJSONRequest(t, router, http.MethodGet, "/api/v1/services?lat=31.23&lng=121.47", nil, loginBody.AccessToken)
	if servicesResponse.Code != http.StatusOK {
		t.Fatalf("services status = %d want %d body=%s", servicesResponse.Code, http.StatusOK, servicesResponse.Body.String())
	}
	var providers []dto.ServiceProviderResponse
	decodeEnvelopeData(t, servicesResponse.Body.Bytes(), &providers)
	if len(providers) == 0 {
		t.Fatal("services list should return seeded providers")
	}
	providerID := providers[0].ID

	availabilityResponse := performJSONRequest(t, router, http.MethodGet, "/api/v1/services/"+providerID+"/availability", nil, loginBody.AccessToken)
	if availabilityResponse.Code != http.StatusOK {
		t.Fatalf("availability status = %d want %d body=%s", availabilityResponse.Code, http.StatusOK, availabilityResponse.Body.String())
	}
	var availability dto.ServiceAvailabilityResponse
	decodeEnvelopeData(t, availabilityResponse.Body.Bytes(), &availability)
	if len(availability.Slots) == 0 {
		t.Fatal("availability should return slots")
	}

	createBookingResponse := performJSONRequest(t, router, http.MethodPost, "/api/v1/bookings", map[string]any{
		"provider_id":  providerID,
		"pet_id":       petBody.ID,
		"service_name": "年度体检",
		"start_time":   availability.Slots[0],
		"price":        199,
	}, loginBody.AccessToken)
	if createBookingResponse.Code != http.StatusCreated {
		t.Fatalf("create booking status = %d want %d body=%s", createBookingResponse.Code, http.StatusCreated, createBookingResponse.Body.String())
	}
	var booking dto.BookingResponse
	decodeEnvelopeData(t, createBookingResponse.Body.Bytes(), &booking)

	reviewBookingResponse := performJSONRequest(t, router, http.MethodPut, "/api/v1/bookings/"+booking.ID+"/review", map[string]any{
		"rating": 5,
		"review": "医生解释清晰，流程顺畅。",
	}, loginBody.AccessToken)
	if reviewBookingResponse.Code != http.StatusOK {
		t.Fatalf("review booking status = %d want %d body=%s", reviewBookingResponse.Code, http.StatusOK, reviewBookingResponse.Body.String())
	}

	reviewsResponse := performJSONRequest(t, router, http.MethodGet, "/api/v1/services/"+providerID+"/reviews", nil, loginBody.AccessToken)
	if reviewsResponse.Code != http.StatusOK {
		t.Fatalf("service reviews status = %d want %d body=%s", reviewsResponse.Code, http.StatusOK, reviewsResponse.Body.String())
	}
	var reviews []dto.BookingResponse
	decodeEnvelopeData(t, reviewsResponse.Body.Bytes(), &reviews)
	if len(reviews) == 0 {
		t.Fatal("service reviews should include the reviewed booking")
	}

	createPostResponse := performJSONRequest(t, router, http.MethodPost, "/api/v1/posts", map[string]any{
		"pet_id":  petBody.ID,
		"title":   "第一次去体检",
		"content": "今天带豆豆做了年度体检，医生建议继续观察饮水量。",
		"tags":    []string{"体检", "医院"},
	}, loginBody.AccessToken)
	if createPostResponse.Code != http.StatusCreated {
		t.Fatalf("create post status = %d want %d body=%s", createPostResponse.Code, http.StatusCreated, createPostResponse.Body.String())
	}
	var post dto.PostResponse
	decodeEnvelopeData(t, createPostResponse.Body.Bytes(), &post)

	likeResponse := performJSONRequest(t, router, http.MethodPost, "/api/v1/posts/"+post.ID+"/like", nil, loginBody.AccessToken)
	if likeResponse.Code != http.StatusOK {
		t.Fatalf("like post status = %d want %d body=%s", likeResponse.Code, http.StatusOK, likeResponse.Body.String())
	}

	createCommentResponse := performJSONRequest(t, router, http.MethodPost, "/api/v1/posts/"+post.ID+"/comments", map[string]any{
		"content": "谢谢分享，这家医院周末也开门吗？",
	}, loginBody.AccessToken)
	if createCommentResponse.Code != http.StatusCreated {
		t.Fatalf("create comment status = %d want %d body=%s", createCommentResponse.Code, http.StatusCreated, createCommentResponse.Body.String())
	}

	commentsResponse := performJSONRequest(t, router, http.MethodGet, "/api/v1/posts/"+post.ID+"/comments", nil, loginBody.AccessToken)
	if commentsResponse.Code != http.StatusOK {
		t.Fatalf("list comments status = %d want %d body=%s", commentsResponse.Code, http.StatusOK, commentsResponse.Body.String())
	}
	var comments []dto.CommentResponse
	decodeEnvelopeData(t, commentsResponse.Body.Bytes(), &comments)
	if len(comments) == 0 {
		t.Fatal("comments list should contain the newly created comment")
	}

	communityAIResponse := performJSONRequest(t, router, http.MethodPost, "/api/v1/community/ask-ai", map[string]any{
		"question": "给狗狗选寄养机构最该确认什么？",
	}, loginBody.AccessToken)
	if communityAIResponse.Code != http.StatusOK {
		t.Fatalf("community ai status = %d want %d body=%s", communityAIResponse.Code, http.StatusOK, communityAIResponse.Body.String())
	}
	var communityAI dto.CommunityAIResponse
	decodeEnvelopeData(t, communityAIResponse.Body.Bytes(), &communityAI)
	if communityAI.Answer == "" {
		t.Fatalf("empty community ai response: %+v", communityAI)
	}
}

func TestPhaseFourTrainingShopAndNotificationFlow(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t)
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Skipf("skipping notification websocket test: cannot open local listener in this environment: %v", err)
	}

	server := httptest.NewUnstartedServer(router)
	server.Listener = listener
	server.Start()
	defer server.Close()

	loginBody := registerAndLogin(t, router, "13800138040")
	petBody := createPet(t, router, loginBody.AccessToken)

	dialer := websocket.Dialer{}
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+loginBody.AccessToken)
	conn, _, err := dialer.Dial("ws"+server.URL[len("http"):]+"/api/v1/ws", headers)
	if err != nil {
		t.Fatalf("dial notification websocket: %v", err)
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var syncMessage dto.NotificationStreamMessage
	if err := conn.ReadJSON(&syncMessage); err != nil {
		t.Fatalf("read notification sync: %v", err)
	}
	if syncMessage.Type != "notification_sync" {
		t.Fatalf("unexpected notification sync message: %+v", syncMessage)
	}

	generateTrainingResponse := performJSONRequest(t, router, http.MethodPost, "/api/v1/pets/"+petBody.ID+"/training/generate", map[string]any{
		"goal":       "学会坐下等待",
		"difficulty": "beginner",
		"category":   "obedience",
	}, loginBody.AccessToken)
	if generateTrainingResponse.Code != http.StatusCreated {
		t.Fatalf("generate training status = %d want %d body=%s", generateTrainingResponse.Code, http.StatusCreated, generateTrainingResponse.Body.String())
	}

	var trainingPlan dto.TrainingPlanResponse
	decodeEnvelopeData(t, generateTrainingResponse.Body.Bytes(), &trainingPlan)
	if !trainingPlan.AIGenerated || len(trainingPlan.Steps) == 0 {
		t.Fatalf("unexpected generated training plan: %+v", trainingPlan)
	}

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var notificationMessage dto.NotificationStreamMessage
	if err := conn.ReadJSON(&notificationMessage); err != nil {
		t.Fatalf("read training notification: %v", err)
	}
	if notificationMessage.Type != "notification" || notificationMessage.Notification == nil {
		t.Fatalf("unexpected notification message: %+v", notificationMessage)
	}

	listTrainingResponse := performJSONRequest(t, router, http.MethodGet, "/api/v1/pets/"+petBody.ID+"/training", nil, loginBody.AccessToken)
	if listTrainingResponse.Code != http.StatusOK {
		t.Fatalf("list training status = %d want %d body=%s", listTrainingResponse.Code, http.StatusOK, listTrainingResponse.Body.String())
	}
	var trainingPlans []dto.TrainingPlanResponse
	decodeEnvelopeData(t, listTrainingResponse.Body.Bytes(), &trainingPlans)
	if len(trainingPlans) == 0 {
		t.Fatal("training list should include generated plan")
	}

	updateTrainingResponse := performJSONRequest(t, router, http.MethodPut, "/api/v1/training/"+trainingPlan.ID, map[string]any{
		"progress": 45,
	}, loginBody.AccessToken)
	if updateTrainingResponse.Code != http.StatusOK {
		t.Fatalf("update training status = %d want %d body=%s", updateTrainingResponse.Code, http.StatusOK, updateTrainingResponse.Body.String())
	}

	getTrainingResponse := performJSONRequest(t, router, http.MethodGet, "/api/v1/training/"+trainingPlan.ID, nil, loginBody.AccessToken)
	if getTrainingResponse.Code != http.StatusOK {
		t.Fatalf("get training status = %d want %d body=%s", getTrainingResponse.Code, http.StatusOK, getTrainingResponse.Body.String())
	}
	var resolvedPlan dto.TrainingPlanResponse
	decodeEnvelopeData(t, getTrainingResponse.Body.Bytes(), &resolvedPlan)
	if resolvedPlan.Progress != 45 {
		t.Fatalf("unexpected training progress: %+v", resolvedPlan)
	}

	productsResponse := performJSONRequest(t, router, http.MethodGet, "/api/v1/shop/products?category=food", nil, loginBody.AccessToken)
	if productsResponse.Code != http.StatusOK {
		t.Fatalf("list products status = %d want %d body=%s", productsResponse.Code, http.StatusOK, productsResponse.Body.String())
	}
	var products []dto.ProductResponse
	decodeEnvelopeData(t, productsResponse.Body.Bytes(), &products)
	if len(products) == 0 {
		t.Fatal("shop product list should return seeded products")
	}

	productDetailResponse := performJSONRequest(t, router, http.MethodGet, "/api/v1/shop/products/"+products[0].ID, nil, loginBody.AccessToken)
	if productDetailResponse.Code != http.StatusOK {
		t.Fatalf("product detail status = %d want %d body=%s", productDetailResponse.Code, http.StatusOK, productDetailResponse.Body.String())
	}

	recommendationsResponse := performJSONRequest(t, router, http.MethodGet, "/api/v1/shop/recommendations/"+petBody.ID, nil, loginBody.AccessToken)
	if recommendationsResponse.Code != http.StatusOK {
		t.Fatalf("recommendations status = %d want %d body=%s", recommendationsResponse.Code, http.StatusOK, recommendationsResponse.Body.String())
	}
	var recommendations []dto.ProductResponse
	decodeEnvelopeData(t, recommendationsResponse.Body.Bytes(), &recommendations)
	if len(recommendations) == 0 || recommendations[0].RecommendedReason == "" {
		t.Fatalf("unexpected recommendations payload: %+v", recommendations)
	}

	notificationsResponse := performJSONRequest(t, router, http.MethodGet, "/api/v1/notifications?page=1&page_size=20", nil, loginBody.AccessToken)
	if notificationsResponse.Code != http.StatusOK {
		t.Fatalf("notifications status = %d want %d body=%s", notificationsResponse.Code, http.StatusOK, notificationsResponse.Body.String())
	}
	var notifications []dto.NotificationResponse
	decodeEnvelopeData(t, notificationsResponse.Body.Bytes(), &notifications)
	if len(notifications) == 0 {
		t.Fatal("notifications list should not be empty")
	}

	readNotificationResponse := performJSONRequest(t, router, http.MethodPut, "/api/v1/notifications/"+notifications[0].ID+"/read", nil, loginBody.AccessToken)
	if readNotificationResponse.Code != http.StatusOK {
		t.Fatalf("read notification status = %d want %d body=%s", readNotificationResponse.Code, http.StatusOK, readNotificationResponse.Body.String())
	}

	readAllResponse := performJSONRequest(t, router, http.MethodPut, "/api/v1/notifications/read-all", nil, loginBody.AccessToken)
	if readAllResponse.Code != http.StatusOK {
		t.Fatalf("read all notifications status = %d want %d body=%s", readAllResponse.Code, http.StatusOK, readAllResponse.Body.String())
	}
}

type envelope[T any] struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Meta    json.RawMessage `json:"meta"`
}

func performJSONRequest(t *testing.T, handler http.Handler, method, path string, payload any, token string) *httptest.ResponseRecorder {
	t.Helper()

	var body io.Reader
	if payload != nil {
		rawBody, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		body = bytes.NewReader(rawBody)
	}

	request := httptest.NewRequest(method, path, body)
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func performMultipartRequest(
	t *testing.T,
	handler http.Handler,
	method,
	path,
	fieldName,
	fileName string,
	content []byte,
	contentType,
	token string,
) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	partHeader := textproto.MIMEHeader{}
	partHeader.Set("Content-Disposition", `form-data; name="`+fieldName+`"; filename="`+fileName+`"`)
	partHeader.Set("Content-Type", contentType)
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		t.Fatalf("create multipart part: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write multipart content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(method, path, &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func decodeEnvelopeData[T any](t *testing.T, rawBody []byte, target *T) {
	t.Helper()

	var response envelope[T]
	if err := json.Unmarshal(rawBody, &response); err != nil {
		t.Fatalf("unmarshal envelope: %v body=%s", err, string(rawBody))
	}
	if err := json.Unmarshal(response.Data, target); err != nil {
		t.Fatalf("unmarshal envelope data: %v body=%s", err, string(rawBody))
	}
}

func createSQLiteSchema(db *gorm.DB) error {
	userSchema := `
CREATE TABLE users (
	id TEXT PRIMARY KEY,
	phone TEXT UNIQUE,
	email TEXT UNIQUE,
	password TEXT NOT NULL,
	nickname TEXT NOT NULL,
	avatar_url TEXT NOT NULL DEFAULT '',
	role TEXT NOT NULL DEFAULT 'user',
	wechat_open_id TEXT UNIQUE,
	apple_id TEXT UNIQUE,
	google_id TEXT UNIQUE,
	latitude REAL,
	longitude REAL,
	plan_type TEXT NOT NULL DEFAULT 'free',
	plan_expiry DATETIME,
	created_at DATETIME,
	updated_at DATETIME,
	deleted_at DATETIME
);`

	petSchema := `
CREATE TABLE pets (
	id TEXT PRIMARY KEY,
	owner_id TEXT NOT NULL,
	name TEXT NOT NULL,
	species TEXT NOT NULL,
	breed TEXT NOT NULL DEFAULT '',
	gender TEXT NOT NULL DEFAULT '',
	birth_date DATETIME,
	weight REAL,
	avatar_url TEXT NOT NULL DEFAULT '',
	microchip TEXT,
	is_neutered BOOLEAN NOT NULL DEFAULT FALSE,
	allergies JSON NOT NULL DEFAULT '[]',
	notes TEXT NOT NULL DEFAULT '',
	health_score INTEGER,
	created_at DATETIME,
	updated_at DATETIME,
	deleted_at DATETIME,
	FOREIGN KEY(owner_id) REFERENCES users(id)
);`

	healthRecordSchema := `
CREATE TABLE health_records (
	id TEXT PRIMARY KEY,
	pet_id TEXT NOT NULL,
	type TEXT NOT NULL,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	data JSON NOT NULL DEFAULT '{}',
	due_date DATETIME,
	provider_id TEXT,
	attachments JSON NOT NULL DEFAULT '[]',
	recorded_at DATETIME NOT NULL,
	created_at DATETIME,
	updated_at DATETIME,
	deleted_at DATETIME,
	FOREIGN KEY(pet_id) REFERENCES pets(id)
);`

	healthAlertSchema := `
CREATE TABLE health_alerts (
	id TEXT PRIMARY KEY,
	pet_id TEXT NOT NULL,
	alert_type TEXT NOT NULL,
	severity TEXT NOT NULL,
	title TEXT NOT NULL,
	message TEXT NOT NULL,
	source TEXT NOT NULL DEFAULT 'ai',
	is_read BOOLEAN NOT NULL DEFAULT FALSE,
	is_dismissed BOOLEAN NOT NULL DEFAULT FALSE,
	created_at DATETIME,
	updated_at DATETIME,
	deleted_at DATETIME,
	FOREIGN KEY(pet_id) REFERENCES pets(id)
);`

	deviceSchema := `
CREATE TABLE devices (
	id TEXT PRIMARY KEY,
	owner_id TEXT NOT NULL,
	pet_id TEXT,
	device_type TEXT NOT NULL,
	brand TEXT NOT NULL DEFAULT '',
	model TEXT NOT NULL DEFAULT '',
	nickname TEXT NOT NULL DEFAULT '',
	serial_number TEXT NOT NULL UNIQUE,
	status TEXT NOT NULL DEFAULT 'offline',
	config JSON NOT NULL DEFAULT '{}',
	last_seen DATETIME,
	firmware_ver TEXT NOT NULL DEFAULT '',
	battery_level INTEGER,
	created_at DATETIME,
	updated_at DATETIME,
	deleted_at DATETIME,
	FOREIGN KEY(owner_id) REFERENCES users(id),
	FOREIGN KEY(pet_id) REFERENCES pets(id)
);`

	deviceDataSchema := `
CREATE TABLE device_data_points (
	id TEXT PRIMARY KEY,
	time DATETIME NOT NULL,
	device_id TEXT NOT NULL,
	metric TEXT NOT NULL,
	value REAL NOT NULL,
	unit TEXT NOT NULL DEFAULT '',
	meta JSON NOT NULL DEFAULT '{}',
	FOREIGN KEY(device_id) REFERENCES devices(id)
);`

	serviceProviderSchema := `
CREATE TABLE service_providers (
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	type TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	address TEXT NOT NULL DEFAULT '',
	latitude REAL NOT NULL DEFAULT 0,
	longitude REAL NOT NULL DEFAULT 0,
	phone TEXT NOT NULL DEFAULT '',
	photos JSON NOT NULL DEFAULT '[]',
	rating REAL NOT NULL DEFAULT 0,
	review_count INTEGER NOT NULL DEFAULT 0,
	is_verified BOOLEAN NOT NULL DEFAULT FALSE,
	open_hours JSON NOT NULL DEFAULT '{}',
	services JSON NOT NULL DEFAULT '[]',
	tags JSON NOT NULL DEFAULT '[]',
	created_at DATETIME,
	updated_at DATETIME,
	deleted_at DATETIME,
	FOREIGN KEY(user_id) REFERENCES users(id)
);`

	bookingSchema := `
CREATE TABLE bookings (
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL,
	pet_id TEXT NOT NULL,
	provider_id TEXT NOT NULL,
	service_name TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'pending',
	start_time DATETIME NOT NULL,
	end_time DATETIME,
	price INTEGER NOT NULL DEFAULT 0,
	currency TEXT NOT NULL DEFAULT 'CNY',
	notes TEXT NOT NULL DEFAULT '',
	cancel_reason TEXT NOT NULL DEFAULT '',
	rating INTEGER,
	review TEXT NOT NULL DEFAULT '',
	created_at DATETIME,
	updated_at DATETIME,
	deleted_at DATETIME,
	FOREIGN KEY(user_id) REFERENCES users(id),
	FOREIGN KEY(pet_id) REFERENCES pets(id),
	FOREIGN KEY(provider_id) REFERENCES service_providers(id)
);`

	postSchema := `
CREATE TABLE posts (
	id TEXT PRIMARY KEY,
	author_id TEXT NOT NULL,
	pet_id TEXT,
	type TEXT NOT NULL DEFAULT 'post',
	title TEXT NOT NULL DEFAULT '',
	content TEXT NOT NULL,
	images JSON NOT NULL DEFAULT '[]',
	tags JSON NOT NULL DEFAULT '[]',
	latitude REAL,
	longitude REAL,
	like_count INTEGER NOT NULL DEFAULT 0,
	comment_count INTEGER NOT NULL DEFAULT 0,
	is_published BOOLEAN NOT NULL DEFAULT TRUE,
	created_at DATETIME,
	updated_at DATETIME,
	deleted_at DATETIME,
	FOREIGN KEY(author_id) REFERENCES users(id),
	FOREIGN KEY(pet_id) REFERENCES pets(id)
);`

	commentSchema := `
CREATE TABLE comments (
	id TEXT PRIMARY KEY,
	post_id TEXT NOT NULL,
	author_id TEXT NOT NULL,
	parent_id TEXT,
	content TEXT NOT NULL,
	like_count INTEGER NOT NULL DEFAULT 0,
	created_at DATETIME,
	updated_at DATETIME,
	deleted_at DATETIME,
	FOREIGN KEY(post_id) REFERENCES posts(id),
	FOREIGN KEY(author_id) REFERENCES users(id)
);`

	postLikeSchema := `
CREATE TABLE post_likes (
	user_id TEXT NOT NULL,
	post_id TEXT NOT NULL,
	created_at DATETIME,
	PRIMARY KEY (user_id, post_id),
	FOREIGN KEY(user_id) REFERENCES users(id),
	FOREIGN KEY(post_id) REFERENCES posts(id)
);`

	trainingPlanSchema := `
CREATE TABLE training_plans (
	id TEXT PRIMARY KEY,
	pet_id TEXT NOT NULL,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	difficulty TEXT NOT NULL DEFAULT 'beginner',
	category TEXT NOT NULL DEFAULT 'obedience',
	steps JSON NOT NULL DEFAULT '[]',
	ai_generated BOOLEAN NOT NULL DEFAULT FALSE,
	progress INTEGER NOT NULL DEFAULT 0,
	created_at DATETIME,
	updated_at DATETIME,
	deleted_at DATETIME,
	FOREIGN KEY(pet_id) REFERENCES pets(id)
);`

	productSchema := `
CREATE TABLE products (
	id TEXT PRIMARY KEY,
	provider_id TEXT,
	name TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	category TEXT NOT NULL DEFAULT '',
	price INTEGER NOT NULL DEFAULT 0,
	currency TEXT NOT NULL DEFAULT 'CNY',
	images JSON NOT NULL DEFAULT '[]',
	pet_species JSON NOT NULL DEFAULT '[]',
	tags JSON NOT NULL DEFAULT '[]',
	external_url TEXT,
	rating REAL NOT NULL DEFAULT 0,
	is_active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at DATETIME,
	updated_at DATETIME,
	deleted_at DATETIME
);`

	notificationSchema := `
CREATE TABLE notifications (
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL,
	type TEXT NOT NULL,
	title TEXT NOT NULL,
	body TEXT NOT NULL DEFAULT '',
	data JSON NOT NULL DEFAULT '{}',
	is_read BOOLEAN NOT NULL DEFAULT FALSE,
	created_at DATETIME,
	updated_at DATETIME,
	deleted_at DATETIME,
	FOREIGN KEY(user_id) REFERENCES users(id)
);`

	pushTokenSchema := `
CREATE TABLE push_tokens (
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL,
	provider TEXT NOT NULL DEFAULT 'expo',
	token TEXT NOT NULL UNIQUE,
	platform TEXT NOT NULL DEFAULT 'unknown',
	is_active BOOLEAN NOT NULL DEFAULT TRUE,
	last_seen_at DATETIME NOT NULL,
	created_at DATETIME,
	updated_at DATETIME,
	deleted_at DATETIME,
	FOREIGN KEY(user_id) REFERENCES users(id)
);`

	if err := db.Exec(userSchema).Error; err != nil {
		return err
	}
	if err := db.Exec(petSchema).Error; err != nil {
		return err
	}
	if err := db.Exec(healthRecordSchema).Error; err != nil {
		return err
	}
	if err := db.Exec(healthAlertSchema).Error; err != nil {
		return err
	}
	if err := db.Exec(deviceSchema).Error; err != nil {
		return err
	}
	if err := db.Exec(deviceDataSchema).Error; err != nil {
		return err
	}
	if err := db.Exec(serviceProviderSchema).Error; err != nil {
		return err
	}
	if err := db.Exec(bookingSchema).Error; err != nil {
		return err
	}
	if err := db.Exec(postSchema).Error; err != nil {
		return err
	}
	if err := db.Exec(commentSchema).Error; err != nil {
		return err
	}
	if err := db.Exec(postLikeSchema).Error; err != nil {
		return err
	}
	if err := db.Exec(trainingPlanSchema).Error; err != nil {
		return err
	}
	if err := db.Exec(productSchema).Error; err != nil {
		return err
	}
	if err := db.Exec(notificationSchema).Error; err != nil {
		return err
	}
	if err := db.Exec(pushTokenSchema).Error; err != nil {
		return err
	}
	return nil
}

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()

	dsn := "file:" + uuid.NewString() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := createSQLiteSchema(db); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	tokenManager := petjwt.NewManager("integration-secret", "petverse-test", 15*time.Minute, 24*time.Hour)
	userRepo := repository.NewUserRepository(db)
	petRepo := repository.NewPetRepository(db)
	healthRepo := repository.NewHealthRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	serviceProviderRepo := repository.NewServiceProviderRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	communityRepo := repository.NewCommunityRepository(db)
	trainingRepo := repository.NewTrainingRepository(db)
	shopRepo := repository.NewShopRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	pushTokenRepo := repository.NewPushTokenRepository(db)
	wsHub := ws.NewHub()
	uploadDir := t.TempDir()
	uploader := upload.NewLocalStore(uploadDir)
	notificationService := service.NewNotificationService(
		notificationRepo,
		wsHub,
		service.WithPushNotifications(pushTokenRepo, nil),
	)

	return NewRouter(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		tokenManager,
		handler.NewAuthHandler(service.NewAuthService(userRepo, tokenManager)),
		handler.NewUserHandler(service.NewUserService(userRepo), uploader),
		handler.NewPetHandler(service.NewPetService(petRepo), uploader),
		handler.NewHealthHandler(service.NewHealthService(petRepo, healthRepo, deviceRepo, ai.NewHealthAI())),
		handler.NewDeviceHandler(service.NewDeviceService(deviceRepo, petRepo, wsHub), wsHub),
		handler.NewServiceMarketHandler(service.NewServiceMarketService(serviceProviderRepo)),
		handler.NewBookingHandler(service.NewBookingService(bookingRepo, serviceProviderRepo, petRepo)),
		handler.NewCommunityHandler(service.NewCommunityService(communityRepo, petRepo)),
		handler.NewTrainingHandler(service.NewTrainingService(trainingRepo, petRepo, notificationService)),
		handler.NewShopHandler(service.NewShopService(shopRepo, petRepo)),
		handler.NewNotificationHandler(notificationService, wsHub),
		handler.NewUploadHandler(uploader),
		uploadDir,
	)
}

func registerAndLogin(t *testing.T, router http.Handler, phone string) dto.AuthResponse {
	t.Helper()

	registerResponse := performJSONRequest(t, router, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"phone":    phone,
		"password": "strong-pass",
		"nickname": "Tester",
	}, "")
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("register status = %d want %d body=%s", registerResponse.Code, http.StatusCreated, registerResponse.Body.String())
	}

	loginResponse := performJSONRequest(t, router, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"phone":    phone,
		"password": "strong-pass",
	}, "")
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("login status = %d want %d body=%s", loginResponse.Code, http.StatusOK, loginResponse.Body.String())
	}

	var loginBody dto.AuthResponse
	decodeEnvelopeData(t, loginResponse.Body.Bytes(), &loginBody)
	return loginBody
}

func createPet(t *testing.T, router http.Handler, token string) dto.PetResponse {
	t.Helper()

	createPetResponse := performJSONRequest(t, router, http.MethodPost, "/api/v1/pets", map[string]any{
		"name":        "DouDou",
		"species":     "dog",
		"breed":       "corgi",
		"gender":      "female",
		"is_neutered": true,
		"allergies":   []string{"beef"},
		"notes":       "friendly",
	}, token)
	if createPetResponse.Code != http.StatusCreated {
		t.Fatalf("create pet status = %d want %d body=%s", createPetResponse.Code, http.StatusCreated, createPetResponse.Body.String())
	}

	var petBody dto.PetResponse
	decodeEnvelopeData(t, createPetResponse.Body.Bytes(), &petBody)
	return petBody
}

func mustPNG(t *testing.T) []byte {
	t.Helper()

	var buffer bytes.Buffer
	imageData := image.NewRGBA(image.Rect(0, 0, 1, 1))
	imageData.Set(0, 0, color.RGBA{R: 16, G: 185, B: 129, A: 255})
	if err := png.Encode(&buffer, imageData); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buffer.Bytes()
}
