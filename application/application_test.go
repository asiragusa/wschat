package application

import (
	"cloud.google.com/go/datastore"
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/iris-contrib/httpexpect"
	"github.com/kataras/iris"
	"github.com/kataras/iris/httptest"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/websocket"
	"strings"
	"testing"
	"time"
)

type ApplicationTestSuite struct {
	suite.Suite
	app *Application
	e   *httpexpect.Expect
}

func TestApplication(t *testing.T) {
	suite.Run(t, new(ApplicationTestSuite))
}

// Cleanup db
func (suite *ApplicationTestSuite) wipeoutDatastoreData(client *datastore.Client) {
	query := datastore.NewQuery("").KeysOnly()
	ctx := context.Background()

	keys, err := client.GetAll(ctx, query, nil)
	suite.Require().NoError(err)

	err = client.DeleteMulti(ctx, keys)
	suite.Require().NoError(err)
}

func getDatastoreClient(projectID string) (*datastore.Client, error) {
	ctx := context.Background()

	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func getPubsubClient(projectID string) (*pubsub.Client, error) {
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (suite *ApplicationTestSuite) SetupSuite() {
	datastoreClient, err := getDatastoreClient("test")
	suite.Require().NoError(err)

	pubsubClient, err := getPubsubClient("test")
	suite.Require().NoError(err)

	appConfig := &AppConfig{
		JwtSecret:       "secret",
		JwtIssuer:       "http://localhost",
		DatastoreClient: datastoreClient,
		PubsubClient:    pubsubClient,
	}

	app, err := NewApplication(appConfig)
	suite.Require().NoError(err)
	suite.Require().NotNil(app)

	suite.app = app
	suite.e = httptest.New(suite.T(), app.GetRouter())

	// Init the http router, for websocket testing
	go func() {
		suite.app.irisApp.Run(iris.Addr(":8080"), iris.WithoutStartupLog, iris.WithoutServerError(iris.ErrServerClosed))
	}()
}

func (suite *ApplicationTestSuite) SetupTest() {
	suite.wipeoutDatastoreData(suite.app.config.DatastoreClient)
}

func (suite *ApplicationTestSuite) TearDownSuite() {
	suite.app.irisApp.Shutdown(context.Background())
}

// Add auth header for the request
func (suite *ApplicationTestSuite) authorize(request *httpexpect.Request, token string) {
	request.WithHeader("Authorization", fmt.Sprintf("Bearer %s", token))
}

var (
	defaultEmail    = "test@test.com"
	defaultPassword = "validPassword"
)

// Register the user given an email
func (suite *ApplicationTestSuite) validRegisterWithUser(email string) string {
	request := suite.e.POST("/register").WithJSON(map[string]string{
		"email":    email,
		"password": defaultPassword,
	})

	expect := request.Expect()
	expect.Status(httptest.StatusCreated)

	json := expect.JSON().Object()
	json.Value("accessToken").String().NotEmpty()

	return json.Value("accessToken").String().Raw()
}

// Registers the default user
func (suite *ApplicationTestSuite) validRegister() string {
	return suite.validRegisterWithUser(defaultEmail)
}

// Creates a message
func (suite *ApplicationTestSuite) createMessage(token, to string) *httpexpect.Object {
	requestData := map[string]interface{}{
		"to":      to,
		"message": "test",
	}

	request := suite.e.POST("/messages").WithJSON(requestData)
	suite.authorize(request, token)

	expect := request.Expect()
	expect.Status(httptest.StatusCreated)

	return expect.JSON().Object()
}

// Test GET /
func (suite *ApplicationTestSuite) TestGetIndex() {
	suite.e.GET("/").Expect().Status(httptest.StatusOK)
}

// Test POST /register with invalid data
func (suite *ApplicationTestSuite) TestUnprocessableRegister() {
	arg := map[string]interface{}{
		"email":    "test",
		"password": "pwd",
	}
	request := suite.e.POST("/register").WithJSON(arg)

	expect := request.Expect()
	expect.Status(httptest.StatusUnprocessableEntity)

	json := expect.JSON().Object()
	json.Value("code").Equal(httptest.StatusUnprocessableEntity)
	json.Value("message").Equal("Unprocessable Entity")

	json.Value("details").Object().
		ContainsKey("email").
		ContainsKey("password")
}

// Test POST /register with existing user
func (suite *ApplicationTestSuite) TestRegisterExistingUser() {
	suite.validRegister()

	request := suite.e.POST("/register").WithJSON(map[string]string{
		"email":    defaultEmail,
		"password": defaultPassword,
	})
	expect := request.Expect()

	expect.Status(httptest.StatusUnprocessableEntity)
	json := expect.JSON().Object()
	json.Equal(map[string]interface{}{
		"code":    httptest.StatusUnprocessableEntity,
		"message": "Unprocessable Entity",
		"details": map[string]interface{}{
			"email": []string{"alreadyExists"},
		},
	})
}

// Test POST /register with a valid user
func (suite *ApplicationTestSuite) TestRegister() {
	suite.validRegister()
}

// Test POST /login with bad password
func (suite *ApplicationTestSuite) TestLoginBadPassword() {
	loginRequest := map[string]string{
		"email":    defaultEmail,
		"password": "invalidPassword",
	}

	request := suite.e.POST("/login").WithJSON(loginRequest)

	expect := request.Expect()
	expect.Status(httptest.StatusUnauthorized)

	json := expect.JSON().Object()
	json.Equal(map[string]interface{}{
		"code":    httptest.StatusUnauthorized,
		"message": "Unauthorized",
	})
}

// Test POST /login OK
func (suite *ApplicationTestSuite) TestLoginOK() {
	suite.validRegister()

	loginRequest := map[string]string{
		"email":    defaultEmail,
		"password": defaultPassword,
	}
	request := suite.e.POST("/login").WithJSON(loginRequest)

	expect := request.Expect()
	expect.Status(httptest.StatusOK)

	json := expect.JSON().Object()
	json.Value("accessToken").String().NotEmpty()
}

// Test POST /users with bad credentials
func (suite *ApplicationTestSuite) TestListUsersUnauthorized() {
	request := suite.e.GET("/users")
	suite.authorize(request, "invalid")

	expect := request.Expect()
	expect.Status(httptest.StatusUnauthorized)

	json := expect.JSON().Object()
	json.Equal(map[string]interface{}{
		"code":    httptest.StatusUnauthorized,
		"message": "Unauthorized",
	})
}

// Test POST /users OK
func (suite *ApplicationTestSuite) TestListUsersOK() {
	token := suite.validRegister()

	request := suite.e.GET("/users")
	suite.authorize(request, token)

	expect := request.Expect()
	expect.Status(httptest.StatusOK)
	json := expect.JSON().Object()
	json.Equal(map[string]interface{}{
		"total": 1,
		"items": []map[string]interface{}{
			{"email": defaultEmail},
		},
	})
}

// Test POST /messages with bad credentials
func (suite *ApplicationTestSuite) TestCreateMessageUnauthorized() {
	request := suite.e.POST("/messages")
	suite.authorize(request, "invalid")

	expect := request.Expect()
	expect.Status(httptest.StatusUnauthorized)

	json := expect.JSON().Object()
	json.Equal(map[string]interface{}{
		"code":    httptest.StatusUnauthorized,
		"message": "Unauthorized",
	})
}

// Test POST /messages OK
func (suite *ApplicationTestSuite) TestCreateMessageOK() {
	token := suite.validRegister()
	suite.validRegisterWithUser("a@b.com")

	json := suite.createMessage(token, "a@b.com")
	json.Value("id").String().NotEmpty()
	json.Value("from").String().Equal(defaultEmail)
	json.Value("to").String().Equal("a@b.com")
	json.Value("message").String().NotEmpty()
	json.Value("createdAt").String().NotEmpty()
}

// Test GET /messages with bad credentials
func (suite *ApplicationTestSuite) TestListMessagesUnauthorized() {
	request := suite.e.GET("/messages")
	suite.authorize(request, "invalid")

	expect := request.Expect()
	expect.Status(httptest.StatusUnauthorized)

	json := expect.JSON().Object()
	json.Equal(map[string]interface{}{
		"code":    httptest.StatusUnauthorized,
		"message": "Unauthorized",
	})
}

// Test GET /messages ok
func (suite *ApplicationTestSuite) TestListMessagesOK() {
	token := suite.validRegister()
	token1 := suite.validRegisterWithUser("a@b.com")
	token2 := suite.validRegisterWithUser("b@b.com")

	suite.createMessage(token, "a@b.com")
	suite.createMessage(token1, "b@b.com")
	suite.createMessage(token2, defaultEmail)

	request := suite.e.GET("/messages")
	suite.authorize(request, token)

	expect := request.Expect()
	expect.Status(httptest.StatusOK)

	json := expect.JSON().Object()
	json.Value("total").Equal(2)

	items := json.Value("items").Array()

	message1 := items.Element(0).Object()
	message1.Value("id").String().NotEmpty()
	message1.Value("from").String().Equal(defaultEmail)
	message1.Value("to").String().Equal("a@b.com")
	message1.Value("message").String().NotEmpty()
	message1.Value("createdAt").String().NotEmpty()

	message2 := items.Element(1).Object()
	message2.Value("id").String().NotEmpty()
	message2.Value("from").String().Equal("b@b.com")
	message2.Value("to").String().Equal(defaultEmail)
	message2.Value("message").String().NotEmpty()
	message2.Value("createdAt").String().NotEmpty()
}

// Test POST /wsToken with bad credentials
func (suite *ApplicationTestSuite) TestCreateWsTokenUnauthorized() {
	request := suite.e.POST("/wsToken")
	suite.authorize(request, "invalid")

	expect := request.Expect()
	expect.Status(httptest.StatusUnauthorized)

	json := expect.JSON().Object()
	json.Equal(map[string]interface{}{
		"code":    httptest.StatusUnauthorized,
		"message": "Unauthorized",
	})
}

// Test POST /wsToken OK
func (suite *ApplicationTestSuite) TestCreateWsTokenOK() {
	token := suite.validRegister()

	request := suite.e.POST("/wsToken")
	suite.authorize(request, token)

	expect := request.Expect()
	expect.Status(httptest.StatusCreated)

	json := expect.JSON().Object()
	json.Value("token").String().NotEmpty()
}

// Helper method
func (suite *ApplicationTestSuite) getWsToken() string {
	token := suite.validRegister()

	request := suite.e.POST("/wsToken")
	suite.authorize(request, token)

	expect := request.Expect()
	expect.Status(httptest.StatusCreated)

	json := expect.JSON().Object()
	return json.Value("token").String().Raw()
}

// Helper method
func (suite *ApplicationTestSuite) sendMesasge(ws *websocket.Conn, event string, v interface{}) {
	req := map[string]interface{}{
		"requestId": "aRequestId",
		"body":      v,
	}

	encoded, err := json.Marshal(req)
	suite.Require().NoError(err)

	msg := fmt.Sprintf("iris-websocket-message:%s;4;%s", event, encoded)
	_, err = ws.Write([]byte(msg))

	suite.Require().NoError(err)
}

// Helper method
func (suite *ApplicationTestSuite) readMessage(ws *websocket.Conn) (string, map[string]interface{}) {
	var msg = make([]byte, 1024)

	n, err := ws.Read(msg)
	suite.Require().NoError(err)

	parts := strings.SplitN(string(msg[:n]), ":", 2)
	suite.Require().Len(parts, 2)
	suite.Require().Equal("iris-websocket-message", parts[0])

	parts = strings.SplitN(parts[1], ";", 3)
	suite.Require().Len(parts, 3)
	suite.Require().Equal(parts[1], "4")

	var v map[string]interface{}
	err = json.Unmarshal([]byte(parts[2]), &v)
	suite.Require().NoError(err)

	body, ok := v["body"].(map[string]interface{})
	suite.Require().True(ok)

	return parts[0], body
}

// Helper method
func (suite *ApplicationTestSuite) getWsConn() *websocket.Conn {
	wsToken := suite.getWsToken()

	origin := "http://localhost:8080/"
	url := fmt.Sprintf("ws://localhost:8080/ws?token=%s", wsToken)

	var ws *websocket.Conn
	var err error

	for i := 0; i < 10; i++ {
		ws, err = websocket.Dial(url, "", origin)
		time.Sleep(time.Millisecond * 500)
		if err == nil {
			return ws
		}
	}
	suite.Require().NoError(err)

	return ws
}

// Tests sending a message via websocket
func (suite *ApplicationTestSuite) TestSendWsMessage() {
	suite.validRegisterWithUser("a@b.com")

	ws := suite.getWsConn()
	suite.sendMesasge(ws, "message", map[string]interface{}{
		"to":      "a@b.com",
		"message": "test",
	})

	event, body := suite.readMessage(ws)
	suite.Require().Equal("sent", event)

	suite.NotEmpty(body["id"])
	suite.Equal(defaultEmail, body["from"])
	suite.Equal("a@b.com", body["to"])
	suite.Equal("test", body["message"])
	suite.NotEmpty(body["createdAt"])

	ws.Close()
}

// Tests receiving a message via websocket
func (suite *ApplicationTestSuite) TestReceiveWsMessage() {
	token := suite.validRegisterWithUser("a@b.com")
	ws := suite.getWsConn()

	go func() {
		time.Sleep(time.Second)
		suite.createMessage(token, defaultEmail)
	}()

	event, body := suite.readMessage(ws)
	suite.Require().Equal("message", event)

	suite.NotEmpty(body["id"])
	suite.Equal("a@b.com", body["from"])
	suite.Equal(defaultEmail, body["to"])
	suite.Equal("test", body["message"])
	suite.NotEmpty(body["createdAt"])
	ws.Close()
}

// Tests sending an invalid message
func (suite *ApplicationTestSuite) TestSendWsMessageError() {
	ws := suite.getWsConn()
	suite.sendMesasge(ws, "message", map[string]interface{}{
		"to":      "notExisting",
		"message": "test",
	})

	event, body := suite.readMessage(ws)
	suite.Require().Equal("error", event)

	suite.Equal("Unprocessable Entity", body["message"])
	ws.Close()
}
