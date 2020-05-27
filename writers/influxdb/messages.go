// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package influxdb

import (
	"math"
	"strconv"
	"time"
	"strings"
//	"fmt"

	"github.com/mainflux/mainflux/transformers/senml"
	"github.com/mainflux/mainflux/writers"
	log "github.com/mainflux/mainflux/logger"
	influxdata "github.com/influxdata/influxdb/client/v2"
)

const pointName = "chan_"

var _ writers.MessageRepository = (*influxRepo)(nil)

type influxRepo struct {
	client influxdata.Client
	cfg    influxdata.BatchPointsConfig
	logger      log.Logger
}

type fields map[string]interface{}
type tags map[string]string

// New returns new InfluxDB writer.
func New(client influxdata.Client, database string, logger log.Logger) writers.MessageRepository {
	return &influxRepo{
		client: client,
		cfg: influxdata.BatchPointsConfig{
			Database: database,
		},
		logger:      logger,
	}
}

func (repo *influxRepo) Save(messages ...senml.Message) error {
	pts, err := influxdata.NewBatchPoints(repo.cfg)
	if err != nil {
		return err
	}

	var nowTime = time.Now()

	for _, msg := range messages {
		tgs, flds := repo.tagsOf(&msg), repo.fieldsOf(&msg)

		sec, dec := math.Modf(msg.Time)
		t := time.Unix(int64(sec), int64(dec*(1e9)))
		if sec == 0 && dec == 0 {
			t = nowTime
                }

		var customPointName = strings.Replace(pointName+msg.Channel, "-", "_", -1)

		//repo.logger.Warn(fmt.Sprintf("Selecting point name: %s, from channel: %s, @%s, ", customPointName, msg.Channel, t.String()))

		pt, err := influxdata.NewPoint(customPointName, tgs, flds, t)
		if err != nil {
			return err
		}
		pts.AddPoint(pt)
	}

	return repo.client.Write(pts)
}

func (repo *influxRepo) tagsOf(msg *senml.Message) tags {
	return tags{
		"subtopic":  msg.Subtopic,
		"publisher": msg.Publisher,
		"name":      msg.Name,
	}
}

func (repo *influxRepo) fieldsOf(msg *senml.Message) fields {
	updateTime := strconv.FormatFloat(msg.UpdateTime, 'f', -1, 64)
	ret := fields{
		"protocol":   msg.Protocol,
		"unit":       msg.Unit,
		"updateTime": updateTime,
	}

	switch {
	case msg.Value != nil:
		ret["value"] = *msg.Value
	case msg.StringValue != nil:
		ret["stringValue"] = *msg.StringValue
	case msg.DataValue != nil:
		ret["dataValue"] = *msg.DataValue
	case msg.BoolValue != nil:
		ret["boolValue"] = *msg.BoolValue
	}

	if msg.Sum != nil {
		ret["sum"] = *msg.Sum
	}

	return ret
}
