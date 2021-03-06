package kafkactl

import (
	"sort"

	"github.com/Shopify/sarama"
	"github.com/spf13/cast"
)

type TopicSummary struct {
	Topic           string
	Parts           string
	RFactor         int
	ISRs            int
	OfflineReplicas int
	Partitions      []int32
}

type TopicMeta struct {
	Topic           string
	Partition       int32
	Leader          int32
	Replicas        []int32
	ISRs            []int32
	OfflineReplicas []int32
}

type TopicOffsetMap struct {
	Topic            string
	TopicMeta        []TopicMeta
	PartitionOffsets map[int32]int64
	PartitionLeaders map[int32]int32
}

// TopicOffsetGet is WiP* for a TopicMeta struct to get the current, newest topic offset for its' partition.
type TopicOffsetGet interface {
	GetPartitionOffset(client KClient, topic string, partition int32)
}

func (kc *KClient) MakeTopicOffsetMap(topicMeta []TopicMeta) []TopicOffsetMap {
	var TOM []TopicOffsetMap
	parts := make(map[string][]int32)
	tmMap := make(map[string][]TopicMeta)
	for _, tm := range topicMeta {
		parts[tm.Topic] = append(parts[tm.Topic], tm.Partition)
		tmMap[tm.Topic] = append(tmMap[tm.Topic], tm)
	}
	for topic := range tmMap {
		poMap := make(map[int32]int64)
		for _, p := range parts[topic] {
			off, err := kc.GetOffsetNewest(topic, p)
			if err != nil {
				off = -7777
			}
			poMap[p] = off
		}
		pLdrMap := make(map[int32]int32, len(tmMap[topic]))
		for _, tm := range tmMap[topic] {
			pLdrMap[tm.Partition] = tm.Leader
		}
		tom := TopicOffsetMap{
			Topic:            topic,
			TopicMeta:        tmMap[topic],
			PartitionOffsets: poMap,
			PartitionLeaders: pLdrMap,
		}
		TOM = append(TOM, tom)
	}
	return TOM
}

func GetTopicSummaries(topicMeta []TopicMeta) []TopicSummary {
	var topicSummary []TopicSummary
	parts := make(map[string][]int32, len(topicMeta))
	isrs := make(map[string][]int32, len(topicMeta))
	reps := make(map[string][]int32, len(topicMeta))
	off := make(map[string][]int32, len(topicMeta))
	done := make(map[string]bool)
	for _, tm := range topicMeta {
		parts[tm.Topic] = append(parts[tm.Topic], tm.Partition)
		isrs[tm.Topic] = append(isrs[tm.Topic], tm.ISRs...)
		reps[tm.Topic] = append(reps[tm.Topic], tm.Replicas...)
		off[tm.Topic] = append(off[tm.Topic], tm.OfflineReplicas...)
	}
	for _, tm := range topicMeta {
		if !done[tm.Topic] {
			done[tm.Topic] = true
			partitions := makeSeqStr(parts[tm.Topic])
			ts := TopicSummary{
				Topic:           tm.Topic,
				Parts:           partitions,
				RFactor:         len(reps[tm.Topic]) / len(parts[tm.Topic]),
				ISRs:            len(isrs[tm.Topic]),
				OfflineReplicas: len(off[tm.Topic]),
				Partitions:      parts[tm.Topic],
			}
			topicSummary = append(topicSummary, ts)
		}
	}
	return topicSummary
}

func (kc *KClient) GetTopicMeta() ([]TopicMeta, error) {
	var topicMeta []TopicMeta
	res, err := kc.ReqMetadata()
	if err != nil {
		return topicMeta, err
	}
	for _, t := range res.Topics {
		topicName := t.Name
		for _, x := range t.Partitions {
			tm := TopicMeta{
				Topic:           topicName,
				Partition:       x.ID,
				Leader:          x.Leader,
				Replicas:        x.Replicas,
				ISRs:            x.Isr,
				OfflineReplicas: x.OfflineReplicas,
			}
			topicMeta = append(topicMeta, tm)
		}
	}
	return topicMeta, err
}

func (kc *KClient) ListTopics() ([]string, error) {
	res, err := kc.ReqMetadata()
	if err != nil {
		return nil, err
	}
	var topics = make([]string, 0, len(res.Topics))
	for _, t := range res.Topics {
		topics = append(topics, t.Name)
	}
	return topics, nil
}

func makeSeqStr(nums []int32) string {
	seqMap := make(map[int][]int32)
	sort.Slice(nums, func(i, j int) bool {
		return nums[i] < nums[j]
	})
	var mapCount int
	var done int
	var switchInt int
	seqMap[mapCount] = append(seqMap[mapCount], nums[done])
	done++
	switchInt = done
	for done < len(nums) {
		if nums[done] == ((seqMap[mapCount][(switchInt - 1)]) + 1) {
			seqMap[mapCount] = append(seqMap[mapCount], nums[done])
			switchInt++
		} else {
			mapCount++
			seqMap[mapCount] = append(seqMap[mapCount], nums[done])
			switchInt = 1
		}
		done++
	}
	var seqStr string
	for k, v := range seqMap {
		if k > 0 {
			seqStr += ","
		}
		if len(v) > 1 {
			seqStr += cast.ToString(v[0])
			seqStr += "-"
			seqStr += cast.ToString(v[len(v)-1])
		} else {
			seqStr += cast.ToString(v[0])
		}
	}
	return seqStr
}

func (kc *KClient) GetOffsetNewest(topic string, partition int32) (int64, error) {
	return kc.cl.GetOffset(topic, partition, sarama.OffsetNewest)
}

func (kc *KClient) GetOffsetOldest(topic string, partition int32) (int64, error) {
	return kc.cl.GetOffset(topic, partition, sarama.OffsetOldest)
}
