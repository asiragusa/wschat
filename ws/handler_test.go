package ws

import (
	context2 "context"
	"encoding/json"
	"fmt"
	"github.com/asiragusa/wschat/entity"
	"github.com/asiragusa/wschat/mocks"
	"github.com/asiragusa/wschat/request"
	"github.com/asiragusa/wschat/response"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/httptest"
	"github.com/kataras/iris/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	websocket2 "golang.org/x/net/websocket"
	"strings"
	"testing"
	"time"
)

type MockCancel struct {
	mock.Mock
}

func (m *MockCancel) Call() {
	m.Called()
}

type HandlerTestSuite struct {
	suite.Suite
	handler    *Handler
	interactor *mocks.CreateMessageInteractor
	pubsub     *mocks.PubsubClient
	validator  *mocks.RequestValidator
	user       *entity.User
	cancel     *MockCancel
	app        *iris.Application
}

func TestListMessagesController(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func (suite *HandlerTestSuite) SetupSuite() {
	suite.handler = NewWsHandler()

	suite.user = &entity.User{
		Email: "a@b.com",
	}

	app := iris.New()
	suite.app = app
	app.Use(func(ctx context.Context) {
		ctx.Values().Set("user", suite.user)
		ctx.Next()
	})

	s := websocket.New(websocket.Config{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	})
	s.OnConnection(suite.handler.HandleConnection)

	app.Get("/", s.Handler())

	go func() {
		app.Run(iris.Addr(":8081"), iris.WithoutStartupLog, iris.WithoutServerError(iris.ErrServerClosed))
	}()
}

func (suite *HandlerTestSuite) SetupTest() {
	suite.interactor = &mocks.CreateMessageInteractor{}
	suite.pubsub = &mocks.PubsubClient{}
	suite.validator = &mocks.RequestValidator{}
	suite.cancel = &MockCancel{}

	suite.handler.CreateMessageInteractor = suite.interactor
	suite.handler.PubsubClient = suite.pubsub
	suite.handler.Validator = suite.validator
}

func (suite *HandlerTestSuite) TearDownTest() {
	suite.interactor.AssertExpectations(suite.T())
	suite.pubsub.AssertExpectations(suite.T())
	suite.validator.AssertExpectations(suite.T())
	suite.cancel.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TearDownSuite() {
	suite.app.Shutdown(context2.Background())
}

func (suite *HandlerTestSuite) getWsConn() *websocket2.Conn {
	origin := "http://localhost:8081/"
	url := "ws://localhost:8081/"

	var ws *websocket2.Conn
	var err error

	for i := 0; i < 10; i++ {
		ws, err = websocket2.Dial(url, "", origin)
		time.Sleep(time.Millisecond * 500)
		if err == nil {
			return ws
		}
	}
	suite.Require().NoError(err)

	return ws
}

func (suite *HandlerTestSuite) readMessage(ws *websocket2.Conn, dst interface{}) string {
	var msg = make([]byte, 1024)

	n, err := ws.Read(msg)
	suite.Require().NoError(err)

	parts := strings.SplitN(string(msg[:n]), ":", 2)
	suite.Require().Len(parts, 2)
	suite.Require().Equal("iris-websocket-message", parts[0])

	parts = strings.SplitN(parts[1], ";", 3)
	suite.Require().Len(parts, 3)
	suite.Require().Equal(parts[1], "4")

	err = json.Unmarshal([]byte(parts[2]), &dst)
	suite.Require().NoError(err)

	return parts[0]
}

func (suite *HandlerTestSuite) sendMesasge(ws *websocket2.Conn, event string, v interface{}) {
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

func (suite *HandlerTestSuite) TestSubscriberAnError() {
	suite.pubsub.On("Subscribe", suite.user.Email, mock.Anything).Return(assert.AnError, nil)
	conn := suite.getWsConn()

	err := conn.Close()
	suite.Require().NoError(err)
}

func (suite *HandlerTestSuite) TestDisconnect() {
	suite.pubsub.On("Subscribe", suite.user.Email, mock.Anything).Return(nil, suite.cancel.Call)
	suite.cancel.On("Call")

	conn := suite.getWsConn()

	err := conn.Close()

	time.Sleep(time.Millisecond * 100)

	suite.Require().NoError(err)
}

func (suite *HandlerTestSuite) TestReceiveMessage() {
	suite.cancel.On("Call")

	var theFn func(entity.Message)
	suite.pubsub.On("Subscribe", suite.user.Email, mock.MatchedBy(func(fn func(entity.Message)) bool {
		theFn = fn
		return true
	})).Return(nil, suite.cancel.Call)

	conn := suite.getWsConn()

	message := entity.Message{
		Id:      "id",
		From:    "from",
		To:      "to",
		Message: "message",
	}
	suite.Require().NotNil(theFn)
	theFn(message)

	type Res struct {
		Body entity.Message `json:"body"`
	}

	var msg Res
	event := suite.readMessage(conn, &msg)
	suite.Require().Equal("message", event)

	suite.Equal(message, msg.Body)

	err := conn.Close()

	time.Sleep(time.Millisecond * 100)

	suite.Require().NoError(err)
}

func (suite *HandlerTestSuite) TestSendMessageInvalid() {
	suite.pubsub.On("Subscribe", suite.user.Email, mock.Anything).Return(nil, suite.cancel.Call)
	suite.cancel.On("Call")

	conn := suite.getWsConn()

	// No action should be done
	conn.Write([]byte(`test`))
	conn.Write([]byte(`iris-websocket-message:%s;4;ko`))
	conn.Write([]byte(`iris-websocket-message:%s;4;{}`))

	// Testing validation
	req := request.CreateMessage{
		From: *suite.user,
	}

	suite.validator.On("Struct", req).Return(assert.AnError)
	suite.validator.On("FormatError", assert.AnError).Return(response.NewError(httptest.StatusUnprocessableEntity))
	suite.sendMesasge(conn, "message", map[string]interface{}{
		"to": "",
	})

	var res WsResponse
	event := suite.readMessage(conn, &res)
	suite.Require().Equal("error", event)

	// Closing the connection
	err := conn.Close()
	suite.Require().NoError(err)

	time.Sleep(time.Millisecond * 100)
}

func (suite *HandlerTestSuite) TestSendMessageInteractorError() {
	suite.pubsub.On("Subscribe", suite.user.Email, mock.Anything).Return(nil, suite.cancel.Call)
	suite.cancel.On("Call")

	conn := suite.getWsConn()

	req := request.CreateMessage{
		From:    *suite.user,
		To:      "a@b.com",
		Message: "message",
	}
	suite.validator.On("Struct", req).Return(nil)
	suite.interactor.On("Call", req).Return(response.NewError(httptest.StatusInternalServerError))
	suite.sendMesasge(conn, "message", map[string]interface{}{
		"to":      "a@b.com",
		"message": "message",
	})

	var res WsResponse
	event := suite.readMessage(conn, &res)
	suite.Require().Equal("error", event)

	err := conn.Close()
	suite.Require().NoError(err)

	time.Sleep(time.Millisecond * 100)
}

func (suite *HandlerTestSuite) TestSendMessageOK() {
	suite.pubsub.On("Subscribe", suite.user.Email, mock.Anything).Return(nil, suite.cancel.Call)
	suite.cancel.On("Call")

	conn := suite.getWsConn()

	req := request.CreateMessage{
		From:    *suite.user,
		To:      "a@b.com",
		Message: "message",
	}
	createMessageResponse := response.CreateMessage{
		Id:      "id",
		From:    "from",
		To:      "to",
		Message: "message",
	}

	suite.validator.On("Struct", req).Return(nil)
	suite.interactor.On("Call", req).Return(createMessageResponse)
	suite.sendMesasge(conn, "message", map[string]interface{}{
		"to":      "a@b.com",
		"message": "message",
	})

	type Success struct {
		RequestId string                 `json:"requestId"`
		Body      response.CreateMessage `json:"body"`
	}

	var res Success
	event := suite.readMessage(conn, &res)
	suite.Require().Equal("sent", event)

	suite.Equal("aRequestId", res.RequestId)
	suite.Equal(createMessageResponse, res.Body)

	err := conn.Close()
	suite.Require().NoError(err)

	time.Sleep(time.Millisecond * 100)
}
