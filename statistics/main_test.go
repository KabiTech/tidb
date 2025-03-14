// Copyright 2021 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package statistics

import (
	"testing"

	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/tidb/sessionctx/stmtctx"
	"github.com/pingcap/tidb/types"
	"github.com/pingcap/tidb/util/testbridge"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	opts := []goleak.Option{
		goleak.IgnoreTopFunction("go.etcd.io/etcd/pkg/logutil.(*MergeLogger).outputLoop"),
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	}
	testbridge.WorkaroundGoCheckFlags()
	goleak.VerifyTestMain(m, opts...)
}

func TestStatistics(t *testing.T) {
	s := createTestStatisticsSuite(t)
	t.Run("TestSketch", SubTestSketch(s))
	t.Run("TestSketchProtoConversion", SubTestSketchProtoConversion(s))
	t.Run("TestFMSketchCoding", SubTestFMSketchCoding(s))
}

func createTestStatisticsSuite(t *testing.T) *testStatisticsSuite {
	s := new(testStatisticsSuite)

	s.count = 100000
	samples := make([]*SampleItem, 10000)
	for i := 0; i < len(samples); i++ {
		samples[i] = &SampleItem{}
	}
	start := 1000
	samples[0].Value.SetInt64(0)
	for i := 1; i < start; i++ {
		samples[i].Value.SetInt64(2)
	}
	for i := start; i < len(samples); i++ {
		samples[i].Value.SetInt64(int64(i))
	}
	for i := start; i < len(samples); i += 3 {
		samples[i].Value.SetInt64(samples[i].Value.GetInt64() + 1)
	}
	for i := start; i < len(samples); i += 5 {
		samples[i].Value.SetInt64(samples[i].Value.GetInt64() + 2)
	}
	sc := new(stmtctx.StatementContext)

	var err error
	s.samples, err = SortSampleItems(sc, samples)
	require.NoError(t, err)

	rc := &recordSet{
		data:   make([]types.Datum, s.count),
		count:  s.count,
		cursor: 0,
	}
	rc.setFields(mysql.TypeLonglong)
	rc.data[0].SetInt64(0)
	for i := 1; i < start; i++ {
		rc.data[i].SetInt64(2)
	}
	for i := start; i < rc.count; i++ {
		rc.data[i].SetInt64(int64(i))
	}
	for i := start; i < rc.count; i += 3 {
		rc.data[i].SetInt64(rc.data[i].GetInt64() + 1)
	}
	for i := start; i < rc.count; i += 5 {
		rc.data[i].SetInt64(rc.data[i].GetInt64() + 2)
	}
	require.NoError(t, types.SortDatums(sc, rc.data))

	s.rc = rc

	pk := &recordSet{
		data:   make([]types.Datum, s.count),
		count:  s.count,
		cursor: 0,
	}
	pk.setFields(mysql.TypeLonglong)
	for i := 0; i < rc.count; i++ {
		pk.data[i].SetInt64(int64(i))
	}
	s.pk = pk

	return s
}
