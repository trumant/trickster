/*
 * Copyright 2018 The Trickster Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package irondb

import (
	"net/http"
	"time"

	"github.com/trickstercache/trickster/v2/pkg/backends/irondb/common"
	"github.com/trickstercache/trickster/v2/pkg/proxy/engines"
	"github.com/trickstercache/trickster/v2/pkg/proxy/errors"
	"github.com/trickstercache/trickster/v2/pkg/proxy/params"
	"github.com/trickstercache/trickster/v2/pkg/proxy/urls"
	"github.com/trickstercache/trickster/v2/pkg/timeseries"
)

// RollupHandler handles requests for numeric timeseries data with specified
// spans and processes them through the delta proxy cache.
func (c *Client) RollupHandler(w http.ResponseWriter, r *http.Request) {
	r.URL = urls.BuildUpstreamURL(r, c.BaseUpstreamURL())
	engines.DeltaProxyCacheRequest(w, r, c.Modeler())
}

// rollupHandlerSetExtent will change the upstream request query to use the
// provided Extent.
func (c *Client) rollupHandlerSetExtent(r *http.Request,
	trq *timeseries.TimeRangeQuery,
	extent *timeseries.Extent) {
	ok, st, et := c.resolveTRQWithExtent(r, trq, extent)
	if !ok {
		return
	}
	q := r.URL.Query()
	q.Set(upStart, common.FormatTimestamp(time.Unix(0, st), true))
	q.Set(upEnd, common.FormatTimestamp(time.Unix(0, et), true))
	r.URL.RawQuery = q.Encode()
}

// rollupHandlerParseTimeRangeQuery parses the key parts of a TimeRangeQuery
// from the inbound HTTP Request.
func (c *Client) rollupHandlerParseTimeRangeQuery(
	r *http.Request) (*timeseries.TimeRangeQuery, error) {
	trq := &timeseries.TimeRangeQuery{}
	trq.Statement = r.URL.Path

	qp := r.URL.Query()
	var err error
	var p string
	if p = qp.Get(upStart); p == "" {
		return nil, errors.MissingURLParam(upStart)
	}

	if trq.Extent.Start, err = common.ParseTimestamp(p); err != nil {
		return nil, err
	}

	if p = qp.Get(upEnd); p == "" {
		return nil, errors.MissingURLParam(upEnd)
	}

	if trq.Extent.End, err = common.ParseTimestamp(p); err != nil {
		return nil, err
	}

	if p = qp.Get(upSpan); p == "" {
		return nil, errors.MissingURLParam(upSpan)
	}

	if trq.Step, err = common.ParseDuration(p); err != nil {
		return nil, err
	}

	return trq, nil
}

// rollupHandlerFastForwardURL returns the url to fetch the Fast Forward value
// based on a timerange URL.
func (c *Client) rollupHandlerFastForwardRequest(
	r *http.Request) (*http.Request, error) {
	nr, start, end, v, err := c.resolveTRQ(r)
	if err != nil {
		return nil, err
	}
	v.Set(upStart, common.FormatTimestamp(time.Unix(start, 0), true))
	v.Set(upEnd, common.FormatTimestamp(time.Unix(end, 0), true))
	params.SetRequestValues(nr, v)
	return nr, nil
}
