package main

import (
  "flag"
  "fmt"
  "golang.org/x/net/html"
  "net/http"
  "os"
  "regexp"
  "strconv"
  "strings"
  "sync"
  "time"

  "github.com/prometheus/client_golang/prometheus"
  "github.com/prometheus/common/log"
  "github.com/prometheus/common/version"
)

const (
  namespace = "surfboard"
)

var (
  listenAddress = flag.String(
    "web.listen-address", ":9239",
    "Address to listen on for web interface and telemetry.",
  )
  metricsPath = flag.String(
    "web.telemetry-path", "/metrics",
    "Path under which to expose metrics.",
  )
  modemAddress = flag.String(
    "modem-address", "192.168.100.1",
    "IP address of Surboard modem.",
  )
  showVersion = flag.Bool(
    "version", false,
    "Print version information.",
  )
  timeout = flag.Duration(
    "timeout", 2*time.Second,
    "Timeout for trying to get stats from Surfboard.",
  )
)

// Exporter collects Surfboard metrics. It implements prometheus.Collector.
type Exporter struct {
  mutex                   sync.Mutex
  client                  *http.Client

  up                      *prometheus.Desc
  downFrequency           *prometheus.Desc
  downPower               *prometheus.Desc
  downSnr                 *prometheus.Desc
  downCodesCorrected      *prometheus.Desc
  downCodesUncorrectable  *prometheus.Desc
  upFrequency             *prometheus.Desc
  upPower                 *prometheus.Desc
}

// NewExporter returns an initialized exporter.
func NewExporter(timeout time.Duration) *Exporter {
  return &Exporter{
    up: prometheus.NewDesc(
      prometheus.BuildFQName(namespace, "", "up"),
      "Could the surfboard be reached.",
      nil,
      nil,
    ),
    downFrequency: prometheus.NewDesc(
      prometheus.BuildFQName(namespace, "downstream", "frequency_hertz"),
      "Downstream frequency in Hertz",
      []string{"channel"},
      nil,
    ),
    downPower: prometheus.NewDesc(
      prometheus.BuildFQName(namespace, "downstream", "power_dbmv"),
      "Downstream power level in dBmv",
      []string{"channel"},
      nil,
    ),
    downSnr: prometheus.NewDesc(
      prometheus.BuildFQName(namespace, "downstream", "snr_db"),
      "Downstream signal to noise ration in dB",
      []string{"channel"},
      nil,
    ),
    downCodesCorrected: prometheus.NewDesc(
      prometheus.BuildFQName(namespace, "downstream", "codewords_corrected_total"),
      "Downstream codewords corrected",
      []string{"channel"},
      nil,
    ),
    downCodesUncorrectable: prometheus.NewDesc(
      prometheus.BuildFQName(namespace, "downstream", "codewords_uncorrectable_total"),
      "Downstream codewords uncorrectable",
      []string{"channel"},
      nil,
    ),
    upFrequency: prometheus.NewDesc(
      prometheus.BuildFQName(namespace, "upstream", "frequency_hertz"),
      "Upstream frequency in Hertz",
      []string{"channel"},
      nil,
    ),
    upPower: prometheus.NewDesc(
      prometheus.BuildFQName(namespace, "upstream", "power_dbmv"),
      "Upstream power level in dBmv",
      []string{"channel"},
      nil,
    ),
    client: &http.Client{
      Timeout: timeout,
    },
  }
}

// Describe describes all the metrics exported by the surfboard exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
  ch <- e.up
  ch <- e.downFrequency
  ch <- e.downPower
  ch <- e.downSnr
  ch <- e.downCodesCorrected
  ch <- e.downCodesUncorrectable
  ch <- e.upFrequency
  ch <- e.upPower
}

