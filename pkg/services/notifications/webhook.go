package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/context/ctxhttp"

	"github.com/grafana/grafana/pkg/log"
	"github.com/grafana/grafana/pkg/services/alerting"
	"github.com/grafana/grafana/pkg/util"
)

//var app = "16649b46-946d-57e7-a624-292ed50306d2"
//var eventType_ = "trigger"
var eventId = "123456"

type Webhook struct {
	Url        string
	User       string
	Password   string
	Key        string
	Body       string
	HttpMethod string
	HttpHeader map[string]string
}

//type EvalMatche struct {
//	Value  float64           `json:"value"`
//	Metric string            `json:"metric"`
//	Tags   map[string]string `json:"tags"`
//}

type StateType string

const (
	ALERTING StateType = "alerting"
	OK       StateType = "ok"
	NODATA   StateType = "no_data"
	ERROR    StateType = "error"
)

type BodyJson struct {
	Title       string               `json:"title"`
	RuleId      int                  `json:"ruleId"`
	RuleName    string               `json:"ruleName"`
	State       string               `json:"state"`
	RuleUrl     string               `json:"ruleUrl"`
	ImageUrl    string               `json:"imageUrl"`
	Message     string               `json:"message"`
	EvalMatches []alerting.EvalMatch `json:"evalMatches"`
}

type OneAlertJson struct {
	App          string `json:"app"`
	AlarmName    string `json:"alarmName"`
	AlarmContent string `json:"alarmContent"`
	EventType    string `json:"eventType"`
	EventId      string `json:"eventId"`
}

var netTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	Dial: (&net.Dialer{
		Timeout: 30 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 5 * time.Second,
}
var netClient = &http.Client{
	Timeout:   time.Second * 30,
	Transport: netTransport,
}

var (
	webhookQueue chan *Webhook
	webhookLog   log.Logger
)

func initWebhookQueue() {
	webhookLog = log.New("notifications.webhook")
	webhookQueue = make(chan *Webhook, 10)
	go processWebhookQueue()
}

func processWebhookQueue() {
	for {
		select {
		case webhook := <-webhookQueue:
			err := sendWebRequestSync(context.Background(), webhook)

			if err != nil {
				webhookLog.Error("Failed to send webrequest ", "error", err)
			}
		}
	}
}

func sendWebRequestSync(ctx context.Context, webhook *Webhook) error {
	webhookLog.Info("Sending webhook", "url", webhook.Url, "http method", webhook.HttpMethod,
		"user", webhook.User, "password", webhook.Password, "body", webhook.Body)

	if webhook.Key == "" {
		return nil
	}
	if webhook.HttpMethod == "" {
		webhook.HttpMethod = http.MethodPost
	}
	bodyJson := &BodyJson{}
	err := json.Unmarshal([]byte(webhook.Body), bodyJson)
	if err != nil {
		//webhook.Url = webhook.Url + "app=" + app + "&eventType=" + eventType + "&eventId=" + eventId + "&alarmName=" + bodyJson.RuleName + "&alarmContent=" + bodyJson.Message
		webhookLog.Error(err.Error())
	}
	var eventType string
	var alarmContent string
	if bodyJson.State == "ok" {
		eventType = "resolve"
		alarmContent = "[RESOLVED] " + bodyJson.Message
	} else if bodyJson.State == "alerting" {
		eventType = "trigger"
		var buf bytes.Buffer
		buf.WriteString("[ALERTING] " + bodyJson.Message)
		for _, eval := range bodyJson.EvalMatches {
			if eval.Tags != nil {
				for _, v := range eval.Tags {
					buf.WriteString(" ")
					buf.WriteString(v)
				}
			} else {
				buf.WriteString(" ")
				buf.WriteString(eval.Metric)
				buf.WriteString(" : ")
				//parse float64 to string
				buf.WriteString(eval.Value.String())
				buf.WriteString(" ")
			}
		}
		alarmContent = buf.String()
	} else {
		return nil
	}
	//webhookLog.Info("---------- alarmContent : " + alarmContent + "-----------")
	alarmJson := &OneAlertJson{
		App:          webhook.Key,
		AlarmName:    bodyJson.RuleName,
		AlarmContent: alarmContent,
		EventType:    eventType,
		EventId:      eventId,
	}

	j, _ := json.Marshal(alarmJson)
	request, err := http.NewRequest(webhook.HttpMethod, webhook.Url, bytes.NewReader([]byte(j)))
	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("User-Agent", "Grafana")
	if webhook.User != "" && webhook.Password != "" {
		request.Header.Add("Authorization", util.GetBasicAuthHeader(webhook.User, webhook.Password))
	}

	for k, v := range webhook.HttpHeader {
		request.Header.Set(k, v)
	}

	resp, err := ctxhttp.Do(ctx, netClient, request)
	if err != nil {
		return err
	}

	if resp.StatusCode/100 == 2 {
		return nil
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	webhookLog.Debug("Webhook failed", "statuscode", resp.Status, "body", string(body))
	return fmt.Errorf("Webhook response status %v", resp.Status)
}

var addToWebhookQueue = func(msg *Webhook) {
	webhookQueue <- msg
}
