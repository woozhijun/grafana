package notifiers

import (
	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/log"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/alerting"
	"io/ioutil"
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"strconv"
	"github.com/grafana/grafana/pkg/cmd/grafana-cli/logger"
)

func init() {
  alerting.RegisterNotifier(&alerting.NotifierPlugin{
    Type:        "argusAlarm",
    Name:        "ArgusAlarm",
    Description: "Sends HTTP POST request to argus alarm.",
    Factory:     NewArgusAlarmNotifier,
    OptionsTemplate: `
      <h3 class="page-heading">ArgusAlarm settings</h3>
      <div class="gf-form">
        <span class="gf-form-label width-10">Url</span>
        <input type="text" required class="gf-form-input max-width-26" ng-model="ctrl.model.settings.url"></input>
      </div>
      <div class="gf-form">
        <span class="gf-form-label width-10">Http Method</span>
        <div class="gf-form-select-wrapper width-14">
          <select class="gf-form-input" ng-model="ctrl.model.settings.httpMethod" ng-options="t for t in ['POST', 'PUT']">
          </select>
        </div>
      </div>
      <div class="gf-form">
        <span class="gf-form-label width-10">alarmGroup</span>
        <div class="gf-form-select-wrapper width-14">
          <select class="gf-form-input" ng-model="ctrl.model.settings.alarmGroup" ng-options="alarmGroup for alarmGroup in ctrl.alarmGroups" ng-change="ctrl.getAlarmGroups(notification, $index)">
          </select>
        </div>
      </div>
      <h3 class="page-heading">Quiet hours</h3>
      <div class="gf-form">
        <span class="gf-form-label width-10">startHour</span>
        <div class="gf-form-select-wrapper width-14">
            <select class="gf-form-input max-width-14" ng-model="ctrl.model.settings.startHour" ng-options="s for s in [
              '---请选择---','00:00','01:00','02:00','03:00','04:00','05:00','06:00','07:00','08:00','09:00','10:00','11:00','12:00',
              '13:00','14:00','15:00','16:00','17:00','18:00','19:00','20:00','21:00','22:00','23:00',
              ]"></select>
          </select>
        </div>
      </div>
      <div class="gf-form">
        <span class="gf-form-label width-10">endHour</span>
        <div class="gf-form-select-wrapper width-14">
          <select class="gf-form-input max-width-14" ng-model="ctrl.model.settings.endHour" ng-options="s for s in [
              '---请选择---','00:00','01:00','02:00','03:00','04:00','05:00','06:00','07:00','08:00','09:00','10:00','11:00','12:00',
              '13:00','14:00','15:00','16:00','17:00','18:00','19:00','20:00','21:00','22:00','23:00',
              ]"></select>
            </select>
          </div>
      </div>
    `,
  })

}

func NewArgusAlarmNotifier(model *m.AlertNotification) (alerting.Notifier, error) {
  url := model.Settings.Get("url").MustString()
  if url == "" {
    return nil, alerting.ValidationError{Reason: "Could not find url property in settings"}
  }

  return &ArgusAlarmNotifier{
      NotifierBase: NewNotifierBase(model.Id, model.IsDefault, model.Name, model.Type, model.Settings),
      Url:          url,
	  HttpMethod:   model.Settings.Get("httpMethod").MustString("POST"),
      alarmGroup:   model.Settings.Get("alarmGroup").MustString(),
      startHour:    model.Settings.Get("startHour").MustString(),
      endHour:      model.Settings.Get("endHour").MustString(),
      log:          log.New("alerting.notifier.argusAlarm."),
    }, nil
}

type ArgusAlarmNotifier struct {
    NotifierBase
    Url        string
	HttpMethod string
	alarmGroup string
    startHour  string
    endHour    string
    log        log.Logger
}

