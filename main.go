package main

import (
	mkr "github.com/mackerelio/mackerel-client-go"
	mp "github.com/mackerelio/go-mackerel-plugin-helper"
	"fmt"
	"flag"
	"strings"
	"math"
	"os"
)

func main() {
	optNodeName := flag.String("node", "", "Mackerel Node Name (default: use hostname)")
	optService := flag.String("service", "", "Service Name")
	optRole := flag.String("role", "", "Role Name")
	optMetricsName := flag.String("metric-name", "", "Metric Name")
	optCliMode := flag.Bool("cli-mode", false, "CLI Mode")

	flag.Parse()

	if *optNodeName == "" {
		*optNodeName = getHostname()
	}

	var ss = StandardScorePlugin{
		Prefix: "standard_score",
		NodeName: *optNodeName,
		Service: *optService,
		MetricName: *optMetricsName,
		Role: strings.Split(*optRole, ","),
		MackerelClient: mkr.NewClient(os.Getenv("MACKEREL_APIKEY")),
	}
	if (*optCliMode) {
		r, _ := ss.FetchMetrics()
		fmt.Printf("%v\n", r)
	} else {
		helper := mp.NewMackerelPlugin(ss)
		helper.Run()
	}
}


type StandardScorePlugin struct {
	Prefix string
	NodeName string
	Service string
	Role []string
	MetricName string
	MackerelClient *mkr.Client
}

func (u StandardScorePlugin) FetchMetrics() (map[string]interface{}, error) {
	ss, err := u.GetStandardScore(u.MetricName)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch " + u.MetricName + " standard score: %s", err)
	}
	return map[string]interface{}{"standard_score": ss}, nil
}

func (u StandardScorePlugin) GraphDefinition() map[string](mp.Graphs) {
	labelPrefix := strings.Title(u.Prefix + "." + u.MetricName)
	return map[string](mp.Graphs) {
		u.Prefix + "." + u.MetricName: mp.Graphs{
			Label: labelPrefix,
			Unit: "float",
			Metrics: [](mp.Metrics) {
				mp.Metrics{
					Name: "standard_score",
					Label: "StandardScore",
				},
			},
		},
	}
}

func (u StandardScorePlugin) GetStandardScore(metricsName string) (float64, error) {
	memoryUsedValues := []float64{}

	hosts, err := u.FetchHosts()
	if err != nil {
		fmt.Println(err)
	}
	var metricValues = mkr.LatestMetricValues{}
	metricValues, err = u.FetchLatestMetricValues(hosts, []string{metricsName})
	if err != nil {
		fmt.Println(err)
	}
	for _, metricValue := range metricValues {
		if metricValue[metricsName] != nil {
			memoryUsedValues = append(memoryUsedValues, metricValue[metricsName].Value.(float64))
		}
	}
	av, err := average(memoryUsedValues)
	if err != nil {
		fmt.Println(err)
	}
	sd, err := standardDeviation(memoryUsedValues, av)
	if err != nil {
		fmt.Println(err)
	}
	NodeID := u.nodeIDByName(hosts, u.NodeName)
	ss, _ := standardScore(metricValues[NodeID][metricsName].Value.(float64), av, sd)
	return ss, err
}

func (u StandardScorePlugin) FetchHosts() ([]*mkr.Host, error){
	return u.MackerelClient.FindHosts(&mkr.FindHostsParam{
		Service: u.Service,
		Roles: u.Role,
		Statuses: []string{"working"},
	})
}

func (u StandardScorePlugin) FetchLatestMetricValues(hosts []*mkr.Host, metricNames []string) (mkr.LatestMetricValues, error) {
	var err error
	LatestMetricValues := mkr.LatestMetricValues{}
	for _, hostChunks := range eachSlice(hosts, 50) {
		var hostIDs = u.hostIDs(hostChunks)
		metricValues, err := u.MackerelClient.FetchLatestMetricValues(hostIDs, metricNames)
		if err != nil {
			fmt.Println(err)
		}
		for key, metricValue := range metricValues {
			LatestMetricValues[key] = metricValue
		}
	}
	return LatestMetricValues, err
}

func (u StandardScorePlugin) hostIDs(hosts []*mkr.Host) (hostIDs []string) {
	for _, host := range hosts {
		hostIDs = append(hostIDs, host.ID)
	}
	return hostIDs
}

func (u StandardScorePlugin) nodeIDByName(hosts []*mkr.Host, nodeName string) (nodeID string) {
	for _, host := range hosts {
		if (host.Name == u.NodeName) {
			nodeID = host.ID
		}
	}
	return nodeID
}
func average(values []float64) (sum float64, err error) {
	for _, value := range values{
		sum += value
	}
	return sum / float64(len(values)), nil
}

func standardDeviation(values[]float64, average float64) (sd float64, err error) {
	for _, value := range values {
		sd = sd + math.Pow(value - average, 2)
	}
	sd = math.Sqrt(sd / float64(len(values)))
	return sd, nil
}

func standardScore(value float64, average float64, standardDeviation float64) (float64, error) {
	ss := (value - average) / standardDeviation * 10 + 50
	return ss, nil
}


func eachSlice(slice []*mkr.Host, size int) [][]*mkr.Host {
	var chunks [][]*mkr.Host

	sliceSize := len(slice)

	for i := 0; i < sliceSize; i += size {
		end := i + size
		if sliceSize < end {
			end = sliceSize
		}
		chunks = append(chunks, slice[i:end])

	}
	return chunks
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Errorf("Failed to get hostname: ", err)
	}
	return hostname
}
