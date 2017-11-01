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
      log:          log.New("alerting.notifier.onealert"),
    }, nil
}

type OneAlertNotifier struct {
    NotifierBase
    Url        string
    Key        string
    HttpMethod string
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

  if err := bus.DispatchCtx(evalContext.Ctx, cmd); err != nil {
    this.log.Error("Failed to send onealert", "error", err, "onealert", this.Name)
    return err
  }

  return nil
}
