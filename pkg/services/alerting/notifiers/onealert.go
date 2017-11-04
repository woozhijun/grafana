package notifiers

import (
  "github.com/grafana/grafana/pkg/bus"
  "github.com/grafana/grafana/pkg/components/simplejson"
  "github.com/grafana/grafana/pkg/log"
  "github.com/grafana/grafana/pkg/metrics"
  m "github.com/grafana/grafana/pkg/models"
  "github.com/grafana/grafana/pkg/services/alerting"
  "bytes"
  "strconv"
  "time"
)

func init() {
  alerting.RegisterNotifier(&alerting.NotifierPlugin{
    Type:        "onealert",
    Name:        "OneAlert",
    Description: "Sends HTTP POST request to one alert",
    Factory:     NewOneAlertNotifier,
    OptionsTemplate: `
      <h3 class="page-heading">Onealert settings</h3>
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
        <span class="gf-form-label width-10">Key</span>
        <input type="text" class="gf-form-input max-width-14" ng-model="ctrl.model.settings.key"></input>
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

func NewOneAlertNotifier(model *m.AlertNotification) (alerting.Notifier, error) {
  url := model.Settings.Get("url").MustString()
  if url == "" {
    return nil, alerting.ValidationError{Reason: "Could not find url property in settings"}
  }

  return &OneAlertNotifier{
      NotifierBase: NewNotifierBase(model.Id, model.IsDefault, model.Name, model.Type, model.Settings),
      Url:          url,
      Key:          model.Settings.Get("key").MustString(),
      HttpMethod:   model.Settings.Get("httpMethod").MustString("POST"),
      startHour:    model.Settings.Get("startHour").MustString(),
      endHour:      model.Settings.Get("endHour").MustString(),
      log:          log.New("alerting.notifier.onealert"),
    }, nil
}

type OneAlertNotifier struct {
    NotifierBase
    Url        string
    Key        string
    HttpMethod string
    startHour  string
    endHour    string
    log        log.Logger
}

func (this *OneAlertNotifier) Notify(evalContext *alerting.EvalContext) error {
  this.log.Info("Sending onealert")
  metrics.M_Alerting_Notification_Sent_Onealert.Inc(1)

  var eventType string
  var alarmContent string
  state := evalContext.Rule.State
  if state == "ok" {
    eventType = "resolve"
    alarmContent = "[RESOLVED] " + evalContext.Rule.Message + "is Ok."
  } else if state == "alerting" {
    eventType = "trigger"
    var buf bytes.Buffer
    buf.WriteString("[ALERTING] " + evalContext.Rule.Message)
    for _, eval := range evalContext.EvalMatches {
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

    buf.WriteString(" Trigger policy(" + evalContext.Rule.Name + "), Please check it.")
    alarmContent = buf.String()
  } else {
    return nil
  }

  bodyJSON := simplejson.New()
  bodyJSON.Set("app", this.Key)
  bodyJSON.Set("alertName", evalContext.Rule.Name)
  bodyJSON.Set("alarmContent", alarmContent)
  bodyJSON.Set("eventType", eventType)
  bodyJSON.Set("eventId", strconv.FormatInt(evalContext.Rule.DashboardId, 10) + "_" +
                              strconv.FormatInt(evalContext.Rule.PanelId, 10) + "_" +
                              strconv.FormatInt(evalContext.Rule.Id, 10))
  body, _ := bodyJSON.MarshalJSON()

  cmd := &m.SendWebhookSync{
    Url:        this.Url,
    Body:       string(body),
    HttpMethod: this.HttpMethod,
  }

  if !validQuietTime(this.startHour, this.endHour) {
    if err := bus.DispatchCtx(evalContext.Ctx, cmd); err != nil {
      this.log.Error("Failed to send onealert", "error", err, "onealert", this.Name)
      return err
    }
  } else {
    this.log.Info("The alarm is quiet time.---", alarmContent)
  }
  return nil
}

func validQuietTime(start string, end string) bool {

  if start == "" || end == "" || start == "---请选择---" || end == "---请选择---" {
    return false
  }
  now := time.Now()
  today := now.Format("2006-01-02")
  startTime, err := time.Parse("2006-01-02 15:04", today + " " + start)
  endTime, err1 := time.Parse("2006-01-02 15:04", today + " " + end)
  current, err1 := time.Parse("2006-01-02 15:04", now.Format("2006-01-02 15:04"))

  if err != nil || err1 != nil {
    return false
  }

  if startTime.Hour() > endTime.Hour() {
    endTime = endTime.Add(24 * 60 * 60 * 1000 * 1000 * 1000)
  }

  if current.After(startTime) && current.Before(endTime) {
    return true
  }
  return false
}