// Collect fetches the statistics from the configured surfboard modem, and
// delivers them as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
  e.mutex.Lock() // To protect metrics from concurrent collects.
  defer e.mutex.Unlock()

  resp, err := e.client.Get(fmt.Sprintf("http://%s/cgi-bin/status", *modemAddress))
  if err != nil {
    ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
    log.Errorf("Failed to collect stats from surfboard: %s", err)
    return
  }
  ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1)

  body := resp.Body
  defer body.Close()
  z := html.NewTokenizer(body)

  var (
    content [3][32][]string
    state int
    count int
    columns int
  )

  L:
  for {
    tt := z.Next()

    switch {
    case tt == html.ErrorToken:
      // End of the document, we're done
      break L;
    case tt == html.StartTagToken:
      t := z.Token()
      if t.Data == "th" {
        for _, a := range t.Attr {
          if a.Key == "colspan" {
            switch {
            case a.Val == "9":
              state = 1
              count = 0
              columns = 9
            case a.Val == "7":
              state = 2
              count = 0
              columns = 7
            }
          }
        }
      }
      if t.Data == "td" {
        inner := z.Next()
        if inner == html.TextToken {
          text := (string)(z.Text())
          t := strings.TrimSpace(text)
          if state == 1 || state == 2 {
            i := count % columns
            content[state][i] = append(content[state][i], t)
            count++
          }
        }
      }
    }
  }

  for table,rows := range content {
    for column,columns := range rows {
      for id,value := range columns {
        id++
        switch {
        case table == 1:
          //Downstream
          switch {
          case column == 4:
            // Frequency
            r, _ := regexp.Compile(`(\d+\.\d+)`)
            s := r.FindString(value)
            v,_ := strconv.ParseFloat(s, 64)
            i := strconv.Itoa(id)
            ch <- prometheus.MustNewConstMetric(e.downFrequency, prometheus.GaugeValue, v * 1000000, i)
          case column == 5:
            // Power
            r, _ := regexp.Compile(`(\d+\.\d+)`)
            s := r.FindString(value)
            v,_ := strconv.ParseFloat(s, 64)
            i := strconv.Itoa(id)
            ch <- prometheus.MustNewConstMetric(e.downPower, prometheus.GaugeValue, v, i)
          case column == 6:
            // SNR
            r, _ := regexp.Compile(`(\d+\.\d+)`)
            s := r.FindString(value)
            v,_ := strconv.ParseFloat(s, 64)
            i := strconv.Itoa(id)
            ch <- prometheus.MustNewConstMetric(e.downSnr, prometheus.GaugeValue, v, i)
          case column == 7:
            // Corrected Codewords
            r, _ := regexp.Compile(`(\d+)`)
            s := r.FindString(value)
            v,_ := strconv.ParseFloat(s, 64)
            i := strconv.Itoa(id)
            ch <- prometheus.MustNewConstMetric(e.downCodesCorrected, prometheus.CounterValue, v, i)
          case column == 8:
            // Uncorrectable Codewords
            r, _ := regexp.Compile(`(\d+)`)
            s := r.FindString(value)
            v,_ := strconv.ParseFloat(s, 64)
            i := strconv.Itoa(id)
            ch <- prometheus.MustNewConstMetric(e.downCodesUncorrectable, prometheus.CounterValue, v, i)
          }
        case table == 2:
          // Upstream
          switch {
          case column == 5:
            // Frequency
            r, _ := regexp.Compile(`(\d+\.\d+)`)
            s := r.FindString(value)
            v,_ := strconv.ParseFloat(s, 64)
            i := strconv.Itoa(id)
            ch <- prometheus.MustNewConstMetric(e.upFrequency, prometheus.GaugeValue, v * 1000000, i)
          case column == 6:
            // Power
            r, _ := regexp.Compile(`(\d+\.\d+)`)
            s := r.FindString(value)
            v,_ := strconv.ParseFloat(s, 64)
            i := strconv.Itoa(id)
            ch <- prometheus.MustNewConstMetric(e.upPower, prometheus.GaugeValue, v, i)
          }
        }
      }
    }
  }

}

func init() {
  prometheus.MustRegister(version.NewCollector("surfboard_exporter"))
}

func main() {
  flag.Parse()

  if *showVersion {
    fmt.Fprintln(os.Stdout, version.Print("surfboard_exporter"))
    os.Exit(0)
  }

  log.Infoln("Starting surfboard exporter", version.Info())
  log.Infoln("Build context", version.BuildContext())

  exporter := NewExporter(*timeout)
  prometheus.MustRegister(exporter)

  http.Handle(*metricsPath, prometheus.Handler())
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`<html>
    <head><title>Surfboard Exporter</title></head>
    <body>
    <h1>Surfboard Exporter</h1>
    <p><a href='` + *metricsPath + `'>Metrics</a></p>
    </body>
    </html>`))
  })
  log.Infof("Listening on %s", *listenAddress)
  log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

