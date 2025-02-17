package cgroups_topics_page

import (
	"github.com/stretchr/testify/assert"
	"ktea/kadmin"
	"ktea/ui"
	"testing"
)

func TestCgroupPartsOffsetsPage(t *testing.T) {

	t.Run("List consumer groups", func(t *testing.T) {
		model, _ := New(kadmin.NewMockKadmin(), "test-group")

		model.Update(kadmin.OffsetListedMsg{
			Offsets: []kadmin.TopicPartitionOffset{
				{
					Topic:     "topic-1",
					Partition: 0,
					Offset:    10,
				},
				{
					Topic:     "topic-1",
					Partition: 1,
					Offset:    11,
				},
				{
					Topic:     "topic-2",
					Partition: 0,
					Offset:    20,
				},
				{
					Topic:     "topic-2",
					Partition: 1,
					Offset:    21,
				},
			},
		})

		view := model.View(ui.NewTestKontext(), ui.TestRenderer)

		assert.Contains(t, view, "topic-1")
		assert.Contains(t, view, "topic-2")
		assert.Contains(t, view, "10")
		assert.Contains(t, view, "11")
		assert.NotContains(t, view, "20")
		assert.NotContains(t, view, "21")
	})

	t.Run("Render empty page when no offsets found", func(t *testing.T) {
		model, _ := New(kadmin.NewMockKadmin(), "test-group")

		model.Update(kadmin.OffsetListedMsg{
			Offsets: nil,
		})

		view := model.View(ui.NewTestKontext(), ui.TestRenderer)

		assert.Contains(t, view, "👀 No Committed Offsets Found")
	})

}
