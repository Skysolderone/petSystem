package ai

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"petverse/server/internal/model"
)

type HealthSummary struct {
	Score              int       `json:"score"`
	Status             string    `json:"status"`
	Insights           []string  `json:"insights"`
	RecommendedActions []string  `json:"recommended_actions"`
	DataPointsAnalyzed int       `json:"data_points_analyzed"`
	GeneratedAt        time.Time `json:"generated_at"`
}

type HealthAI struct{}

func NewHealthAI() *HealthAI {
	return &HealthAI{}
}

func (h *HealthAI) Summarize(pet model.Pet, records []model.HealthRecord, alerts []model.HealthAlert, dataPoints []model.DeviceDataPoint) HealthSummary {
	score := 92
	insights := []string{}
	actions := []string{}

	activeAlerts := 0
	for _, alert := range alerts {
		if alert.IsDismissed {
			continue
		}
		activeAlerts++
		switch alert.Severity {
		case "critical":
			score -= 18
		case "high":
			score -= 12
		case "medium":
			score -= 7
		default:
			score -= 3
		}
		insights = append(insights, fmt.Sprintf("%s：%s", alert.Title, alert.Message))
	}

	if len(records) == 0 {
		score -= 6
		insights = append(insights, "近期没有新增健康记录，健康档案连续性不足。")
		actions = append(actions, "补充最近的体重、用药或体检记录。")
	}

	var latestWeight float64
	var hasWeight bool
	for _, record := range records {
		if record.Type == "weight" {
			if value, ok := numericValue(record.Data, "weight"); ok {
				latestWeight = value
				hasWeight = true
				break
			}
		}
	}
	if hasWeight {
		insights = append(insights, fmt.Sprintf("最近一次体重记录为 %.1fkg。", latestWeight))
	} else {
		actions = append(actions, "记录近 30 天体重，方便形成趋势判断。")
	}

	metricTotals := map[string]float64{}
	for _, point := range dataPoints {
		metricTotals[point.Metric] += point.Value
	}
	if water, ok := metricTotals["water_intake"]; ok && water > 0 && water < 400 {
		score -= 5
		insights = append(insights, fmt.Sprintf("最近采样周期饮水量偏低，累计约 %.0fml。", water))
		actions = append(actions, "关注饮水行为和排尿情况，必要时咨询兽医。")
	}
	if feed, ok := metricTotals["feeding_amount"]; ok && feed > 0 {
		insights = append(insights, fmt.Sprintf("自动喂食数据正常，最近累计投喂 %.0fg。", feed))
	}
	if motion, ok := metricTotals["motion"]; ok && motion < 50 {
		score -= 4
		actions = append(actions, "增加互动和活动量，观察精神状态变化。")
	}

	if activeAlerts == 0 {
		insights = append(insights, "当前没有未处理的健康预警。")
	}

	score = clamp(score, 35, 99)
	if len(actions) == 0 {
		actions = append(actions, "维持当前护理节奏，并按周期补录健康记录。")
	}

	return HealthSummary{
		Score:              score,
		Status:             scoreStatus(score),
		Insights:           dedupe(insights),
		RecommendedActions: dedupe(actions),
		DataPointsAnalyzed: len(dataPoints),
		GeneratedAt:        time.Now(),
	}
}

func (h *HealthAI) Answer(question string, summary HealthSummary) string {
	parts := []string{
		fmt.Sprintf("基于当前健康评分 %d/%s", summary.Score, summary.Status),
	}

	if len(summary.Insights) > 0 {
		parts = append(parts, "重点观察："+summary.Insights[0])
	}
	if len(summary.RecommendedActions) > 0 {
		parts = append(parts, "建议："+summary.RecommendedActions[0])
	}

	lowered := strings.ToLower(question)
	switch {
	case strings.Contains(question, "饮水") || strings.Contains(lowered, "water"):
		parts = append(parts, "如果饮水持续下降并伴随精神差、频繁排尿或排尿困难，应尽快线下就诊。")
	case strings.Contains(question, "体重") || strings.Contains(lowered, "weight"):
		parts = append(parts, "建议固定时间称重并连续记录，至少形成 2 到 4 周趋势后再判断是否异常。")
	case strings.Contains(question, "要不要去医院") || strings.Contains(question, "就医"):
		parts = append(parts, "若存在高严重级别预警、呕吐腹泻反复、出血或抽搐，应优先就医。")
	default:
		parts = append(parts, "这是一条规则生成的建议，不能替代兽医诊断。")
	}

	return strings.Join(parts, "；")
}

func numericValue(raw []byte, key string) (float64, bool) {
	payload := decodeMap(raw)
	value, ok := payload[key]
	if !ok {
		return 0, false
	}
	switch typed := value.(type) {
	case float64:
		return typed, true
	case int:
		return float64(typed), true
	default:
		return 0, false
	}
}

func decodeMap(raw []byte) map[string]any {
	result := map[string]any{}
	_ = jsonUnmarshal(raw, &result)
	return result
}

func jsonUnmarshal(raw []byte, out any) error {
	return json.Unmarshal(raw, out)
}

func scoreStatus(score int) string {
	switch {
	case score >= 85:
		return "excellent"
	case score >= 70:
		return "stable"
	case score >= 55:
		return "watch"
	default:
		return "critical"
	}
}

func clamp(value, minimum, maximum int) int {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func dedupe(values []string) []string {
	if len(values) == 0 {
		return values
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || slices.Contains(result, value) {
			continue
		}
		result = append(result, value)
	}
	return result
}