func (this *ArgusAlarmNotifier) Notify(evalContext *alerting.EvalContext) error {

	this.log.Info("Sending argusAlarm...")
	var status int
	var alarmGroup string
	var alarmContent string
	state := evalContext.Rule.State
	//condition := evalContext.Rule.Conditions
	//for cond := range condition {
	//
	//	data, err := json.Marshal(cond)
	//	if err != nil {
	//		this.log.Info("解析错误" + err.Error())
	//	}
	//	this.log.Info(">>>.condition:" + string(data))
	//}
	bodyJSON := simplejson.New()
	bodyJSON.Set("source", "grafana")
	this.log.Info("Sending Rule state is " + string(state))
	if state == "ok" {
		status = 1
		alarmContent = evalContext.Rule.Message + "(" + evalContext.Rule.Name + ") is ok."
	} else if state == "alerting" {
		status = 0
		alarmContent = evalContext.Rule.Message + "(" + evalContext.Rule.Name + ") is alerting."
	} else {
		this.log.Warn(">>>.other state:" + string(state))
		return nil
	}

	bodyJSON.Set("status", status)
	bodyJSON.Set("message", alarmContent)
	bodyJSON.Set("datetime", time.Now().Unix() * 1000)
	if this.alarmGroup == "---请选择---" {
		alarmGroup = ""
	} else {
		alarmGroup = this.alarmGroup
	}
	bodyJSON.Set("alarmGroup", alarmGroup)

	graphId := strconv.FormatInt(evalContext.Rule.DashboardId, 10) + "_" + strconv.FormatInt(evalContext.Rule.PanelId, 10) + "_" +
		strconv.FormatInt(evalContext.Rule.Id, 10)

	if state == "ok" {
		body, _ := bodyJSON.MarshalJSON()
		this.log.Info("Ok params：" + string(body))
		sendArgusAlarm(this, evalContext, string(body))
	} else {
		for _, eval := range evalContext.EvalMatches {

			bodyJSON.Set("metric", eval.Metric)
			tags := eval.Tags
			if  tags == nil {
				tags = make(map[string]string)
			}
			tags["graphId"] = graphId
			bodyJSON.Set("tags", tags)
			bodyJSON.Set("value", eval.Value)
			body, _ := bodyJSON.MarshalJSON()

			this.log.Info("params：" + string(body))
			sendArgusAlarm(this, evalContext, string(body))
		}
	}
	return nil
}

func sendArgusAlarm(this *ArgusAlarmNotifier, evalContext *alerting.EvalContext, params string) error {

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	cmd := &m.SendWebhookSync{
		Url:        this.Url,
		Body:       params,
		HttpMethod: this.HttpMethod,
		HttpHeader: header,
	}
	if !validQuietTime(this.startHour, this.endHour) {
		if err := bus.DispatchCtx(evalContext.Ctx, cmd); err != nil {
			this.log.Error("Failed to send argusAlarm", "error", err, "argusAlarm", this.Name)
			return err
		}
	} else {
		this.log.Info("The alarm is quiet time.---" + this.startHour + "~" + this.endHour)
	}
	return nil
}

func httpGetAlarmGroups() string {

	resp, err := http.Get("http://dispatcher.monitor.mobike.io/api/alarmGroup/allAlarmGroup")
	if err != nil {
		logger.Error("err-httpGet: " + err.Error())
		return ""
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("err-ioutilReadAll: " + err.Error())
		return ""
	}

	alarmGroup := &AlarmGroupResult{}
	error := json.Unmarshal(body, &alarmGroup)
	if error != nil {
		logger.Error("err-Unmarshal: " + error.Error())
		return ""
	}

	result, err := json.Marshal(append([]string{"---请选择---"}, alarmGroup.Data...))
	if err != nil {
		logger.Error("err-Marshal: " + error.Error())
		return ""
	}
	return strings.Replace(string(result),"\"", "'", -1)
}

type AlarmGroupResult struct {
	Code    int16
	Message string
	Data    []string
}
