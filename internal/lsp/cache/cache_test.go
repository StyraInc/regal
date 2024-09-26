package cache

import (
	"reflect"
	"testing"

	"github.com/styrainc/regal/pkg/report"
)

func TestManageAggregates(t *testing.T) {
	t.Parallel()

	reportAggregatesFile1 := map[string][]report.Aggregate{
		"my-rule-name": {
			{
				"aggregate_data": map[string]any{
					"foo": "bar",
				},
				"aggregate_source": map[string]any{
					"file":         "file1.rego",
					"package_path": []string{"p"},
				},
				"rule": map[string]any{
					"category": "my-rule-category",
					"title":    "my-rule-name",
				},
			},
			{
				"aggregate_data": map[string]any{
					"more": "things",
				},
				"aggregate_source": map[string]any{
					"file":         "file1.rego",
					"package_path": []string{"p"},
				},
				"rule": map[string]any{
					"category": "my-rule-category",
					"title":    "my-rule-name",
				},
			},
		},
	}

	reportAggregatesFile2 := map[string][]report.Aggregate{
		"my-rule-name": {
			{
				"aggregate_data": map[string]any{
					"foo": "baz",
				},
				"aggregate_source": map[string]any{
					"file":         "file2.rego",
					"package_path": []string{"p"},
				},
				"rule": map[string]any{
					"category": "my-rule-category",
					"title":    "my-rule-name",
				},
			},
		},
		"my-other-rule-name": {
			{
				"aggregate_data": map[string]any{
					"foo": "bax",
				},
				"aggregate_source": map[string]any{
					"file":         "file2.rego",
					"package_path": []string{"p"},
				},
				"rule": map[string]any{
					"category": "my-other-rule-category",
					"title":    "my-other-rule-name",
				},
			},
		},
	}

	c := NewCache()

	c.SetFileAggregates("file1.rego", reportAggregatesFile1)
	c.SetFileAggregates("file2.rego", reportAggregatesFile2)

	aggs1 := c.GetFileAggregates("file1.rego")
	if len(aggs1) != 1 { // there is one cat/rule for file1
		t.Fatalf("unexpected number of aggregates for file1.rego: %d", len(aggs1))
	}

	aggs2 := c.GetFileAggregates("file2.rego")
	if len(aggs2) != 2 {
		t.Fatalf("unexpected number of aggregates for file2.rego: %d", len(aggs2))
	}

	file1ComplimentAggs := c.GetFileComplimentAggregates("file1.rego")

	if !reflect.DeepEqual(file1ComplimentAggs, aggs2) {
		t.Fatalf("unexpected compliment aggregates for file1.rego, exp\n%v\ngot\n%v", aggs2, file1ComplimentAggs)
	}

	allAggs := c.GetFileAggregates()

	if len(allAggs) != 2 {
		t.Fatalf("unexpected number of aggregates: %d", len(allAggs))
	}

	if _, ok := allAggs["my-other-rule-category/my-other-rule-name"]; !ok {
		t.Fatalf("missing aggregate my-other-rule-name")
	}

	c.SetAggregates(reportAggregatesFile1) // update aggregates to only contain file1.rego's aggregates

	allAggs = c.GetFileAggregates()

	if len(allAggs) != 1 {
		t.Fatalf("unexpected number of aggregates: %d", len(allAggs))
	}

	if _, ok := allAggs["my-rule-category/my-rule-name"]; !ok {
		t.Fatalf("missing aggregate my-rule-name")
	}

	// remove file1 from the cache
	c.Delete("file1.rego")

	allAggs = c.GetFileAggregates()

	if len(allAggs) != 0 {
		t.Fatalf("unexpected number of aggregates: %d", len(allAggs))
	}
}
