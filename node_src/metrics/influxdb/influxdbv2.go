//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
// Enhanced blockchain implementation by Circle Layer <https://circlelayer.com>
package influxdb

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type v2Reporter struct {
	reg      metrics.Registry
	interval time.Duration

	endpoint     string
	token        string
	bucket       string
	organization string
	namespace    string
	tags         map[string]string

	client influxdb2.Client
	write  api.WriteAPI

	cache map[string]int64
}

// InfluxDBWithTags starts a InfluxDB reporter which will post the from the given metrics.Registry at each d interval with the specified tags
func InfluxDBV2WithTags(r metrics.Registry, d time.Duration, endpoint string, token string, bucket string, organization string, namespace string, tags map[string]string) {
	rep := &v2Reporter{
		reg:          r,
		interval:     d,
		endpoint:     endpoint,
		token:        token,
		bucket:       bucket,
		organization: organization,
		namespace:    namespace,
		tags:         tags,
		cache:        make(map[string]int64),
	}

	rep.client = influxdb2.NewClient(rep.endpoint, rep.token)
	defer rep.client.Close()

	// async write client
	rep.write = rep.client.WriteAPI(rep.organization, rep.bucket)
	errorsCh := rep.write.Errors()

	// have to handle write errors in a separate goroutine like this b/c the channel is unbuffered and will block writes if not read
	go func() {
		for err := range errorsCh {
			log.Warn("write error", "err", err.Error())
		}
	}()
	rep.run()
}

func (r *v2Reporter) run() {
	intervalTicker := time.Tick(r.interval)
	pingTicker := time.Tick(time.Second * 5)

	for {
		select {
		case <-intervalTicker:
			r.send()
		case <-pingTicker:
			_, err := r.client.Health(context.Background())
			if err != nil {
				log.Warn("Got error from influxdb client health check", "err", err.Error())
			}
		}
	}

}

func (r *v2Reporter) send() {
	r.reg.Each(func(name string, i interface{}) {
		now := time.Now()
		namespace := r.namespace

		switch metric := i.(type) {

		case metrics.Counter:
			v := metric.Count()
			l := r.cache[name]

			measurement := fmt.Sprintf("%s%s.count", namespace, name)
			fields := map[string]interface{}{
				"value": v - l,
			}

			pt := influxdb2.NewPoint(measurement, r.tags, fields, now)
			r.write.WritePoint(pt)

			r.cache[name] = v

		case metrics.Gauge:
			ms := metric.Snapshot()

			measurement := fmt.Sprintf("%s%s.gauge", namespace, name)
			fields := map[string]interface{}{
				"value": ms.Value(),
			}

			pt := influxdb2.NewPoint(measurement, r.tags, fields, now)
			r.write.WritePoint(pt)

		case metrics.GaugeFloat64:
			ms := metric.Snapshot()

			measurement := fmt.Sprintf("%s%s.gauge", namespace, name)
			fields := map[string]interface{}{
				"value": ms.Value(),
			}

			pt := influxdb2.NewPoint(measurement, r.tags, fields, now)
			r.write.WritePoint(pt)

		case metrics.Histogram:
			ms := metric.Snapshot()

			if ms.Count() > 0 {
				ps := ms.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999, 0.9999})
				measurement := fmt.Sprintf("%s%s.histogram", namespace, name)
				fields := map[string]interface{}{
					"count":    ms.Count(),
					"max":      ms.Max(),
					"mean":     ms.Mean(),
					"min":      ms.Min(),
					"stddev":   ms.StdDev(),
					"variance": ms.Variance(),
					"p50":      ps[0],
					"p75":      ps[1],
					"p95":      ps[2],
					"p99":      ps[3],
					"p999":     ps[4],
					"p9999":    ps[5],
				}

				pt := influxdb2.NewPoint(measurement, r.tags, fields, now)
				r.write.WritePoint(pt)
			}

		case metrics.Meter:
			ms := metric.Snapshot()

			measurement := fmt.Sprintf("%s%s.meter", namespace, name)
			fields := map[string]interface{}{
				"count": ms.Count(),
				"m1":    ms.Rate1(),
				"m5":    ms.Rate5(),
				"m15":   ms.Rate15(),
				"mean":  ms.RateMean(),
			}

			pt := influxdb2.NewPoint(measurement, r.tags, fields, now)
			r.write.WritePoint(pt)

		case metrics.Timer:
			ms := metric.Snapshot()
			ps := ms.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999, 0.9999})

			measurement := fmt.Sprintf("%s%s.timer", namespace, name)
			fields := map[string]interface{}{
				"count":    ms.Count(),
				"max":      ms.Max(),
				"mean":     ms.Mean(),
				"min":      ms.Min(),
				"stddev":   ms.StdDev(),
				"variance": ms.Variance(),
				"p50":      ps[0],
				"p75":      ps[1],
				"p95":      ps[2],
				"p99":      ps[3],
				"p999":     ps[4],
				"p9999":    ps[5],
				"m1":       ms.Rate1(),
				"m5":       ms.Rate5(),
				"m15":      ms.Rate15(),
				"meanrate": ms.RateMean(),
			}

			pt := influxdb2.NewPoint(measurement, r.tags, fields, now)
			r.write.WritePoint(pt)

		case metrics.ResettingTimer:
			t := metric.Snapshot()

			if len(t.Values()) > 0 {
				ps := t.Percentiles([]float64{50, 95, 99})
				val := t.Values()

				measurement := fmt.Sprintf("%s%s.span", namespace, name)
				fields := map[string]interface{}{
					"count": len(val),
					"max":   val[len(val)-1],
					"mean":  t.Mean(),
					"min":   val[0],
					"p50":   ps[0],
					"p95":   ps[1],
					"p99":   ps[2],
				}

				pt := influxdb2.NewPoint(measurement, r.tags, fields, now)
				r.write.WritePoint(pt)
			}
		}
	})

	// Force all unwritten data to be sent
	r.write.Flush()
}
